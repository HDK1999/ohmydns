package util

import (
	"net"
	"net/http"
	"time"
)

// Time: 16:35
// Auther: hdk

//dns监控
//通过键值对实现，一个域名对应一个计时器
var specialdomain = make(map[string]time.Ticker)

//错误处理
type dnsError struct {
	s string
}

func (d *dnsError) Error() string {
	return d.s
}

//根据输入的IP监控对应的实验域名
func WatchDomain(i string) error {
	// 将实验编号嵌入表中，存在于表中的实验会被监控
	t := time.NewTicker(10 * time.Second)
	d := IP2Code(net.ParseIP(i))
	specialdomain[d] = *t
	func() {
		for {
			select {
			//到达限定时间
			case <-specialdomain[d].C:
				err := StopWatchDomain(d)
				if err != nil {
					return
				}
				return
			}
		}
	}()
	return nil
}

//停止对域名的监控
func StopWatchDomain(c string) error {
	for i, _ := range specialdomain {
		if i == c {
			//删除对应实验的数据
			delete(specialdomain, i)
			_, err := http.Get("http://localhost:2153/del?eid=" + c)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return &dnsError{"未对该实验进行监控"}
}
