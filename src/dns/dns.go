// dns 特殊处理
package dns

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	"ohmydns/src/util"
	"strings"
)

// 	v4-v6切换回合
const turn = 2

// CNAME 链长度
const len = 2 * turn

// dns应答
var dnsAnswer layers.DNSResourceRecord

var buf = gopacket.NewSerializeBuffer()
var opts = gopacket.SerializeOptions{} // See SerializeOptions for more details.

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

//****************************************DNS记录解析

// A记录请求处理函数
func HandleA(d DNSdata) {
	a, _, _ := net.ParseCIDR(d.rr.Record + "/24")
	dnsAnswer.Type = layers.DNSTypeA
	dnsAnswer.IP = a
	dnsAnswer.Name = []byte(d.Name)
	fmt.Println(d.Name)
	dnsAnswer.Class = layers.DNSClassIN
	// 返回消息填充
	d.rep.QR = true
	d.rep.ANCount = 1
	d.rep.OpCode = layers.DNSOpCodeNotify
	d.rep.AA = true
	d.rep.Answers = append(d.rep.Answers, dnsAnswer)
	d.rep.ResponseCode = layers.DNSResponseCodeNoErr

	err := d.rep.SerializeTo(buf, opts)
	if err != nil {
		panic(err)
	}
	d.u.WriteTo(buf.Bytes(), d.cAddr)

}

// AAAA记录请求处理函数
func HandleAAAA(d DNSdata) {
	a, _, _ := net.ParseCIDR(d.rr.Record + "/32")
	dnsAnswer.Type = layers.DNSTypeAAAA
	dnsAnswer.IP = a
	dnsAnswer.Name = []byte(d.Name)
	fmt.Println(d.Name)
	dnsAnswer.Class = layers.DNSClassIN
	// 返回消息填充
	d.rep.QR = true
	d.rep.ANCount = 1
	d.rep.OpCode = layers.DNSOpCodeNotify
	d.rep.AA = true
	d.rep.Answers = append(d.rep.Answers, dnsAnswer)
	d.rep.ResponseCode = layers.DNSResponseCodeNoErr

	err := d.rep.SerializeTo(buf, opts)
	if err != nil {
		panic(err)
	}
	d.u.WriteTo(buf.Bytes(), d.cAddr)
}

// NS记录请求处理函数
func HandleNS(sdata DNSdata) {

}

// CNAME记录请求处理函数
func HandleCN(d DNSdata) {
	dnsAnswer.Type = layers.DNSTypeCNAME
	dnsAnswer.CNAME = []byte(makedomain(d.cAddr.IP, d.Name))
	dnsAnswer.Name = []byte(d.Name)
	dnsAnswer.Class = layers.DNSClassIN

	// 返回消息填充
	d.rep.QR = true
	d.rep.ANCount = 1
	d.rep.OpCode = layers.DNSOpCodeNotify
	d.rep.AA = true
	d.rep.Answers = append(d.rep.Answers, dnsAnswer)
	d.rep.ResponseCode = layers.DNSResponseCodeNoErr

	err := d.rep.SerializeTo(buf, opts)
	if err != nil {
		panic(err)
	}
	d.u.WriteTo(buf.Bytes(), d.cAddr)
}

// 解析器跟踪
func TrackResvIP(s string) {

}

//
//// 根据参数定义对记录进行处理
//// TODO: argServeMUX实现
////func HandleRR(s string) string {
////	strs := strings.Split(s, " ")
////	handleRR:= map[string]func(){
////		'-i':makedomain()
////	}
////
////}
