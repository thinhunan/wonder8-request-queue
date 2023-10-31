package logic

import (
	"time"
)

const (
	MessageLeave = "LEAVE"
)

// Message 给用户发送的消息
type Message struct {
	Left     int       //总共还剩多少人排队
	Position int       //排在多少位
	Estimate int       //预估需要排多少秒
	MsgTime  time.Time //消息时间
	OverLoad bool
	Ip       string
}

func NewMessage(left, position, estimate int) *Message {
	message := &Message{
		Left:     left,
		Position: position,
		Estimate: estimate,
		MsgTime:  time.Now(),
	}
	return message
}
