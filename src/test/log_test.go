package test

// Time: 11:38
// Auther: hdk
import (
	"ohmydns/src/util"
	"testing"
)

func TestLogger(t *testing.T) {
	logger := new(util.Logger)
	logger.InitPath("../log/main.log")
	logger.Initlogger()
	logger.Info("test1")
	logger.Error("test2")
	logger.Warn("test3")
}
