package core

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"k8s-leader-election/pkg/server/model"
	"k8s.io/klog/v2"
	"log"
	"sync"
	"time"
)

type WsClientManager interface {
	// Register 注册客户端连接
	Register(conn *websocket.Conn, name string)
	// Send 发送命令
	Send(clientName string, input *model.WsRequest) (*model.WsResult, error)
	// Remove 删除客户端
	Remove(conn *WsClientToStore)
}

var WsManager WsClientManager

func init() {
	WsManager = &wsClientManager{
		Data: sync.Map{},
	}
}

// wsClientManager 维护所有客户端的连接实例
type wsClientManager struct {
	// 当有客户端conn要建立连接时，存入map
	Data sync.Map // key:客户端name value:扩展的websocket连接对象
}

// ResultMap 用于存储 "等待异步发送"的 chan
type ResultMap struct {
	ResultChanMap map[string]chan *model.WsResult
	Lock          sync.Mutex
}

var CallBackResult *ResultMap

func init() {
	CallBackResult = &ResultMap{
		ResultChanMap: map[string]chan *model.WsResult{},
		Lock:          sync.Mutex{},
	}
}

// Register 注册 ws 连接
// 存入conn信息到map中，代表ws实例已经被服务端管理起来，并启动3个goroutine
// 维护每个连接实例
func (c *wsClientManager) Register(conn *websocket.Conn, name string) {
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
func (c *wsClientManager) SendAll(input interface{}) error {
	var err error
	c.Data.Range(func(key, value any) bool {
		client := value.(*WsClientToStore)
		err = client.WsClient.Conn.WriteJSON(input) // 遍历客户端发送msg消息
		if err != nil {
			c.Remove(client)
			log.Println(err)
			return false
		}
		return true
	})

	if err != nil {
		return err
	}

	return nil
}

// Send 从维护的map中找到特定的局点，并发送消息
func (c *wsClientManager) Send(clientName string, input *model.WsRequest) (*model.WsResult, error) {
	// 1. 从map中查找需要的client端
	value, ok := c.Data.Load(clientName)
	if !ok {
		klog.Error("client not found: ", clientName)
		return nil, fmt.Errorf("not found client: %s", clientName)
	}
	client := value.(*WsClientToStore)
	a, _ := uuid.NewUUID()
	input.Uuid = a.String()

	// 存入 全局维护的 chan 中，用于存储异步返回的结果
	resultChan := make(chan *model.WsResult)

	CallBackResult.Lock.Lock()
	CallBackResult.ResultChanMap[a.String()] = resultChan
	CallBackResult.Lock.Unlock()

	// 结束后需要删除 chan 对象
	defer func() {
		CallBackResult.Lock.Lock()
		delete(CallBackResult.ResultChanMap, input.Uuid)
		CallBackResult.Lock.Unlock()
	}()

	// 2. 调用该客户端并发送
	err := client.WsClient.Conn.WriteJSON(input)
	if err != nil {
		klog.Error("execute send err: ", err)
		return nil, err
	}

	// 接口请求时会阻塞等待结果返回后才发送
	return <-resultChan, nil
}

// Remove 删除 map中的 conn 信息
func (c *wsClientManager) Remove(conn *WsClientToStore) {
	c.Data.Delete(conn.ClientName)
}
