package immodels

import (
	"IM/pkg/constants"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatLog struct {
	ID primitive.ObjectID `bson:"id,omitempty" json:"id,omitempty"`

	// 通过 SendId 和 RecvId 来获得具体 ConversationId
	ConversationId string          `bson:"conversationId"`
	SendId         string          `bson:"sendId"`
	RecvId         string          `bson:"recvId"`
	MsgFrom        int             `bson:"msgFrom"`
	MsgType        constants.MType `bson:"msgType"`
	// 信息类型
	MsgContent string `bson:"msgContent"`
	SendTime   int64  `bson:"sendTime"`
	Status     int    `bson:"status"`
	// 聊天类型：私聊 群聊
	ChatType constants.ChatType `bson:"chatType"`
	// TODO: Fill your own fields
	UpdateAt time.Time `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
