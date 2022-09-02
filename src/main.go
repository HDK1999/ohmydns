package main

// ohmydns入口文件

import (
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/jinzhu/gorm"
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

//解析参数,并赋值给相应的flag
func parseparam() {
	// 重传选项
	retran := flag.Bool("rt", false, "延迟解析器响应诱发解析器进行重传")
	Dns.Retran_flag = *retran
	// 数据库选项
	addr := flag.String("ml", "127.0.0.1", "mysql服务的地址，用以记录日志")
	port := flag.Int("mp", 3306, "mysql服务的端口")
	pass := flag.String("mP", "1234", "mysql服务的密码")
	flag.Parse()
	util.Mconf.Addr = *addr
	util.Mconf.Port = *port
	util.Mconf.Pass = *pass
}

func main() {
	//初始化解析配置
	parseparam()
	db := util.Initmysql()
	defer func(db *gorm.DB) {
		err := db.Close()
		if err != nil {
			panic("数据库关闭失败")
		}
	}(db)
	// 初始化日志工具
	util.Initlogger("./log/main.log")
	// 初始化实验记录缓冲区
	util.NewResolverLog()
	// 初始化计时器，用于定时清空变量空间，减小系统负担
	t := time.NewTicker(time.Hour * 48)
	go exejob(*t)

	//Listen on UDP Port at ipv4&ipv6
	Serveaddr := net.UDPAddr{
		Port: 113,
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
		if dns.Questions[0].Type == layers.DNSTypeAAAA || dns.Questions[0].Type == layers.DNSTypeA {
			// 只记录A和AAAA记录请求
			util.Dnslog(db, addr, dns)
		}
		go dnsserver.ServeDNS(u, clientAddr, dns)
	}
}
