package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

func main() {
	serverAddr := "ws://42.193.17.123:31180/ws/echo/"

	testWebsocket(serverAddr)
}

func testWebsocket(serverAddr string) {

	dialer := websocket.Dialer{}
	//向服务器发送连接请求，websocket

	reqHeader := map[string][]string{
		"clientname": []string{"ClientRegion"},
	}

	connect, _, err := dialer.Dial(serverAddr, reqHeader)
	if nil != err {
		log.Println(err)
		return
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
			log.Println(err)
			break
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
		msg := "from client to server"
		err := connect.WriteMessage(websocket.TextMessage, []byte(msg))
		if nil != err {
			log.Println(err)
			break
		}
		//休息一秒
		time.Sleep(time.Second)
	}
}
