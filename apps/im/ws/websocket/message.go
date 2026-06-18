package websocket

// 用户消息结构体，根据它来定义相对应的服务路由
type Message struct {
	Method string      `json:"method"`
	FormId string      `json:"form_id"`
	Data   interface{} `json:"data"`
}

func NewMessage(formId string, data interface{}) *Message {
	return &Message{
		FormId: formId,
		Data:   data,
	}
}
