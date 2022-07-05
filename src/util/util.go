package util

import (
	"container/list"
	"flag"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// 实验所需记录请求日志的存储结构体
type ResolverLog struct {
	ChangeFlag map[string]bool
	Logmap     map[string]mapset.Set
}

// 请求源IP的记录结构体
type IPLog struct {
	ChangeFlag map[string]bool
	Logmap     map[string]list.List
}

var RLog *ResolverLog
var IpLog *IPLog

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

func NewResolverLog() {
	RLog = new(ResolverLog)
	RLog.Logmap = make(map[string]mapset.Set)
	RLog.ChangeFlag = make(map[string]bool)
}
func NewIPLog() {
	IpLog = new(IPLog)
	IpLog.Logmap = make(map[string]list.List)
	IpLog.ChangeFlag = make(map[string]bool)
}

// 添加解析器请求记录，n——实验编号，l——对应日志
func (r ResolverLog) Add(n, l string) {
	// 搜寻是否存在对应记录
	s, ok := r.Logmap[n]
	// 不存在对应记录就新建
	if !ok {
		// 记录所有对应的实验编号的交互
		set := mapset.NewSet()
		set.Add(l)
		r.Logmap[n] = set
		r.ChangeFlag[n] = true
		return
	}
	//存在对应记录就加入记录
	//判断新增日志是否是全新的
	r.ChangeFlag[n] = r.IfChange(s, l)
	s.Add(l)
	return
}

// 添加请求源IP的记录，n——实验编号，ip——源IP
func (i IPLog) Add(n, ip string) {
	// 搜寻是否存在对应记录
	s, ok := i.Logmap[n]
	// 不存在对应记录就新建
	if !ok {
		// 记录所有对应的实验编号的交互
		log := list.New()
		log.PushBack(ip)
		i.Logmap[n] = *log
		//i.ChangeFlag[n] = true
		return
	}
	//存在对应记录就加入记录
	//i.ChangeFlag[n] = true
	s.PushBack(ip)
	return
}

// 判断加入的日志是否会使原日志发生变化
func (r ResolverLog) IfChange(s mapset.Set, l string) bool {
	return !s.Contains(l)
}

// 将对应记录集合转为字符串，格式为[str1,str2...]
func (r ResolverLog) NumLog2Str(n string) (string, bool) {
	// 搜寻是否存在对应记录
	s, ok := r.Logmap[n]
	if !ok {
		Warn("不存在对应的实验编号")
		return "实验编号不存在", ok
	}
	// 存在记录，遍历集合重新格式化
	c := s.Iter()
	str := "[" + fmt.Sprint(<-c)
	for {
		if b, ok := <-c; ok {
			str = str + "," + fmt.Sprint(b)
		} else {
			str = str + "]"
			break
		}
	}
	return str, !ok
}

//将对应IP记录集合转为字符串，格式为[str1,str2...]
func (i IPLog) Log2Str(n string) (string, bool) {
	// 搜寻是否存在对应记录
	s, ok := i.Logmap[n]
	if !ok {
		Warn("不存在对应的实验编号")
		return "实验编号不存在", ok
	}
	// 存在记录，遍历集合重新格式化
	str := "["
	for i := s.Front(); i.Next() != nil; i = i.Next() {
		str = str + fmt.Sprint(i.Next().Value) + ","
	}

	str = str + fmt.Sprint(s.Back().Value) + "]"
	return str, !ok
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

// 定义在记录中可以使用的参数
//TODO:完善
func InitRRarg() {
	_ = flag.Bool("i", false, "将请求的IP嵌入到域名中返回结果，仅适用于CNAME记录")
	_ = flag.Bool("r", false, "父级域名替换，用于CNAME，例如请求的域名为b.a.com将返回b.c.live，可与-i选项叠加")
	return
}
