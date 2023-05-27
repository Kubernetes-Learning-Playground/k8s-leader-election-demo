package model

type SendRequestBody struct {
	// 选取的客户端名称
	ClientName string `json:"clientName"`
	// 发送的数据
	Data interface{} `json:"data"`
}
