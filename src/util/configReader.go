package util

// Time: 18:21
// Auther: hdk
// 读取config文件，并配置ohmydns
import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "gopkg.in/yaml.v2"
	"reflect"
	"strings"
)

// 配置文件结构体
type AllConfig struct {
	Test string       `yaml:"test"`
	G    GenerConfig  `yaml:"general"`
	C    ClientConfig `yaml:"client"`
	S    ServerConfig `yaml:"server"`
}

// 全局配置
type GenerConfig struct {
	WorkMode string `yaml:"mode"`
}

// Client模式配置
type ClientConfig struct {
	ScanMode     string `yaml:"scan_mode"`
	ProbeFormate string `yaml:"probe_formate"`
}

// Server模式配置
type ServerConfig struct {
	// 二级域名
	serveDomain string `yaml:"serve_domain"`
	// v4/v6 NS域名
	v6NS   string `yaml:"v6_ns"`
	v4NS   string `yaml:"v4_ns"`
	reTran bool
	port   int
	db     DBConfig
	laddr  string
}

// 数据库配置
type DBConfig struct {
	user string
	pwd  string
	// 数据库位置
	ip   string
	port int
}

//
// ConfigReader
//
// @Description: 读取配置文件加载配置，命令行仅支持定义配置文件位置
//
func ConfigReader() {
	var A AllConfig
	// 配置reader
	reader := viper.New()

	// 设置配置文件的类型
	reader.SetConfigType("yaml")
	// 查看是否存在额外配置的配置文件
	config := pflag.StringP("cfile", "f", "", "配置文件位置")
	// TODO:配置文件的帮助信息
	//pflag.String("chelp", "", "test\ntest")
	if *config == "" {
		//项目默认运行在bin目录
		// 默认情况
		reader.SetConfigName("ohmydns")
		// 添加配置文件的路径，指定 config 目录下寻找
		reader.AddConfigPath("../config")
	} else {
		// 对输入进行分片
		s := strings.Split(*config, "/")
		// 文件名
		confName := strings.Split(s[len(s)-1], ".")[0]
		print(confName)
		reader.SetConfigName(confName)
		// 文件路径
		path, _, ifpath := strings.Cut(*config, confName)
		if ifpath {
			print(path)
			reader.AddConfigPath(path)
		} else {
			panic(fmt.Errorf("配置文件位置输入有误"))
		}
	}
	// 寻找配置文件并读取
	//尝试进行配置读取
	if err := reader.ReadInConfig(); err != nil {
		panic(fmt.Errorf("配置文件读取失败: %v", err))
	}
	err := reader.Unmarshal(&A)
	if err != nil {
		panic(fmt.Errorf("unable to decode into struct, %v", err))
	}
	if IsConfigOK(A) {
		print("完成配置读取")
	}
}

//
// IsConfigOK
//  @Description: 判断配置是否完善
//  @param A 配置文件结构体
//
func IsConfigOK(A AllConfig) bool {
	var ok bool
	switch A.G.WorkMode {
	case "client":
		ok = IsCSconfigOK(A.C)
	case "server":
		ok = IsCSconfigOK(A.S)
	default:
		panic("配置文件存在问题")
	}
	return ok
}

//
// IsCSconfigOK
//  @Description: 判断客户端/服务端配置是否正确
//
func IsCSconfigOK[T ClientConfig | ServerConfig](config T) bool {
	val := reflect.ValueOf(config)
	for num := 0; num < val.NumField(); num++ {
		if val.Field(num).IsZero() {
			panic(fmt.Errorf("存在问题： %v", val.Field(num)))
		}
	}
	return true
}
