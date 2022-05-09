package test

// Time: 11:38
// Auther: hdk
import (
	"ohmydns/src/util"
	"testing"
)

func TestLogger(t *testing.T) {
	util.Initlogger("../log/main.log")
	//logger.InitPath("../log/main.log")
	//logger.Initlogger()
	util.Info("test")
	util.Warn("kaajajajaja")
}
