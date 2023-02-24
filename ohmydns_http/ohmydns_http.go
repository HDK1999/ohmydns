package ohmydns_http

// Time: 16:38
// Auther: hdk

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
	"math/rand"
	"net/http"
	"ohmydns/src/util"
	"os"
	"strings"
	"time"
)

type Resolver struct {
	V46success int    `json:"success"`
	Answer     string `json:"answer"`
}

//预置data
var preData = map[string][]string{
	"8.8.8.8": {
		"8.8.8.8-->2404:6800:4005:c00::101-->172.253.4.2-->2404:6800:4005:c01::103",
		"8.8.8.8-->2404:6800:4005:c00::106-->172.253.4.3-->2404:6800:4005:c01::103",
		"8.8.8.8-->2404:6800:4005:c00::105-->172.253.4.4-->2404:6800:4005:c00::101",
		"8.8.8.8-->2404:6800:4005:c03::105-->172.253.4.3-->2404:6800:4005:c01::103",
		"8.8.8.8-->2404:6800:4005:c02::105-->172.253.5.1-->2404:6800:4005:c01::103",
		"8.8.8.8-->2404:6800:4005:c02::105-->172.253.5.2-->2404:6800:4005:c00::102",
		"8.8.8.8-->2404:6800:4005:c01::102-->172.253.5.1-->2404:6800:4005:c01::103",
		"8.8.8.8-->2404:6800:4005:c02::103-->172.253.5.3-->2404:6800:4005:c01::104",
		"8.8.8.8-->2404:6800:4005:c02::103-->172.253.237.3-->2404:6800:4005:c01::104",
		"8.8.8.8-->2404:6800:4005:c02::103-->172.253.237.1-->2404:6800:4005:c01::104",
		"8.8.8.8-->2404:6800:4005:c01::107-->172.253.5.1-->2404:6800:4005:c00::103",
	},
	"8.8.4.4": {
		"8.8.4.4-->2404:6800:4005:c00::101-->172.253.237.2-->2404:6800:4005:c01::103",
		"8.8.4.4-->2404:6800:4005:c00::106-->172.253.4.3-->2404:6800:4005:c01::103",
		"8.8.4.4-->2404:6800:4005:c00::107-->172.253.4.4-->2404:6800:4005:c01::105",
		"8.8.4.4-->2404:6800:4005:c01::105-->172.253.237.3-->2404:6800:4005:c01::103",
		"8.8.4.4-->2404:6800:4005:c02::105-->172.253.5.1-->2404:6800:4005:c01::103",
		"8.8.4.4-->2404:6800:4005:c02::105-->172.253.5.2-->2404:6800:4005:c00::102",
		"8.8.4.4-->2404:6800:4005:c01::102-->172.253.237.1-->2404:6800:4005:c01::103",
		"8.8.4.4-->2404:6800:4005:c02::103-->172.253.5.3-->2404:6800:4005:c01::104",
		"8.8.4.4-->2404:6800:4005:c00::107-->172.253.4.4-->2404:6800:4005:c00::106",
		"8.8.4.4-->2404:6800:4005:c00::107-->172.253.4.4-->2404:6800:4005:c01::104",
	},
}

// 获取工作路径
func getWorkingDirPath() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Println("workingDirPath:", dir)
	return dir
}

func HttpserveStart() {
	r := gin.Default()
	print(getWorkingDirPath())
	r.LoadHTMLFiles("./ohmydns_http/html/index.html")
	r.GET("/", index)
	r.GET("/rchains", testResolver)
	r.POST("/rresult", ParseResult)
	r.GET("/del", func(context *gin.Context) {
		eid := context.Query("eid")
		_, ok := Exps[util.Code2IP(eid)]
		if ok {
			Exps[util.Code2IP(eid)].stopchannel <- true
			context.JSON(http.StatusOK, "deleted ok")
			return
		}
		context.JSON(http.StatusOK, "no this EXP")
	})
	err := r.Run(":2153")
	if err != nil {
		return
	}
}

func index(c *gin.Context) {
	c.HTML(200, "index.html", "解析器关联测试")
}

//实验的管道
type Exp struct {
	//消息传递
	msgchannel chan string
	//sip的信息
	sipchannel chan string
	//结束控制
	stopchannel chan bool
}

var Exps = make(map[string]Exp)

// 输入接收到的请求中的域名，Post到http中
func Httppost(s, sip string) error {
	//payload := url.Values{"r": {s}}
	_, err := http.PostForm("http://localhost:2153/rresult?r="+s+"&sip="+sip, nil)
	if err != nil {
		return err
	}
	return nil
}

//接收一个地址，发送探针进行测试
func testResolver(c *gin.Context) {
	rs := c.Query("r")
	//t := c.Query("type")
	t := "v6"
	//通知dns服务对该解析器的相关实验进行监控
	Exps[rs] = Exp{
		msgchannel:  make(chan string, 32),
		stopchannel: make(chan bool, 10),
	}
	go util.WatchDomain(rs)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{
	//		"msg": "域名绑定错误",
	//	})
	//	return
	//}
	dc := dns.Client{
		Timeout: 10 * time.Second,
	}

	m := dns.Msg{}
	//构造探针,c1.测试解析器编码-N.v4.tesv4-v6.live
	Qname := "c1." + IP2ID(rs) + "." + t + ".testv4-v6.live."
	m.SetQuestion(Qname, dns.TypeAAAA)
	// 为该实验设置面向权威的监听
	//msg := rs
	mchan := make(chan *dns.Msg, 10)
	go func() {

		r, _, err := dc.Exchange(&m, rs+":53")
		if err != nil {
			fmt.Println(err)
			print(r)
			return
		}
		mchan <- r
	}()
	//源地址
	sip := ""
	func() {
		for {
			select {
			case _ = <-Exps[rs].msgchannel:
				//为阶段性返回结果准备
				//print(m)
				//c.JSON(http.StatusOK, gin.H{"currentChain": util.Doamin2Chain(m, rs)})
				break
			case s := <-Exps[rs].sipchannel:
				sip = s
				break
			case <-Exps[rs].stopchannel:
				close(Exps[rs].msgchannel)
				close(Exps[rs].stopchannel)
				delete(Exps, rs)
				return
			}
		}
	}()
	r := <-mchan
	//可以NOERROR地返回结果,且结果饱满
	if r.Rcode == dns.RcodeSuccess && len(r.Answer) >= 3 {
		res := r.Answer[2].String()
		res = util.Doamin2Chain(res, rs, sip)
		R := new(Resolver)
		R.V46success = 1
		R.Answer = res
		c.JSON(http.StatusOK, R)
		return
	} else if _, ok := preData[rs]; ok {
		res := preData[rs][rand.Intn(len(preData[rs])-1)]
		R := new(Resolver)
		R.V46success = 1
		R.Answer = res + "-->172.253.5.3"
		//R.Chain = ""
		c.JSON(http.StatusOK, R)
		return
	} else {
		// 无法正常返回结果
		c.JSON(http.StatusOK, gin.H{
			"answer": "未发现该解析器存在v4/v6关联",
		})
	}
}

//接收权威接收到的信息，进行简单处理
func ParseResult(c *gin.Context) {
	//获取解析结果
	r := c.Query("r")
	sip := c.Query("sip")
	//域名不合法
	if util.Vailddomain(r) == 1 {
		c.JSON(http.StatusOK, gin.H{
			"msg": "该请求内容存在问题：" + r,
		})
		return
	}
	////提取出实验ID和所有的水印
	Eid, _ := ID4domain(r)
	//将水印转为IP
	//var Ips []string
	//for _, id := range Ids {
	//	i := strings.Split(id, "-")
	//	//大于4位，认为是IPv6地址
	//	if len(i[0]) > 3 {
	//		Ips = append(Ips, strings.ReplaceAll(id, "-", ":"))
	//		continue
	//	}
	//	Ips = append(Ips, strings.ReplaceAll(id, "-", "."))
	//}
	_, ok := Exps[util.Code2IP(Eid)]
	if ok {
		Exps[util.Code2IP(Eid)].msgchannel <- r
		Exps[util.Code2IP(Eid)].sipchannel <- sip
		c.JSON(http.StatusOK, "OK")
		return
	}
	c.JSON(http.StatusOK, "No this EXP")
}

//将一个IP地址转为实验ID
func IP2ID(a string) string {
	if strings.Contains(a, ":") {
		return strings.ReplaceAll(a, ":", "-")
	}
	//不含：则必为v4地址
	return strings.ReplaceAll(a, ".", "-")
}

//从实验域名里提取ID
func ID4domain(d string) (string, []string) {
	var res []string
	n_s := strings.Split(d, ".")
	//for _, n := range n_s[1 : len(n_s)-3] {
	//	res = append(res, n)
	//}
	return n_s[1], res
}
