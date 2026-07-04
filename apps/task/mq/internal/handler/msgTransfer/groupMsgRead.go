package msgTransfer

import (
	"IM/apps/im/ws/ws"
	"IM/pkg/constants"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type groupMsgRead struct {
	mu             sync.Mutex
	conversationId string
	// 主要用于记录消息
	push *ws.Push
	// 最终推送
	pushCh chan *ws.Push
	count  int
	// 上次推送时间
	pushTime time.Time
	done     chan struct{}
}

func newGroupMsgRead(push *ws.Push, pushChan chan *ws.Push) *groupMsgRead {
	return &groupMsgRead{
		push:           push,
		conversationId: push.ConversationId,
		pushCh:         pushChan,
		count:          0,
		pushTime:       time.Now(),
		done:           make(chan struct{}),
	}
}

// 合并消息
func (m *groupMsgRead) mergePush(push *ws.Push) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.count++
	for msgId, read := range push.ReadRecords {
		m.push.ReadRecords[msgId] = read
	}
}

func (m *groupMsgRead) IsIdle() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isIdle()
}

func (m *groupMsgRead) isIdle() bool {
	pushTime := m.pushTime
	// 由于阻塞原因*2放宽时间
	val := GroupMsgReadRecordDelayTime*2 - time.Since(pushTime)
	if val <= 0 && m.push == nil && m.count == 0 {
		return true
	}
	return false
}

func (m *groupMsgRead) clear() {
	select {
	case <-m.done:
	default:
		close(m.done)
	}
	m.push = nil
}

// 检查
func (m *groupMsgRead) transfer() {

	timer := time.NewTimer(GroupMsgReadRecordDelayTime / 2)
	defer timer.Stop()

	for {
		select {
		case <-m.done:
			return
		case <-timer.C:
			// 1. 超时发送
			m.mu.Lock()
			pushTime := m.pushTime
			val := GroupMsgReadRecordDelayTime - time.Since(pushTime)
			push := m.push

			if val > 0 && m.count < GroupMsgReadRecordDelayCount || push == nil {
				if val > 0 {
					timer.Reset(val)
				}
				// 未达标
				m.mu.Unlock()
				continue
			}
			m.pushTime = time.Now()
			timer.Reset(GroupMsgReadRecordDelayTime / 4)
			m.push = nil
			m.count = 0
			m.mu.Unlock()

			// 推送
			logx.Infof("over merge condition push %v", push)
			m.pushCh <- push
		default:
			// 2.超量发送
			m.mu.Lock()

			if m.count >= GroupMsgReadRecordDelayCount {
				push := m.push
				m.push = nil
				m.count = 0
				m.mu.Unlock()

				// 推送
				logx.Infof("over merge condition push %v", push)
				m.pushCh <- push
				continue
			}
			if m.isIdle() {
				m.mu.Unlock()
				// 使用msgReadTransfer释放
				m.pushCh <- &ws.Push{
					ChatType:       constants.GroupChatType,
					ConversationId: m.conversationId,
				}
				continue
			}

			m.mu.Unlock()

			timeDelay := GroupMsgReadRecordDelayTime / 4
			if timeDelay > time.Second {
				timeDelay = time.Second
			}
			time.Sleep(timeDelay)
		}

	}

}
