package main

// ohmydns入口文件

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	Dns "ohmydns/src/dns"
	"ohmydns/src/util"
	"strconv"
)

func main() {
	// 初始化日志工具
	util.Initlogger("./log/main.log")
	// 初始化实验记录缓冲区
	util.NewResolverLog()

	//Listen on UDP Port at ipv4&ipv6
	Serveaddr := net.UDPAddr{
		Port: 53,
		IP:   net.ParseIP("localhost"),
	}
	//ipv4和ipv6解析
	u, _ := net.ListenUDP("udp", &Serveaddr)
	util.Info("开启监听端口:" + strconv.Itoa(Serveaddr.Port))

	// 初始化DNS服务
	dnsserver := Dns.InitdnsServer()
	// 获取UDP层的信息
	for {
		tmp := make([]byte, 1024)
		_, addr, _ := u.ReadFromUDP(tmp)
		clientAddr := addr
		packet := gopacket.NewPacket(tmp, layers.LayerTypeDNS, gopacket.Default)
		dnsPacket := packet.Layer(layers.LayerTypeDNS)
		dns, _ := dnsPacket.(*layers.DNS)
		go dnsserver.ServeDNS(u, clientAddr, dns)
	}
}
