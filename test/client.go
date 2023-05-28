package main

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"k8s.io/klog"
	"time"
)

/*
	客户端连接的测试代码，服务端做了选主高可用，在客户端需要保证重试逻辑
*/

var (
	serverAddr = "ws://42.193.17.123:31180/ws/echo/"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 重试机制
	RetryTimeout(ctx, time.Second*2, testWebsocket)
}

func testWebsocket(ctx context.Context) error {

	dialer := websocket.Dialer{}

	// 使用header作为唯一标示
	reqHeader := map[string][]string{
		"Clientname": []string{"test-client"},
	}

	//向服务器发送连接请求，websocket
	connect, _, err := dialer.Dial(serverAddr, reqHeader)
	if nil != err {
		klog.Error(err)
		return err
	}
	// 关闭连接
	defer connect.Close()

	// 定时向客户端发送数据
	go tickWriter(connect)

	//启动数据读取循环，读取客户端发送来的数据
	for {
		//从 websocket 中读取数据
		//messageType 消息类型，websocket 标准
		//messageData 消息数据
		messageType, messageData, err := connect.ReadMessage()
		if nil != err {
			klog.Error(err)
			return err
		}
		switch messageType {
		case websocket.TextMessage: //文本数据
			fmt.Println(">> server response: " + string(messageData))
		case websocket.BinaryMessage: //二进制数据
			fmt.Println(messageData)
		case websocket.CloseMessage: //关闭
		case websocket.PingMessage: //Ping
		case websocket.PongMessage: //Pong
		default:
		}
	}
}

func tickWriter(connect *websocket.Conn) {
	for i := 0; i < 5; i++ {
		//向客户端发送类型为文本的数据
		msg := "from client to server, send test message"
		err := connect.WriteMessage(websocket.TextMessage, []byte(msg))
		if nil != err {
			klog.Error(err)
			break
		}
		//休息一秒
		time.Sleep(time.Second)
	}
}

// RetryTimeout 重试模式
func RetryTimeout(ctx context.Context, retryInterval time.Duration, execute func(ctx context.Context) error) {
	for {
		klog.Info("execute func\n")
		if err := execute(ctx); err == nil {
			klog.Info("work finished successfully\n")
			return
		}
		klog.Info("execute if timeout has expired\n")
		if ctx.Err() != nil {
			klog.Errorf("time expired 1 : %s\n", ctx.Err())
			return
		}
		klog.Infof("wait %s before trying again\n", retryInterval)
		// 创建一个计时器
		t := time.NewTimer(retryInterval)
		select {
		case <-ctx.Done():
			klog.Errorf("timed expired 2 :%s\n", ctx.Err())
			t.Stop()
			return
		// 定时执行！
		case <-t.C:
			klog.Info("retry again")
		}
	}
}
