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
