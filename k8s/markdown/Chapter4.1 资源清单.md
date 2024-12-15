# 一、资源类别

- 名称空间级资源

- - 工作负载型资源：Pod、ReplicaSet、Deployment...
  - 服务发现及负载均衡型资源：Service、 Ingress...
  - 配置与存储型资源：Volume、CSI...
  - 特殊类型的存储卷：ConfigMap、Secre

- 集群级资源

- - Namespace、Node、ClusterRole、ClusterRoleBinding

- 元数据型资源

- - HPA、PodTemplate、LimitRange

# 二、资源清单

```
apiVersion: v1
kind: Pod
metadata:
  name: pod-demo
  namespace: default
  labels:
    app: myapp
spec:
  containers:
  - name: myapp-1
    image: swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/wangyanglinux/myapp:v1
  - name: busybox-1
    image: swr.cn-north-4.myhuaweicloud.com/ddn-k8s/registry.k8s.io/e2e-test-images/busybox:1.29-4
    command:
      - "/bin/sh"
      - "-c"
      - "sleep 3600"
```

## 资源清单组成

### 1.1 group/apiVersion

```
kubectl api-versions
```

![img](https://cloud-pic.wpsgo.com/bFBMbXRBbDNzNGNUakNwRDd3dkNCRFB0eG9uYjY2QVVzYXd3N0dTZUlSSmdtMWlwUTU3bEZsUlIrblNHTFdycVpaeDcwMUhlaGVnTXpFNnpaY2tHUVMzZjk4OVlrRDNPc1JqUHhNcGR1eSs2Y3puYU9HbVp2aFlOQ3d5a2QzMi9rZktidWp4cllyUjk4NitJUVk2RVpGUDFibi9Xd0ZjRld3K0ZYRUdUSlltZys5bnI1anVpc21IRldmQlhEUThzU1dUTTJ1UnlLdDFxQUhVRzR5UDJWZ1JSSlJxY1B1Y3pnd0lzWHhGRm1NckkxTHV1bVE2UXl1b1lsRlJzZE4zbjVrWldQajdwNm1xdThVOTBzb1hzOEE9PQ==/attach/object/AIFKSDI5ADAHS?)

`apiVersion: v1` 是定义 API 版本的字段，它指定了 Kubernetes 使用的 API 版本。这是 Kubernetes 对象配置文件的必要字段之一，用于告诉 Kubernetes 管理器如何解析该资源的配置文件。Kubernetes 提供了不同版本的 API（例如 v1、apps/v1、batch/v1 等），每个版本支持不同的资源和特性。`v1` 是 Kubernetes 最基本的稳定版本，通常用于核心的 API 资源（如 Pods、Services、Namespaces 等）。

### 常见的 API 版本：

- `v1`：用于 Kubernetes 核心资源（如 Pod、Service、ConfigMap 等）。
- `apps/v1`：用于工作负载控制器（如 Deployment、ReplicaSet、StatefulSet）。
- `batch/v1`：用于批处理作业和 CronJob。
- `extensions/v1beta1`：一些资源的旧版本，Kubernetes 1.16 后逐渐废弃。

### 1.2 kind 类别

### 1.3 metadata 元数据

### 1.4 spec 期望

### 1.5 status状态

## 资源清单编写

可以先通过`kubectl explain`查询资源清单对象属性，比如：

```
# kubectl explain pod.spec.containers
KIND:       Pod
VERSION:    v1

FIELD: containers <[]Container>

DESCRIPTION:
    List of containers belonging to the pod. Containers cannot currently be
    added or removed. There must be at least one container in a Pod. Cannot be
    updated.
    A single application container that you want to run within a pod.
```

