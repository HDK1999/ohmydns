// dns 特殊处理
package main

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/miekg/dns"
	"net"
	"ohmydns/src/util"
	"strings"
)

const turn = 2

// CNAME 链长度
const len = 2 * turn

// 通过改变子域名表现实验进展位置，完成全部返回"stop"
func makeprogress(domain string) string {
	str := []byte(domain)
	if str[1] < 48+len {
		str[1] += 1
		return string(str)
	}
	return "stop"
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
	return Domain46(ip, domain)
}

//DNS解析响应,0代表成功解析，1代表解析存在问题
func serveDNS(u *net.UDPConn, clientAddr *net.UDPAddr, request *layers.DNS) int {
	replyMess := request
	var dnsAnswer layers.DNSResourceRecord
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{} // See SerializeOptions for more details.

	var rr RR
	var err error
	var ok bool

	// 循环判断question是否为key的子域名
	for k := range records {
		//泛域名特殊处理，且要保留原有键
		n := strings.ReplaceAll(k, "*.", "")
		if dns.IsSubDomain(n, string(request.Questions[0].Name)) {
			rr, ok = records[k]
			fmt.Println("成功解析")
		}
	}

	if !ok {
		//Todo: Log no data present for the IP and handle:todo
		//不存在对应记录
		err := request.SerializeTo(buf, opts)
		if err != nil {
			panic(err)
		}
		fmt.Println("不存在对应记录")
		u.WriteTo(buf.Bytes(), clientAddr)
		return 1
	}

	// A记录处理
	if rr.Type == "A" {
		a, _, _ := net.ParseCIDR(rr.Record + "/24")
		dnsAnswer.Type = layers.DNSTypeA
		dnsAnswer.IP = a
		dnsAnswer.Name = []byte(request.Questions[0].Name)
		fmt.Println(string(request.Questions[0].Name))
		dnsAnswer.Class = layers.DNSClassIN
	}
	// CNAME记录处理
	if rr.Type == "CNAME" {
		dnsAnswer.Type = layers.DNSTypeCNAME
		dnsAnswer.CNAME = []byte(makedomain(clientAddr.IP, string(request.Questions[0].Name)))
		dnsAnswer.Name = []byte(request.Questions[0].Name)
		dnsAnswer.Class = layers.DNSClassIN
	}

	// 返回消息填充
	replyMess.QR = true
	replyMess.ANCount = 1
	replyMess.OpCode = layers.DNSOpCodeNotify
	replyMess.AA = true
	replyMess.Answers = append(replyMess.Answers, dnsAnswer)
	replyMess.ResponseCode = layers.DNSResponseCodeNoErr

	err = replyMess.SerializeTo(buf, opts)
	if err != nil {
		panic(err)
	}
	u.WriteTo(buf.Bytes(), clientAddr)
	return 0
}

// 根据参数定义对记录进行处理
//TODO: argServeMUX实现
//func HandleRR(s string) string {
//	strs := strings.Split(s, " ")
//	handleRR:= map[string]func(){
//		'-i':makedomain()
//	}
//
//}
