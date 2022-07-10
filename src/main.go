package main

// ohmydns入口文件

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	Dns "ohmydns/src/dns"
	"ohmydns/src/util"
	"reflect"
	"strconv"
	"time"
)

// 定时任务
func exejob(t time.Ticker) {
	for {
		select {
		case <-t.C:
			fmt.Println("begin clean")
			fmt.Println(util.RLog.NumLog2Str("111"))
			fmt.Println(reflect.TypeOf(util.RLog))
			util.NewResolverLog()
		}
	}
}

func main() {
	// 初始化日志工具
	util.Initlogger("./log/main.log")
	// 初始化实验记录缓冲区
	util.NewResolverLog()
	// 初始化IP记录缓冲区，已废除
	//util.NewIPLog()
	// 初始化计时器，用于定时清空变量空间，减小系统负担
	t := time.NewTicker(time.Hour * 48)
	go exejob(*t)

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
