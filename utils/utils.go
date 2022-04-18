package utils

import (
	"fmt"
	"net"
	"strings"
)

// ServerPort 远程Actor监听端口
const ServerPort = 8080

// GetOutBoundIP 获取本机IP地址
func GetOutBoundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		fmt.Println(err)
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	fmt.Println(localAddr.String())
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}