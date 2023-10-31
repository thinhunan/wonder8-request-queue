package logic

import (
	"context"
	"errors"
	"io"
	"log"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	StatusInit    = 0
	StatusClosing = 1
	StatusClosed  = 2
)

type User struct {
	IP             string        `json:"ip"`
	Port           string        `json:"port"`
	EnterAt        time.Time     `json:"enter_at"`
	Status         int           `json:"status"`
	MessageChannel chan *Message `json:"-"`
	conn           *websocket.Conn
}

func NewUser(conn *websocket.Conn, ip, port string) *User {
	user := &User{
		IP:             ip,
		Port:           port,
		EnterAt:        time.Now(),
		MessageChannel: make(chan *Message, 32),
		conn:           conn,
	}
	return user
}

func (u *User) SendMessage(ctx context.Context) {
	for msg := range u.MessageChannel {
		if u.Status == StatusClosed {
			continue
		}
		err := wsjson.Write(ctx, u.conn, msg)
		if err != nil {
			log.Println(err)
		}
		if u.Status == StatusClosing {
			u.Status = StatusClosed
			time.Sleep(50 * time.Millisecond)
			_ = u.conn.Close(websocket.StatusNormalClosure, "")
		}
	}
}

func (u *User) ReceiveMessage(ctx context.Context) error {
	var (
		receiveMsg map[string]string
		err        error
	)
	for {
		err = wsjson.Read(ctx, u.conn, &receiveMsg)
		if err != nil {
			// 判定连接是否关闭了，正常关闭，不认为是错误
			var closeErr websocket.CloseError
			if errors.As(err, &closeErr) {
				return nil
			} else if errors.Is(err, io.EOF) {
				return nil
			}
			u.Status = StatusClosed
			return err
		}
		if msg, ok := receiveMsg["Event"]; ok {
			if msg == MessageLeave {
				u.Status = StatusClosed
				return nil
			}
		}
	}
}

// CloseMessageChannel 避免 goroutine 泄露
func (u *User) CloseMessageChannel() {
	close(u.MessageChannel)
}
