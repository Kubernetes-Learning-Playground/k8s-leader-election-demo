package core

import (
	"github.com/gorilla/websocket"
	"k8s.io/klog/v2"
	"time"
)

// WsClientToStore websocket连接对象
type WsClientToStore struct {
	WsClient   *WsClient
	ClientName string
}

func NewWsClientToStore(wsClient *WsClient, clientName string) *WsClientToStore {
	return &WsClientToStore{WsClient: wsClient, ClientName: clientName}
}

type WsClient struct {
	Conn      *websocket.Conn
	readChan  chan *WsMessage // 读队列(chan)
	closeChan chan struct{}
}

func NewWsClient(conn *websocket.Conn) *WsClient {
	return &WsClient{
		Conn:      conn,
		readChan:  make(chan *WsMessage),
		closeChan: make(chan struct{}),
	}
}

// Ping 服务端定时发送给map中的client
// 轮巡waitTime时间进行ping操作。
func (w *WsClientToStore) Ping(waittime time.Duration) {
	for {
		time.Sleep(waittime)
		err := w.WsClient.Conn.WriteMessage(websocket.PingMessage, []byte(""))
		if err != nil {
			ClientMap.Remove(w)
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
			ClientMap.Remove(w)
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
			klog.Infof(string(msg.MessageData))
			// 解析收到的请求
			// TODO: 可以给予回应
			//	w.WsClient.Conn.WriteMessage()

		case <-w.WsClient.closeChan:
			klog.Infof("chan closed!")
			return
		}

	}

}
