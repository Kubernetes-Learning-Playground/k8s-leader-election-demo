package core

import (
	"github.com/gorilla/websocket"
	"net/http"
)

// Upgrader 请求升级为websocket所用的变量
var Upgrader websocket.Upgrader

func init() {
	Upgrader = websocket.Upgrader{
		// 暂且让所有都通过，可以在这里校验
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
}
