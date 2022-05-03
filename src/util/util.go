package util

import (
	"flag"
	"net"
	"strings"
)

type RR struct {
	Record string
	Type   string
}

// RR特殊参数的状态
type RRarg struct {
	EmbIP        bool //	将请求的IP嵌入到域名中返回结果，仅适用于CNAME记录
	ReplaceCNAME bool //	父级域名替换，用于CNAME，例如请求的域名为b.a.com将返回b.c.live，可与-i选项叠加
}

//将ip嵌入域名中,作为下一级的子域名
func IPembed(ip net.IP, domain string) string {
	addr := net.IPAddr{IP: ip}
	if strings.Contains(addr.String(), ":") {
		//	ipv6地址
		return strings.ReplaceAll(addr.String(), ":", "-") + "." + domain
	}
	//ipv4地址
	return strings.ReplaceAll(addr.String(), ".", "-") + "." + domain
}

//判断一个域名是否是泛域名
func IsWildomain(s string) bool {
	//如果第一位为'*',则认为是泛域名
	if s[0] == '*' {
		return true
	}
	return false
}

// 定义在记录中客可以使用的参数
//TODO:完善
func InitRRarg() {
	_ = flag.Bool("i", false, "将请求的IP嵌入到域名中返回结果，仅适用于CNAME记录")
	_ = flag.Bool("r", false, "父级域名替换，用于CNAME，例如请求的域名为b.a.com将返回b.c.live，可与-i选项叠加")
	return
}
