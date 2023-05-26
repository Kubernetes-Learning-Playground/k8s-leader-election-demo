package model

type SendRequestBody struct {
	ClientName string      `json:"clientName"`
	Data       interface{} `json:"data"`
}
