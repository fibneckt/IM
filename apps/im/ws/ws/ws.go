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
		ReceiverId         string `mapstructure:"receiverId"`
		SendTime           int64  `mapstructure:"sendTime"`
		Msg                `mapstructure:"msg"`
	}
)
