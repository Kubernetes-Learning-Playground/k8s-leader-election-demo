package core

import (
	"fmt"
	"github.com/gorilla/websocket"
	"k8s.io/klog/v2"
	"log"
	"sync"
	"time"
)

// ClientMap 需要保留多个客户端所用
var ClientMap *ClientMapStruct

func init() {
	ClientMap = &ClientMapStruct{}
}

type ClientMapStruct struct {
	// 当有客户端conn要建立连接时，存入map
	Data sync.Map // key:客户端name value:扩展的websocket连接对象

}

// Store 存入conn信息到map中，代表ws实例已经被服务端管理起来，并启动3个goroutine
// 维护每个连接实例
func (c *ClientMapStruct) Store(conn *websocket.Conn, name string) {
	wsClient := NewWsClient(conn)
	wsClientToStore := NewWsClientToStore(wsClient, name)
	klog.Info("save client name: ", wsClientToStore.ClientName)
	c.Data.Store(name, wsClientToStore) // 存入 key value需要记得

	// 存入后需要启动执行ping命令，测试是否conn还在
	go wsClientToStore.Ping(time.Second * 1)

	// 处理读data，写入对象中的chan中
	go wsClientToStore.ReadLoop()
	// 处理chan数据
	go wsClientToStore.HandlerLoop()

}

// SendAll 发送msg给所有客户端
func (c *ClientMapStruct) SendAll(msg string) {

	c.Data.Range(func(key, value any) bool {
		client := value.(*WsClientToStore)
		err := client.WsClient.Conn.WriteMessage(websocket.TextMessage, []byte(msg)) // 遍历客户端发送msg消息
		if err != nil {
			c.Remove(client)
			log.Println(err)
			return false
		}

		return true
	})
}

// Send 从维护的map中找到特定的局点，并发送消息
func Send(clientMap *ClientMapStruct, clientName string, input interface{}) error {
	// 1. 从map中查找需要的client端
	value, ok := clientMap.Data.Load(clientName)
	if !ok {
		klog.Error("client not found: ", clientName)
		return fmt.Errorf("not found client: %s", clientName)
	}
	client := value.(*WsClientToStore)

	// 2. 调用该客户端并发送
	err := client.WsClient.Conn.WriteJSON(input)
	if err != nil {
		klog.Error("execute send err: ", err)
		return err
	}

	return nil

}

// Remove 删除map中的conn信息
func (c *ClientMapStruct) Remove(conn *WsClientToStore) {
	c.Data.Delete(conn.ClientName)
}
