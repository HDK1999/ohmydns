package main

// ohmydns入口文件

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	"ohmydns/src/util"
)

type RR util.RR

var records map[string]RR
var logger util.Logger

func main() {
	// 初始化日志工具
	logger := util.Initlogger("/project/ohmydns/src/log/", "main.log")
	// 域名解析记录，特殊参数定义详见util/InitRRarg
	records = map[string]RR{
		"baidu.com":  {"223.34.34.34", "A"},
		"github.com": {"79.52.123.201", "A"},

		"*.v4.testv4-v6.live": {"v6.testv4-v6.live -i -r", "CNAME"},
		"*.v6.testv4-v6.live": {"v4.testv4-v6.live -i -r", "CNAME"},
		//"*.v4.testv4-v6.live": {"v6.testv4-v6.live -i -r", "CNAME"},
		//"*.v6.testv4-v6.live": {"v4.testv4-v6.live -i -r", "CNAME"},
	}

	//Listen on UDP Port at ipv4&ipv6
	Serveaddr := net.UDPAddr{
		Port: 53,
		IP:   net.ParseIP("localhost"),
	}
	//ipv4和ipv6解析
	u, _ := net.ListenUDP("udp", &Serveaddr)
	logger.Info("开启监听端口:" + string(Serveaddr.Port))

	// Wait to get request on that port
	for {
		tmp := make([]byte, 1024)
		_, addr, _ := u.ReadFromUDP(tmp)
		clientAddr := addr
		packet := gopacket.NewPacket(tmp, layers.LayerTypeDNS, gopacket.Default)
		dnsPacket := packet.Layer(layers.LayerTypeDNS)
		dns, _ := dnsPacket.(*layers.DNS)
		go serveDNS(u, clientAddr, dns)
	}
}
