package server

import (
	"net/http"

	"github.com/thinhunan/wonder8/request/queue/logic"
)

func RegisterHandle() {
	// 开始咨讯台服务消息
	go logic.Consultant.Start()

	http.HandleFunc("/", homeHandleFunc)
	http.HandleFunc("/websocket", WebSocketHandleFunc)
}
