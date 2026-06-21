package mq

import "IM/pkg/constants"

// 消息格式
type MsgChatTransfer struct {
	ConversationId     string `mapstructure:"conversationId"`
	constants.ChatType `mapstructure:"chatType"`
	SendId             string `mapstructure:"sendId"`
	RecvId             string `mapstructure:"recvId"`
	SendTime           int64  `mapstructure:"sendTime"`
	constants.MType    `json:"mType"`
	Content            string `json:"content"`
}
