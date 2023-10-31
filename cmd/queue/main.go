package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "net/http/pprof"
	"github.com/thinhunan/wonder8/request/queue/global"
	"github.com/thinhunan/wonder8/request/queue/server"
)

var (
	addr   = ":80"
	banner = `
 ____ ___  ____  _   ____  _     _____ _     _____
queue system，start on：%s
`
)

func init() {
	global.Init()
}

func main() {
	addr = ":" + strconv.Itoa(global.Port)
	fmt.Printf(banner, addr)
	server.RegisterHandle()
	log.Fatal(http.ListenAndServe(addr, nil))
}
