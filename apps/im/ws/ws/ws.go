package ws

import (
	"IM/pkg/constants"
)

type (
	// 具体内容
	Msg struct {
		constants.MType `mapstructure:"mType"`
		Content         string `mapstructure:"content"`
	}

	// 对话
	Chat struct {
		ConversationId     string `mapstructure:"conversationId"`
		constants.ChatType `mapstructure:"chatType"`
		SenderId           string `mapstructure:"senderId"`
		ResvId             string `mapstructure:"resvId"`
		SendTime           int64  `mapstructure:"sendTime"`
		Msg                `mapstructure:"msg"`
	}

	Push struct {
		ConversationId     string `mapstructure:"conversationId"`
		constants.ChatType `mapstructure:"chatType"`
		SendId             string `mapstructure:"sendId"`
		RecvId             string `mapstructure:"recvId"`
		SendTime           int64  `mapstructure:"sendTime"`
		constants.MType    `json:"mType"`
		Content            string `json:"content"`

		RecvIds []string `mapstructure:"recvIds"` // 消息队列中对多用户的接收
	}
)
