package dns

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/miekg/dns"
	"net"
	"ohmydns/src/util"
	"strings"
	"sync"
)

// Time: 14:14
// Auther: hdk

// 解析记录的结果结构体
type RR util.RR

// 进一步封装关键信息,方便对记录进行精细化处理
type DNSdata struct {
	// 请求域名
	Name string
	// 匹配的记录类型
	RType string
	// 请求的类型
	QType int
	// 请求响应
	rep *layers.DNS
	// 对应解析记录
	rr RR
	// 请求IP
	cAddr *net.UDPAddr
	// udp连接
	u *net.UDPConn
}

// 域名解析记录，特殊参数定义详见util/InitRRarg
var records = map[string]RR{
	"baidu.com":  {"223.34.34.34", "A"},
	"github.com": {"79.52.123.201", "A"},

	"*.v4.testv4-v6.live": {".v4.>>.v6. -i -r", "CNAME"},
	"*.v6.testv4-v6.live": {".v6.>>.v4. -i -r", "CNAME"},

	"lastdomain.testv4-v6.live": {"fe80::bcc0:e4ff:fe5f:9fa4", "AAAA"},
}

var Retran_flag bool

// DNS处理路由结构体(A、CNAME...)
type DNSServeMux struct {
	mu sync.RWMutex        //路由信息读写锁
	m  map[string]muxEntry //路由表
}

// 请求类型-处理函数对
type muxEntry struct {
	h Handler //对应处理函数
	t string  //对应DNS请求类型
}

// 处理函数
type Handler interface {
	ServeDNS(d DNSdata)
}

// 封装一个普通函数作为处理函数
type HandlerFunc func(DNSdata)

// 启动
func InitdnsServer() *DNSServeMux {
	mux := NewDNSServeMux()
	// 路由绑定
	mux.HandlerFunc("A", HandleA)
	mux.HandlerFunc("AAAA", HandleAAAA)
	mux.HandlerFunc("NS", HandleNS)
	mux.HandlerFunc("CNAME", HandleCN)
	return mux
}

func (h HandlerFunc) ServeDNS(d DNSdata) {
	h(d)
}

// 新建一个DNS处理路由
func NewDNSServeMux() *DNSServeMux {
	return new(DNSServeMux)
}

// 路由绑定
func (mux *DNSServeMux) Handle(t string, handler Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()
	if t == "" {
		panic("DNS: invalid DNStype")
	}
	if handler == nil {
		panic("DNS: nil handler")
	}
	if _, exist := mux.m[t]; exist {
		panic("DNS: multiple registrations for " + t)
	}

	if mux.m == nil {
		mux.m = make(map[string]muxEntry)
	}
	e := muxEntry{h: handler, t: t}
	mux.m[t] = e
}

// 接收一个具体的处理函数将其包装成 Handler,并调用Handle()进行绑定
func (mux *DNSServeMux) HandlerFunc(t string, handler func(d DNSdata)) {
	if handler == nil {
		//main.logger.Error("DNS:no handler")
		panic("no handler!!")
	}
	mux.Handle(t, HandlerFunc(handler))
}

// 根据所给的dns请求类型返回合适的handler
func (mux *DNSServeMux) Handler(dnsType string) (h Handler, t string) {
	return mux.handler(dnsType)
}

// Handler的主要实现
func (mux *DNSServeMux) handler(t string) (h Handler, dnstype string) {
	// 读写锁保持同步
	mux.mu.RLock()
	defer mux.mu.RUnlock()
	if t != "" {
		h, dnstype = mux.match(t)
	}
	if h == nil {
		//	没有匹配的handler
		panic("!!!no handler!!")
	}
	return
}

// 匹配合适的处理函数
func (mux *DNSServeMux) match(t string) (h Handler, dnstype string) {
	v, ok := mux.m[t]
	// 匹配成功
	if ok {
		return v.h, v.t
	}
	// 没有合适匹配
	return nil, ""
}

// udp解析之后到这里进行DNS数据初步解析(记录匹配，处理函数分发)
// 返回值：0代表正常解析，1代表因为某种原因导致解析失败（例如：无对应记录）
// TODO：有没有更好的无解析记录的处理方式
func (mux *DNSServeMux) ServeDNS(u *net.UDPConn, clientAddr *net.UDPAddr, request *layers.DNS) int {
	//replyMess := request
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{} // See SerializeOptions for more details.

	var rr RR
	ok := false

	// 循环判断question是否为key的子域名
	for k := range records {
		//泛域名特殊处理，且要保留原有键
		n := strings.ReplaceAll(k, "*.", "")
		if request.Questions != nil {
			if dns.IsSubDomain(n, string(request.Questions[0].Name)) {
				rr, ok = records[k]
				// TODO：最大前缀匹配
				break
			}
		} else {
			return 1
		}
	}
	// TODO:对于请求类型是否匹配的判断
	if !ok {
		//不存在对应记录
		err := request.SerializeTo(buf, opts)
		if err != nil {
			panic(err)
		}
		util.Warn("不存在对应" + string(request.Questions[0].Name) + "的解析记录")
		u.WriteTo(buf.Bytes(), clientAddr)
		return 1
	}
	// 将需要的关键数据集成到一个结构体中交由具体的函数处理
	dnsdata := DNSdata{
		// 默认认为只有一个查询域名
		Name:  string(request.Questions[0].Name),
		QType: int(request.Questions[0].Type),
		RType: rr.Type,
		rep:   request,
		rr:    rr,
		cAddr: clientAddr,
		u:     u,
	}
	h, _ := mux.Handler(dnsdata.RType)
	h.ServeDNS(dnsdata)
	return 0
}
