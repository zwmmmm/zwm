package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"

	pb "grpcDemo/proto/helloworld" // 引入编译生成的包

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

type greeter struct {
	pb.UnimplementedGreeterServer // 嵌入这个结构体以实现所有必须的方法
}

// 模板代码
// SayHello 实现了 helloworld.GreeterServer 接口
// 该接口定义在 helloworld_grpc.pb.go 中
//	func (UnimplementedGreeterServer) SayHello(context.Context, *HelloRequest) (*HelloReply, error) {
//		return nil, status.Errorf(codes.Unimplemented, "method SayHello not implemented")
//	}

func (*greeter) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Println("grpc server receive pb.HelloRequest: ", req)
	reply := &pb.HelloReply{Reply: fmt.Sprintf("Hello %s", req.Name)}
	return reply, nil
}

func (*greeter) SayHelloStream(stream pb.Greeter_SayHelloStreamServer) error {
	for {
		args, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		fmt.Println("grpc server receive pb.HelloRequest: ", args)
		reply := &pb.HelloReply{Reply: fmt.Sprintf("Hello %s", args.Name)}
		if err := stream.Send(reply); err != nil {
			return err
		}
	}
}

func Auth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("missing credentials")
	}
	var user string
	var password string

	if val, ok := md["user"]; ok {
		user = val[0]
	}
	if val, ok := md["password"]; ok {
		password = val[0]
	}

	if user != "admin" || password != "admin" {
		return grpc.Errorf(codes.Unauthenticated, "invalid token")
	}
	return nil
}

// server端只需要实现业务逻辑，不需要关心网络通信，这些都由 grpc 框架处理
// server端只需要将业务逻辑注册到 grpc server 中
// server端只需要监听端口，启动服务
// client端调用 server端的方法，只需要知道 server端的地址和端口
// server会将 client端的请求转发到对应的业务逻辑上，如SayHello、SayHelloStream

func main() {
	lis, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 一元拦截器
	var unaryAuthInterceptor grpc.UnaryServerInterceptor = func(
		ctx context.Context,
		req interface{}, // 一元client端会给 req 和 context
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		//拦截普通方法请求，验证 Token
		err = Auth(ctx)
		if err != nil {
			return
		}
		// 继续处理请求
		return handler(ctx, req)
	}

	// 流拦截器
	var streamAuthInterceptor grpc.StreamServerInterceptor = func(
		srv interface{},
		ss grpc.ServerStream, // 流式client端只给 context
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// 从流的上下文中提取 Token，进行验证
		err := Auth(ss.Context())
		if err != nil {
			return err // 验证失败，直接返回错误
		}
		// 验证通过，继续处理流请求
		return handler(srv, ss)
	}

	// 证书认证-双向认证
	// 从证书相关文件中读取和解析信息，得到证书公钥、密钥对
	cert, err := tls.LoadX509KeyPair("../cert/server.pem", "../cert/server.key")
	if err != nil {
		log.Fatalf("Failed to load client certificate: %v", err)
	}
	// 创建一个新的、空的 CertPool
	certPool := x509.NewCertPool()
	ca, _ := ioutil.ReadFile("../cert/ca.pem")
	if err != nil {
		log.Fatalf("Failed to read ca certificate: %v", err)
	}
	// 尝试解析所传入的 PEM 编码的证书。如果解析成功会将其加到 CertPool 中，便于后面的使用
	certPool.AppendCertsFromPEM(ca)
	// 构建基于 TLS 的 TransportCredentials 选项
	creds := credentials.NewTLS(&tls.Config{
		// 设置证书链，允许包含一个或多个
		Certificates: []tls.Certificate{cert},
		// 要求必须校验客户端的证书。可以根据实际情况选用以下参数
		ClientAuth: tls.RequireAndVerifyClientCert,
		// 设置根证书的集合，校验方式使用 ClientAuth 中设定的模式
		ClientCAs: certPool,
	})
	if err != nil {
		log.Fatalf("Failed to generate credentials %v", err)
	}
	server := grpc.NewServer(
		grpc.Creds(creds), // 使用 TLS 证书
		grpc.UnaryInterceptor( // 一元拦截器
			grpc_middleware.ChainUnaryServer(
				unaryAuthInterceptor, // Token 认证拦截器 authInterceptor
				/*
					grpc_validator.UnaryServerInterceptor() 是 go-proto-validators 提供的一个拦截器，
					它会自动读取 .proto 文件中定义的验证规则（如 validator.field 中的规则）并对请求数据进行验证。
				*/
				grpc_validator.UnaryServerInterceptor(),
			),
		),
		grpc.StreamInterceptor( // 流拦截器
			grpc_middleware.ChainStreamServer(
				streamAuthInterceptor, // Token 认证拦截器 streamAuthInterceptor
				// grpc_validator.StreamServerInterceptor() 会自动读取 .proto 文件中定义的验证规则，验证每个流式消息的数据有效性。
				grpc_validator.StreamServerInterceptor(),
			),
		),
	)
	// 注册 grpcurl 所需的 reflection 服务
	reflection.Register(server)
	// 注册业务服务
	pb.RegisterGreeterServer(server, &greeter{})

	fmt.Println("grpc server start ...")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
