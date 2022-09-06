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
var opts = gopacket.SerializeOptions{FixLengths: true} // See SerializeOptions for more details.

//****************************************DNS记录解析

// A记录处理函数
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
	d.rep.OpCode = layers.DNSOpCodeQuery
	d.rep.AA = true
	d.rep.Answers = append(d.rep.Answers, dnsAnswer)
	d.rep.ResponseCode = layers.DNSResponseCodeNoErr

	err := d.rep.SerializeTo(buf, opts)
	if err != nil {
		panic(err)
	}
	if !Retran_flag {
		_, err := d.u.WriteTo(buf.Bytes(), d.cAddr)
		if err != nil {
			return
		}
	}
	err = buf.Clear()
	if err != nil {
		return
	}
}

// AAAA记录处理函数
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
	d.rep.OpCode = layers.DNSOpCodeQuery
	d.rep.AA = true
	d.rep.Answers = append(d.rep.Answers, dnsAnswer)
	d.rep.ResponseCode = layers.DNSResponseCodeNoErr

	err := d.rep.SerializeTo(buf, opts)
	if err != nil {
		panic(err)
	}
	// 控制是否返回最终结果
	if !Retran_flag {
		_, err := d.u.WriteTo(buf.Bytes(), d.cAddr)
		if err != nil {
			return
		}
	}
	err = buf.Clear()
	if err != nil {
		return
	}
}

// NS记录处理函数
func HandleNS(sdata DNSdata) {
	//TODO：NS解析
	dnsAnswer.Type = layers.DNSTypeNS
	dnsAnswer.NS = []byte(sdata.rr.Record)
	dnsAnswer.Class = layers.DNSClassIN
	dnsAnswer.Name = []byte(sdata.Name)
	dnsAnswer.TTL = 3600
	// 返回消息填充
	sdata.rep.QR = true
	sdata.rep.ANCount = 1
	sdata.rep.OpCode = layers.DNSOpCodeQuery
	sdata.rep.AA = true
	sdata.rep.Answers = append(sdata.rep.Answers, dnsAnswer)
	sdata.rep.ResponseCode = layers.DNSResponseCodeNoErr
	aInfo, ns := AuthInfo(sdata.Name)
	sdata.rep.Authorities = append(sdata.rep.Authorities, aInfo)
	sdata.rep.Additionals = append(sdata.rep.Additionals, AdditionalInfo(ns))

	err := sdata.rep.SerializeTo(buf, opts)
	if err != nil {
		panic(err)
	}
	if !Retran_flag {
		_, err2 := sdata.u.WriteTo(buf.Bytes(), sdata.cAddr)
		if err2 != nil {
			return
		}
	}
	err = buf.Clear()
	if err != nil {
		return
	}
}

// CNAME记录处理函数
func HandleCN(d DNSdata) {
	//根据实验需要，检测是否需要转化为AAAA记录返回
	str := []byte(d.Name)
	if str[0] == 99 {
		//	实验用域名
		if str[1] >= 48+Len {
			dName := "lastdomain.testv4-v6.live"
			d.RType = "AAAA"
			d.rr = records[dName]
			HandleAAAA(d)
		} else {
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
			d.rep.OpCode = layers.DNSOpCodeQuery
			d.rep.AA = true
			d.rep.Answers = append(d.rep.Answers, dnsAnswer)
			d.rep.ResponseCode = layers.DNSResponseCodeNoErr
			aInfo, ns := AuthInfo(d.Name)
			d.rep.Authorities = append(d.rep.Authorities, aInfo)
			d.rep.Additionals = append(d.rep.Additionals, AdditionalInfo(ns))

			err := d.rep.SerializeTo(buf, opts)
			if err != nil {
				panic(err)
			}
			if !Retran_flag {
				_, err2 := d.u.WriteTo(buf.Bytes(), d.cAddr)
				if err2 != nil {
					return
				}
			}
			err = buf.Clear()
			if err != nil {
				return
			}
		}
	} else {
		//	非实验用域名
		d.rep.QR = true
		err := d.rep.SerializeTo(buf, opts)
		_, err = d.u.WriteTo(buf.Bytes(), d.cAddr)
		if err != nil {
			return
		}
	}

}

// 返回权威服务器的授权信息
func AuthInfo(s string) (layers.DNSResourceRecord, string) {
	dnsserver := new(layers.DNSResourceRecord)
	var ns string
	dnsserver.Type = layers.DNSTypeNS
	// 根据不同的子域名返回对应信息，只需对v4.查询即可分类
	if strings.Contains(s, "v4.") {
		dnsserver.NS = []byte("ns4.testv4-v6.live")
		ns = "ns4.testv4-v6.live"
		dnsserver.Name = []byte("v4.testv4-v6.live")
	} else {
		dnsserver.NS = []byte("ns6.testv4-v6.live")
		ns = "ns6.testv4-v6.live"
		dnsserver.Name = []byte("v6.testv4-v6.live")
	}
	dnsserver.Class = layers.DNSClassIN
	dnsserver.TTL = 3600
	return *dnsserver, ns
}

func AdditionalInfo(s string) layers.DNSResourceRecord {
	dnsadd := new(layers.DNSResourceRecord)
	// 根据不同的NS返回额外信息
	if strings.Contains(s, "ns6") {
		dnsadd.Type = layers.DNSTypeAAAA
		a, _, _ := net.ParseCIDR("240c:4081:8002:8910::5" + "/64")
		dnsadd.IP = a
		dnsadd.Class = layers.DNSClassIN
	} else {
		dnsadd.Type = layers.DNSTypeA
		a, _, _ := net.ParseCIDR("120.48.148.235" + "/24")
		dnsadd.IP = a
		dnsadd.Class = layers.DNSClassIN
	}
	dnsadd.Name = []byte(s)
	dnsadd.TTL = 3600
	return *dnsadd
}

// 不要在服务端进行解析器跟踪，花销太大
//func TrackResvIP(s string) {
//
//}
