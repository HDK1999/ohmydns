package util

import (
	"container/list"
	"flag"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/google/gopacket/layers"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// mysql配置
type Mysqlconfig struct {
	user    string
	Pass    string
	Addr    string
	Port    int
	dbname  string
	tabname string
}

var Mconf = Mysqlconfig{
	dbname:  "test46",
	tabname: "altas-resolver",
	user:    "root",
}

// reolver表
type Resolvbase struct { // 需要修改的字段映射
	Id    uint32    `gorm:"column:ID;PRIMARY_KEY"` // 记录ID
	Eid   string    `gorm:"column:EX_ID"`          // 对应的实验ID
	Date  time.Time `gorm:"column:DATE"`           // 记录输入时间
	Q     string    `gorm:"column:QVALUE"`         // 查询内容
	QType string    `gorm:"column:QTYPE"`          //查询类型
	N     string    `gorm:"column:EX_N"`           // 实验进展（c1，c2。。。）
	SIP   string    `gorm:"column:SIP"`            //查询的源IP
}

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
	//只要经过修改就输出
	//ChangeFlag map[string]bool
	Logmap map[string]list.List
	//Logmap     map[string]mapset.Set
}

// 请求源IP的记录结构体，已废弃
//type IPLog struct {
//	ChangeFlag map[string]bool
//	Logmap     map[string]list.List
//}

var RLog *ResolverLog

//var IpLog *IPLog

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
	reg1 := regexp.MustCompile(`\.([^\.]*)`)
	if reg1 == nil {
		Error("regexp err")
		panic("regexp error")
	}
	//根据规则提取关键信息
	result1 := reg1.FindStringSubmatch(domain)
	if len(result1) < 1 {
		return "noip"
	}
	return result1[len(result1)-4][1:]
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
	RLog.Logmap = make(map[string]list.List)
	//RLog.ChangeFlag = make(map[string]bool)
}

// 添加解析器请求记录，n——实验编号，l——对应日志
func (r ResolverLog) Add(n, l string) {
	// 搜寻是否存在对应记录
	s, ok := r.Logmap[n]
	// 不存在对应记录就新建
	if !ok {
		// 记录所有对应的实验编号的交互
		//set := mapset.NewSet() //日志唯一记录
		//set.Add(l)
		//r.Logmap[n] = set
		log := list.New()
		log.PushBack(l)
		r.Logmap[n] = *log
		//r.ChangeFlag[n] = true
		return
	}
	//存在对应记录就加入记录
	//判断新增日志是否是全新的，在以<ip,domain,qtype>存在新记录时输出的场景下使用
	//r.ChangeFlag[n] = r.IfChange(s, l)
	s.PushBack(l)
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
	// 存在记录，遍历集合重新格式化（当以set为记录底层时，已废除）
	//c := s.Iter()
	//str := "[" + fmt.Sprint(<-c)
	//for {
	//	if b, ok := <-c; ok {
	//		str = str + "," + fmt.Sprint(b)
	//	} else {
	//		str = str + "]"
	//		break
	//	}
	//}
	str := "["
	for i := s.Front(); i.Next() != nil; i = i.Next() {
		str = str + fmt.Sprint(i.Value) + ","
	}
	str = str + fmt.Sprint(s.Back().Value) + "]"
	return str, !ok
}

// 定义在记录中可以使用的参数
//TODO:完善
func InitRRarg() {
	_ = flag.Bool("i", false, "将请求的IP嵌入到域名中返回结果，仅适用于CNAME记录")
	_ = flag.Bool("r", false, "父级域名替换，用于CNAME，例如请求的域名为b.a.com将返回b.c.live，可与-i选项叠加")
	return
}

// 建立mysql连接
func Initmysql() *gorm.DB {
	// 不考虑数据库不存在的情况
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8&parseTime=True&loc=Local", Mconf.user, Mconf.Pass, Mconf.Addr, Mconf.Port, Mconf.dbname)
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		// 数据库连接失败
		panic(err)
	}
	if !db.HasTable(&Resolvbase{}) {
		err = db.Model(&Resolvbase{}).Debug().
			AutoMigrate(&Resolvbase{}).Error
		if err != nil {
			// 数据表创建发生错误
			panic(err)
		}
		if db.HasTable(&Resolvbase{}) {
			fmt.Println("表创建成功")
		} else {
			fmt.Println("表创建失败")
		}
	} else {
		fmt.Println("表已存在")
	}
	return db
}

// 记录dns信息
func Dnslog(db *gorm.DB, caddr *net.UDPAddr, data *layers.DNS) {
	// 是实验所需的请求
	if strings.Contains(string(data.Questions[0].Name), "testv4-v6") && data.Questions[0].Name[0] == 99 {
		r := Resolvbase{
			Eid:   GetNum(string(data.Questions[0].Name)),
			Date:  time.Now(),
			Q:     string(data.Questions[0].Name),
			QType: data.Questions[0].Type.String(),
			N:     string(data.Questions[0].Name)[1:2],
			SIP:   caddr.IP.String(),
		}
		db = db.Create(&r)
	}
}
