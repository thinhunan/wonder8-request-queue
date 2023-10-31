将超出服务集群处理能力的请求放入队列中，需要Load balance的配合

/cmd/queue/main.go --启动入口
/config/queue.yaml --配置文件
/global/config.go -- 配置对象
/global/init.go --读取配置，并且在配置更新时更新值
/logic/consultant.go --重点，处理消息
/logic/message.go --下发消息体结构
/logic/user.go --保存连上来的用户的conn, MessageChannel, ip是key
/server/handle.go --注册handle，启动consultant
/server/home.go --测试页
/server/realip.go --求客户端ip，优先级是x-forwarded-for,x-real-ip,remoteip，x-forwarded-for中取第一个非私有ip
/server/websocket.go -- websocket连接handle
