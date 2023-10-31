package server

import (
	"log"
	" github.com/thinhunan/wonder8/request/queue/logic"
	"net/http"
	"nhooyr.io/websocket"
)

func WebSocketHandleFunc(w http.ResponseWriter, req *http.Request) {
	// Accept 从客户端接受 WebSocket 握手，并将连接升级到 WebSocket。
	// 如果 Origin 域与主机不同，Accept 将拒绝握手，除非设置了 InsecureSkipVerify 选项（通过第三个参数 AcceptOptions 设置）。
	// 换句话说，默认情况下，它不允许跨源请求。如果发生错误，Accept 将始终写入适当的响应
	conn, err := websocket.Accept(w, req, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		log.Println("websocket accept error:", err)
		return
	}
	var clientIp string
	if q := req.URL.Query()["TestIp"]; len(q) > 0 {
		clientIp = q[0]
	}
	if len(clientIp) < 1 {
		clientIp = IP.FromRequest(req)
	}
	clientPort := IP.PortFromRequest(req)
	newUser := logic.NewUser(conn, clientIp, clientPort)
	go newUser.SendMessage(req.Context())
	logic.Consultant.UserEntering(newUser)
	err = newUser.ReceiveMessage(req.Context())
	newUser.Status = logic.StatusClosed

	logic.Consultant.UserLeaving(newUser)
	// 根据读取时的错误执行不同的 Close
	if err == nil {
		_ = conn.Close(websocket.StatusNormalClosure, "")
	} else {
		log.Println("read from client error:", err)
		_ = conn.Close(websocket.StatusInternalError, "Read from client error")
	}
}
