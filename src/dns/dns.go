// dns 特殊处理
package dns

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	"strings"
)

// 	v4-v6切换回合
const turn = 2

// CNAME 链长度
const Len = 2 * turn

// dns应答
var dnsAnswer layers.DNSResourceRecord

var buf = gopacket.NewSerializeBuffer()
var opts = gopacket.SerializeOptions{} // See SerializeOptions for more details.

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
	//TODO：NS解析
}

// CNAME记录请求处理函数
func HandleCN(d DNSdata) {
	//根据实验需要，检测是否需要转化为AAAA记录返回
	str := []byte(d.Name)
	if str[1] >= 48+Len {
		d.Name = "last.*.testv4-v6.live"
		d.Type = "AAAA"
		d.rr = records[d.Name]
		// TODO:在这里记录关键信息
		HandleAAAA(d)
	}
	//正常解析过程
	dnsAnswer.Type = layers.DNSTypeCNAME
	s := strings.Split(d.rr.Record, " ")
	cname := s[0]
	//含有特殊选项
	if len(s) > 1 {
		// 默认从第二个开始为选项
		for _, v := range s {
			if v == "-i" {
				cname = HandleIPembed(d.Name, d.cAddr.IP)
				continue
			}
			if v == "-r" {
				cname = HandleReplacedomain(cname, d.rr.Record)
				continue
			}
		}
	}
	dnsAnswer.CNAME = []byte(cname)
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
