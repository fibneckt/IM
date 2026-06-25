package websocket

import "time"

// 用户消息结构体，根据它来定义相对应的服务路由
// msg, Id, seq, ack
type Message struct {
	FrameType `json:"frameType"`
	Id        string `json:"id"`
	AckSeq    int    `json:"ackSeq"`
	// 向目标方发送的时间
	ackTime time.Time `json:"ackTime"`
	// 发送失败次数 到限制数取消发送 ack
	errCount int         `json:"errCount"`
	Method   string      `json:"method"`
	FormId   string      `json:"form_id"`
	Data     interface{} `json:"data"` // json 序列化后 是一个 map[string]interface{} 数据结构体
}

type FrameType uint8

const (
	FramePing  FrameType = 0x0
	FrameData  FrameType = 0x1
	FrameAck   FrameType = 0x2
	FrameNoAck FrameType = 0x3

	FrameErr FrameType = 0x9
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
