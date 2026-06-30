package mq

import "IM/pkg/constants"

// 消息格式
type MsgChatTransfer struct {
	ConversationId     string `json:"conversationId"`
	constants.ChatType `json:"chatType"`
	SendId             string   `json:"sendId"`
	RecvId             string   `json:"recvId"`
	SendTime           int64    `json:"sendTime"`
	RecvIds            []string `json:"recvIds"`

	constants.MType `json:"mType"`
	Content         string `json:"content"`
}
