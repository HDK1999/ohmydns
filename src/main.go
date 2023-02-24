package main

// ohmydns入口文件

import (
	"flag"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	ohttp "ohmydns/ohmydns_http"
	Dns "ohmydns/src/dns"
	"ohmydns/src/util"
	"strconv"
)

// 定时任务,由于不使用log存储，因此取消
//func exejob(t time.Ticker) {
//	for {
//		select {
//		case <-t.C:
//			fmt.Println("begin clean")
//			fmt.Println(util.RLog.NumLog2Str("111"))
//			fmt.Println(reflect.TypeOf(util.RLog))
//			util.NewResolverLog()
//		}
//	}
//}
type Conf struct {
	Port int
	IP   string
}

//解析参数,并赋值给相应的flag
func parseparam() Conf {
	// 重传选项
	retran := flag.Bool("rt", false, "延迟解析器响应诱发解析器进行重传")
	Dns.Retran_flag = *retran
	// 数据库选项
	maddr := flag.String("ml", "124.221.228.62", "mysql服务的地址，用以记录日志")
	mport := flag.Int("mp", 3306, "mysql服务的端口")
	mpass := flag.String("mP", "hdk19990815", "mysql服务的密码")
	//服务选项
	port := flag.Int("p", 53, "DNS服务的端口")
	ip := flag.String("b", "localhost", "DNS服务的监听地址")
	flag.Parse()
	util.Mconf.Addr = *maddr
	util.Mconf.Port = *mport
	util.Mconf.Pass = *mpass
	return Conf{IP: *ip, Port: *port}
}

func main() {
	//初始化解析配置
	conf := parseparam()
	go ohttp.HttpserveStart()
	//db := util.Initmysql()
	//defer func(db *gorm.DB) {
	//	err := db.Close()
	//	if err != nil {
	//		panic("数据库关闭失败")
	//	}
	//}(db)
	// 初始化日志工具
	util.Initlogger("./log/main.log")
	// 初始化实验记录缓冲区
	util.NewResolverLog()
	// 初始化计时器，用于定时清空变量空间，减小系统负担，现无需该功能
	//t := time.NewTicker(time.Hour * 48)
	//go exejob(*t)

	//Listen on UDP Port at ipv4&ipv6
	Serveaddr := net.UDPAddr{
		Port: conf.Port,
		IP:   net.ParseIP(conf.IP),
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
		//域名符合要求
		if util.Vailddomain(string(dns.Questions[0].Name)) == 0 {
			//if dns.Questions[0].Type == layers.DNSTypeAAAA || dns.Questions[0].Type == layers.DNSTypeA {
			// 记录请求
			//util.Dnslog(db, addr, dns)
			//}
			// 处理请求
			go dnsserver.ServeDNS(u, clientAddr, dns)
		}
		//不符合要求的不做任何响应
	}
}
