package test

import (
	"fmt"
	"net"
	"ohmydns/src/util"
	"testing"
)

var ip4 = net.IP{10, 1, 1, 1}
var ip6 = net.IP{0xfc, 0, 10, 0, 0, 0, 01, 0, 10, 0, 0, 0, 0, 0, 0, 0}

func TestIPembed(t *testing.T) {
	//ip4 := net.IP{10, 1, 1, 1}
	newdomain := util.IPembed(ip4, "c1.testv4-v6.live")
	fmt.Println(newdomain)
	//ip6 := net.IP{0xfc, 0, 10, 0, 0, 0, 01, 0, 10, 0, 0, 0, 0, 0, 0, 0}
	newdomain6 := util.IPembed(ip6, "c1.testv4-v6.live")
	fmt.Println(newdomain6)
}

func TestGetNum(t *testing.T) {
	domain := "c4.--1.127-0-0-1.rip2003-8fac--5fa-1.v6.testv4-v6.live"
	str := util.GetNum(domain)
	fmt.Println(str)
}

func TestNumLog2Str(t *testing.T) {
	r := util.NewResolverLog()
	r.Add("127-0-0-1", "c1.v4.testv4-v6.live|A")
	r.Add("127-0-0-1", "c1.v6.testv4-v6.live|AAAA")

	str, err := r.NumLog2Str("127-0-0-1")
	if !err {
		fmt.Println(str)
	}
}

//func TestMakeprogress(t *testing.T) {
//	domain := "c4.testv4-v6.live"
//	fmt.Println(util.Makeprogress(domain))
//}

//func TestDomain46(t *testing.T) {
//	domain := "c1.fc00-a00-0-100-a00--.testv4-v6.live"
//	fmt.Println(Domain46(ip4, domain))
//}

//func TestChangesubdomain(t *testing.T) {
//	dom4 := "10-1-1-1.c1.testv4-v6.live"
//	//dom6 := "fc00-a00-0-100-a00--.c1.testv4-v6.live"
//	sub := "c2"
//	fmt.Println(util.(sub, dom4))
//}
