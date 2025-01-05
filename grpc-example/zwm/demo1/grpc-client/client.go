package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	pb "grpcDemo/proto/helloworld" // 引入编译生成的包
	"io"
	"io/ioutil"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type PerRPCCredentials interface {
	// GetRequestMetadata gets the current request metadata, refreshing
	// tokens if required. This should be called by the transport layer on
	// each request, and the data should be populated in headers or other
	// context. If a status code is returned, it will be used as the status
	// for the RPC. uri is the URI of the entry point for the request.
	// When supported by the underlying implementation, ctx can be used for
	// timeout and cancellation.
	// TODO(zhaoq): Define the set of the qualified keys instead of leaving
	// it as an arbitrary string.
	GetRequestMetadata(ctx context.Context, uri ...string) (
		map[string]string, error,
	)
	// RequireTransportSecurity indicates whether the credentials requires
	// transport security.
	RequireTransportSecurity() bool
}

type Authentication struct {
	User     string
	Password string
}

func (a *Authentication) GetRequestMetadata(context.Context, ...string) (
	map[string]string, error,
) {
	return map[string]string{"user": a.User, "password": a.Password}, nil
}

func (a *Authentication) RequireTransportSecurity() bool {
	return false
}

func main() {
	// 证书认证-双向认证
	// 从证书相关文件中读取和解析信息，得到证书公钥、密钥对
	cert, err := tls.LoadX509KeyPair("../cert/client.pem", "../cert/client.key")
	if err != nil {
		log.Fatalf("Failed to load client certificate: %v", err)
	}
	// 创建一个新的、空的 CertPool
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("../cert/ca.pem")
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
		ServerName: "www.example.grpcdev.cn",
		RootCAs:    certPool,
	})
	if err != nil {
		log.Fatal(err)
	}
	conn, err := grpc.NewClient("localhost:8000",
		// grpc.WithTransportCredentials(insecure.NewCredentials()), // 不使用TLS
		grpc.WithTransportCredentials(creds), // 使用TLS
		grpc.WithPerRPCCredentials(&Authentication{ // 自定义认证
			User:     "admin",
			Password: "admin",
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 非流式调用
	// client := pb.NewGreeterClient(conn)
	// reply, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "wqy"})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("grpc client receive pb.HelloRequest: ", reply.Reply)

	// 流式调用
	client := pb.NewGreeterClient(conn)

	// 流处理
	stream, err := client.SayHelloStream(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 启动一个协程，间隔1秒发送一次消息
	go func() {
		for {
			if err := stream.Send(&pb.HelloRequest{Name: "zzz"}); err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Second)
		}
	}()

	// 接收server端返回来的消息
	for {
		reply, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		fmt.Println(reply.Reply)
	}
}
