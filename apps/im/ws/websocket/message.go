package websocket

// 用户消息结构体，根据它来定义相对应的服务路由
type Message struct {
	FrameType `json:"frameType"`
	Method    string      `json:"method"`
	FormId    string      `json:"form_id"`
	Data      interface{} `json:"data"` // json 序列化后 是一个 map[string]interface{} 数据结构体
}

type FrameType uint8

const (
	FramePing FrameType = 0x0
	FrameData FrameType = 0x1
	FrameErr  FrameType = 0x9
)

func NewMessage(formId string, data interface{}) *Message {
	return &Message{
		FrameType: FrameData,
		FormId:    formId,
		Data:      data,
	}
}

func NewErrMessage(err error) *Message {
	return &Message{
		FrameType: FrameErr,
		Data:      err.Error(),
	}
}
