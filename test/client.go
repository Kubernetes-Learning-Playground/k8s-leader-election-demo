package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"k8s-leader-election/pkg/server/model"
	"k8s-leader-election/test/config"
	"k8s-leader-election/test/plugin"
	"k8s.io/klog"
	"log"
	"sync"
	"time"
)

/*
	客户端连接的测试代码，服务端做了选主高可用，在客户端需要保证重试逻辑
*/

var (
	serverAddr string
	clientName string
)

func main() {

	// 读取配置文件
	cfg, err := config.LoadConfig("./config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	serverAddr = fmt.Sprintf("ws://%s:%s/ws/echo/", cfg.ServerIp, cfg.ServerPort)
	clientName = cfg.ClientName

	// 启动插件
	for _, pluginName := range cfg.Plugins {
		component := plugin.ComponentPluginMap[pluginName]
		err := component.SetAvailable()
		if err != nil {
			log.Fatal(err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 重试机制
	RetryTimeout(ctx, time.Second*2, testWebsocket)
}

func testWebsocket(ctx context.Context) error {

	dialer := websocket.Dialer{}

	// 使用header作为唯一标示
	reqHeader := map[string][]string{
		"Clientname": []string{clientName},
	}

	//向服务器发送连接请求，websocket
	connect, _, err := dialer.Dial(serverAddr, reqHeader)
	if nil != err {
		klog.Error(err)
		return err
	}
	// 关闭连接
	defer connect.Close()

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

			var reqBody model.WsRequest

			err := json.Unmarshal(messageData, &reqBody)
			if err != nil {
				log.Fatal(err)
			}

			var result model.WsResult
			switch reqBody.Type {
			case model.Connected:
				result.ClientName = reqBody.ClientName
				result.Type = reqBody.Type
				result.Uuid = reqBody.Uuid
			default:
				result.ClientName = reqBody.ClientName
				result.Type = reqBody.Type
				result.Uuid = reqBody.Uuid
				result.StatusList = make([]model.Status, 0)

				var wg sync.WaitGroup
				var mu sync.Mutex
				for _, v := range reqBody.Operations {
					wg.Add(1)

					go func(op model.Operation) {
						defer wg.Done()

						var status string
						var res model.Status
						component, ok := plugin.ComponentPluginMap[op.Service]
						if !ok {
							status, _ = plugin.ComponentPluginMap["component_error"].Start(ctx, &op)
						} else {

							switch component.IsAvailable() {
							case true:
								status, err = component.Start(ctx, &op)
								if err != nil {
									status = "DoFail"
								}
							case false:
								status, _ = plugin.ComponentPluginMap["component_error"].Start(ctx, &op)
							}

						}

						res.Service = op.Service
						res.Status = status

						// 使用互斥锁保证线程安全地追加结果

						mu.Lock()
						result.StatusList = append(result.StatusList, res)
						mu.Unlock()
					}(v)
				}

				wg.Wait() // 阻塞主 goroutine，直到等待组计数器归零

			}
			err = connect.WriteJSON(result)
			if err != nil {
				log.Fatal(err)
			}
		case websocket.BinaryMessage: //二进制数据
			fmt.Println(messageData)
		case websocket.CloseMessage: //关闭
		case websocket.PingMessage: //Ping
		case websocket.PongMessage: //Pong
		default:
		}
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
