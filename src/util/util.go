package util

import (
	"flag"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type RR struct {
	Record string
	Type   string
}

// RR特殊参数的状态
type RRarg struct {
	EmbIP        bool //	将请求的IP嵌入到域名中返回结果，仅适用于CNAME记录
	ReplaceCNAME bool //	父级域名替换，用于CNAME，例如请求的域名为b.a.com将返回b.c.live，可与-i选项叠加
	NuminDomain  bool
}

var src = rand.NewSource(time.Now().UnixNano())

const (
	// 6 bits to represent a letter index
	letterIdBits = 6
	// All 1-bits as many as letterIdBits
	letterIdMask = 1<<letterIdBits - 1
	letterIdMax  = 63 / letterIdBits
)

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

// 从一段域名中获取到编号
func GetNum(domain string) string {
	//解析正则表达式，如果成功返回解释器
	reg1 := regexp.MustCompile(`\.rip(.*?)\.`)
	if reg1 == nil {
		Error("regexp err")
		panic("regexp error")
	}
	//根据规则提取关键信息
	result1 := reg1.FindAllStringSubmatch(domain, -1)
	return result1[0][1]
}

//判断一个域名是否是泛域名
func IsWildomain(s string) bool {
	//如果第一位为'*',则认为是泛域名
	if s[0] == '*' {
		return true
	}
	return false
}

// 获取实际可执行文件位置
func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	path = path[:index]
	return path
}

// 从l中随机生成长度为n的字符串
// 该功能已弃用
//func RandStr(n int, l string) string {
//	b := make([]byte, n)
//	// A rand.Int63() generates 63 random bits, enough for letterIdMax letters!
//	for i, cache, remain := n-1, src.Int63(), letterIdMax; i >= 0; {
//		if remain == 0 {
//			cache, remain = src.Int63(), letterIdMax
//		}
//		if idx := int(cache & letterIdMask); idx < len(l) {
//			b[i] = l[idx]
//			i--
//		}
//		cache >>= letterIdBits
//		remain--
//	}
//	return *(*string)(unsafe.Pointer(&b))
//}

// 定义在记录中客可以使用的参数
//TODO:完善
func InitRRarg() {
	_ = flag.Bool("i", false, "将请求的IP嵌入到域名中返回结果，仅适用于CNAME记录")
	_ = flag.Bool("r", false, "父级域名替换，用于CNAME，例如请求的域名为b.a.com将返回b.c.live，可与-i选项叠加")
	return
}
