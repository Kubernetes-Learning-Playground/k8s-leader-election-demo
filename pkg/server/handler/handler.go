package handler

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"k8s-leader-election/pkg/server/core"
	"k8s-leader-election/pkg/server/model"
	"k8s.io/klog"
	"log"
	"net/http"
	"os"
)

// Test 测试handler
func Test(w http.ResponseWriter, req *http.Request) {
	fmt.Println("test server")
	fmt.Println("pod name: ", os.Getenv("POD_NAME"))
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("pod name: %v\n", os.Getenv("POD_NAME"))))
}

// Echo 客户端启动时注册到 server 端的 handler
// client 主动发送连接给 server 端
func Echo(writer http.ResponseWriter, request *http.Request) {

	// 1. 从 header 取出 client 信息
	name := request.Header["Clientname"]

	// 2. 使用websocket升级协议
	client, err := core.Upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Fatal(err)
	}

	if name == nil {
		name = []string{"sample" + uuid.New().String()}
	}

	// 3. 发送已经完成连接通知
	connectionCheck := &model.WsRequest{
		ClientName: name[0],
		Type:       model.Connected,
		Uuid:       uuid.New().String(),
	}

	err = client.WriteJSON(connectionCheck) // 发送回客户端确认收到连接请求

	if err != nil {
		klog.Error("connect error: ", err)
		return
	}

	// 4. 升级后要放入server端中进行保存，记得需要区别每个client的身分
	core.WsManager.Register(client, name[0])

}

// Operation 执行远程操作
func Operation(writer http.ResponseWriter, request *http.Request) {
	s, _ := ioutil.ReadAll(request.Body) // body内容读入字符串s
	klog.Infof("\n%s\n", string(s))      // 在返回页面中显示内容
	var reqBody model.WsRequest
	err := json.Unmarshal(s, &reqBody)

	// 发送操作
	result, err := core.WsManager.Send(reqBody.ClientName, &reqBody)
	if err != nil {
		writer.Write([]byte("sending message error:" + err.Error()))
	} else {
		a, _ := json.Marshal(result)
		writer.Write([]byte(a))
	}
}
