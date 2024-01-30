package model

type SendRequestBody struct {
	// 选取的客户端名称
	ClientName string `json:"clientName"`
	// 发送的数据
	Data interface{} `json:"data"`
}

// WsRequest 操作请求
type WsRequest struct {
	// ClientName 客户端名，用于区分哪个客户端
	ClientName string `json:"clientName"`
	// Type 类型
	Type string `json:"type"`
	// Uuid 用于标示缓存 chan, 每个请求都有一个 uuid
	Uuid string `json:"uuid"`
	// Operations 具体操作列表
	Operations []Operation `json:"operations"`
}

// WsResult 操作结果
type WsResult struct {
	// ClientName 客户端名，用于区分哪个客户端
	ClientName string `json:"clientName"`
	// Type 类型
	Type string `json:"type"`
	// 用于标示缓存 chan, 每个请求都有一个 uuid
	Uuid string `json:"uuid"`
	// Operations 具体操作列表
	StatusList []Status `json:"statusList"`
}

// ws 连接制定的动作
const (
	Connected   = "connected"
	Closed      = "closed"
	ReadAction  = "read"
	WriteAction = "write"
)

// Operation 操作, 由组件服务抽象而成
type Operation struct {
	// Service 组件名
	Service string `json:"service"`
	// Method 方法
	Method string `json:"method"`
	// 请求资源
	Url string `json:"url"`
	// Body 请求参数
	Body string `json:"body"`
}

// Status 操作后返回的结果
type Status struct {
	// Service 组件名
	Service string `json:"service"`
	// Status 执行状态
	Status string `json:"status"`
}
