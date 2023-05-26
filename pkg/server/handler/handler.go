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

func Test(w http.ResponseWriter, req *http.Request) {
	fmt.Println("test server")
	fmt.Println("pod name: ", os.Getenv("POD_NAME"))
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("pod name: %v\n", os.Getenv("POD_NAME"))))
}

// Echo 客户端启动时注册到server端的handler
func Echo(writer http.ResponseWriter, request *http.Request) {

	// 1. 从header取出client信息
	name := request.Header["Clientname"]

	// 2. 使用websocket升级协议
	client, err := core.Upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Fatal(err)
	}

	// 3. 发送已经完成连接通知
	if name == nil {
		name = []string{"sample" + uuid.New().String()}
	}
	data := name[0] + "is connected..."
	err = client.WriteJSON(data) // 发送回客户端同意
	klog.Infof(data)
	if err != nil {
		klog.Error("connect error: ", err)
		return
	}
	// 4. 升级后要放入server端中进行保存，记得需要区别每个client的身分
	core.ClientMap.Store(client, name[0])

}

func Send(writer http.ResponseWriter, request *http.Request) {
	s, _ := ioutil.ReadAll(request.Body) // 把body内容读入字符串s
	klog.Infof("\n%s\n", string(s))      // 在返回页面中显示内容
	var reqBody model.SendRequestBody
	err := json.Unmarshal(s, &reqBody)
	klog.Infof("%s, %s", reqBody.ClientName, reqBody.Data)
	err = core.Send(core.ClientMap, reqBody.ClientName, reqBody.Data)
	if err != nil {
		writer.Write([]byte("sending message error:" + err.Error()))
	} else {
		writer.Write([]byte("sending message successfully OK"))
	}
}
