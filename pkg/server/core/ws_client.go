package core

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"k8s-leader-election/pkg/server/model"
	"k8s.io/klog/v2"
	"log"
	"time"
)

// WsClientToStore websocket 连接对象
type WsClientToStore struct {
	WsClient   *WsClient
	ClientName string
}

func NewWsClientToStore(wsClient *WsClient, clientName string) *WsClientToStore {
	return &WsClientToStore{WsClient: wsClient, ClientName: clientName}
}

type WsClient struct {
	Conn *websocket.Conn
	// 读队列(chan)
	readChan  chan *WsMessage
	closeChan chan struct{}
}

func NewWsClient(conn *websocket.Conn) *WsClient {
	return &WsClient{
		Conn:      conn,
		readChan:  make(chan *WsMessage),
		closeChan: make(chan struct{}),
	}
}

// Ping 服务端定时发送给 map 中的 client
// 轮巡 waitTime 时间进行 ping 操作。
func (w *WsClientToStore) Ping(waittime time.Duration) {
	for {
		time.Sleep(waittime)
		err := w.WsClient.Conn.WriteMessage(websocket.PingMessage, []byte(""))
		if err != nil {
			WsManager.Remove(w)
			return
		}
	}
}

func (w *WsClientToStore) ReadLoop() {
	for {
		t, data, err := w.WsClient.Conn.ReadMessage()
		if err != nil {
			w.WsClient.Conn.Close()
			klog.Errorf("client exits and is removed from the map: ", err)
			WsManager.Remove(w)
			// 出错，通知close
			w.WsClient.closeChan <- struct{}{}
			break
		}
		// 如果读取正确，放入chan中
		msg := NewWsMessage(t, data)
		w.WsClient.readChan <- msg
	}
}

func (w *WsClientToStore) HandlerLoop() {

	for {
		// 处理读请求的chan
		select {
		case msg := <-w.WsClient.readChan:
			// 收到请求后返回给前端展示
			klog.Infof(string(msg.MessageData))

			var result *model.WsResult
			err := json.Unmarshal(msg.MessageData, &result)
			if err != nil {
				log.Fatal(err)
			}

			if result.Type != model.Connected {
				CallBackResult.Lock.Lock()
				v, ok := CallBackResult.ResultChanMap[result.Uuid]
				CallBackResult.Lock.Unlock()
				if !ok {
					continue
				}
				v <- result
			}

		case <-w.WsClient.closeChan:
			klog.Infof("chan closed!")
			return
		}

	}
}
