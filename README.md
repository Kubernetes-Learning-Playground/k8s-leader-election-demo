### 基于k8s提供的分布式锁实现集群内的pod选主机制
![](https://github.com/googs1025/k8s-leader-election-demo/blob/main/image/%E6%B5%81%E7%A8%8B%E5%9B%BE%20(1).jpg?raw=true)
#### 项目思路：
使用k8s提供的分布式选主工具，实现pod间有状态应用。

本项目使用websocket server 连接的有状态应用实现。

### 部署步骤：
前置步骤：先确保启动项目机器上有k8s集群。
1. 进入项目根目录，打包docker镜像
```bash 
docker build -t leaderelection:v1 .
```

2. 进入deploy目录
```bash
[root@VM-0-16-centos k8s-leader-election]# cd deploy/
[root@VM-0-16-centos deploy]# kubectl apply -f .
deployment.apps/leaderelection configured
service/leaderelection-svc unchanged
serviceaccount/leaderelection-sa unchanged
role.rbac.authorization.k8s.io/leaderelection-role unchanged
rolebinding.rbac.authorization.k8s.io/leaderelection-rolebinding unchanged
```

3. 查看服务状态
```bash 
[root@VM-0-16-centos deploy]# kubectl get pods | grep leaderelection
leaderelection-857c88477d-2lrhd             1/1     Running             0               31h
leaderelection-857c88477d-dvfwd             1/1     Running             0               31h
leaderelection-857c88477d-nmd58             1/1     Running             0               31h
```

4. 查看各副本状态：
使用logs命令查看pod日志，可以发现三个副本仅有其中一个对外提供接口，其馀都是待命状态，当发生选举切换时，会进行相应的回调
```bash
[root@VM-0-16-centos deploy]# kubectl logs -f leaderelection-857c88477d-2lrhd
I0527 00:31:12.438913       1 leaderelection.go:248] attempting to acquire leader lease default/example-lease...
I0527 00:31:12.447462       1 main.go:104] new leader is leaderelection-857c88477d-nmd58
```

```bash
[root@VM-0-16-centos deploy]# kubectl logs -f leaderelection-857c88477d-dvfwd
I0527 00:31:12.466720       1 leaderelection.go:248] attempting to acquire leader lease default/example-lease...
I0527 00:31:12.476857       1 main.go:104] new leader is leaderelection-857c88477d-nmd58
```

```bash
[root@VM-0-16-centos deploy]# kubectl logs -f leaderelection-857c88477d-nmd58
I0527 00:31:12.405744       1 leaderelection.go:248] attempting to acquire leader lease default/example-lease...
I0527 00:31:12.429921       1 leaderelection.go:258] successfully acquired lease default/example-lease
I0527 00:31:12.430135       1 main.go:101] still the leader!
I0527 00:31:12.430149       1 main.go:88] leader election server running...
I0527 00:31:12.430171       1 server.go:38] :8080
I0527 00:31:12.431351       1 server.go:29] 0.0.0.0:9998
I0527 00:38:34.717758       1 handler.go:42] sample226d81e3-57e2-404a-9fbc-51f23a5bec69is connected...
I0527 00:38:34.717777       1 client_map.go:29] save client对象: sample226d81e3-57e2-404a-9fbc-51f23a5bec69
I0527 00:38:35.133023       1 ws_client.go:71] from client to server, send test message
I0527 00:38:36.134495       1 ws_client.go:71] from client to server, send test message
I0527 00:38:37.144660       1 ws_client.go:71] from client to server, send test message
I0527 00:38:38.142288       1 ws_client.go:71] from client to server, send test message
I0527 00:38:39.460931       1 ws_client.go:71] from client to server, send test message
```

5. 调用接口调适

本项目server对外暴露接口，可使用websocket接口与http接口调适。
   
ws服务接口调用可使用websocket在线测试或test目录中的测试代码。   
```bash
测试接口: 返回当前pod名，GET, http://xxx.xxx.xxx.xxx:31180/test
ws连接接口: 建立客户端与服务端的websocket连接，ws://xxx.xxx.xxx.xxx:31180/ws/echo/
send接口: 选择特定客户端，并发对远端客户端发送数据，POST，http://xxx.xxx.xxx.xxx:31180/send
```
![](https://github.com/googs1025/k8s-leader-election-demo/blob/main/image/img.png?raw=true)

![](https://github.com/googs1025/k8s-leader-election-demo/blob/main/image/img_1.png?raw=true)
