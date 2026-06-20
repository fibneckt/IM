package websocket

// 用户消息结构体，根据它来定义相对应的服务路由
type Message struct {
	FrameType `json:"frameType"`
	Method    string      `json:"method"`
	FormId    string      `json:"form_id"`
	Data      interface{} `json:"data"`
}

type FrameType uint8

const (
	FrameData    FrameType = 0x0
	FrameHeaders FrameType = 0x1
)

func NewMessage(formId string, data interface{}) *Message {
	return &Message{
		FrameType: FrameData,
		FormId:    formId,
		Data:      data,
	}
}
