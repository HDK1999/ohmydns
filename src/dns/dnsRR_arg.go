package dns

import (
	"net"
	"ohmydns/src/util"
	"strings"
)

// 通过改变子域名表现实验进展位置
// 处理的域名均不是最后一步的域名
// 因为会在对域名进行预解析时拦截全部无需处理的实验用请求
func makeprogress(domain string) string {
	str := []byte(domain)
	str[1] += 1
	return string(str)
}

// 构建v4-v6cname所需的域名，最低一级的子域名代表了实验进展的位置
func Domain46(ip net.IP, domain string) string {
	if makeprogress(domain) != "stop" {
		domain = makeprogress(domain)
	}
	sub := strings.Split(domain, ".") // 每一级的域名
	return sub[0] + "." + util.IPembed(ip, domain[3:])
}

// 根据域名情况动态生成新的域名
func makedomain(ip net.IP, domain string) string {
	// 用于v4-v6关联用
	return Domain46(ip, domain)
}

// 处理i参数,将IP嵌入到请求的domain中
func HandleIPembed(domain string, IP net.IP) string {
	return makedomain(IP, domain)
}

// 处理r参数,将请求domain中部分字符根据解析记录中的rule进行替换，规则为olds>>news
func HandleReplacedomain(domain, r string) string {
	rule := strings.Split(strings.Split(r, " ")[0], ">>")
	ndomain := strings.ReplaceAll(domain, rule[0], rule[1])
	return ndomain
}
