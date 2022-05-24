package util

import (
	"fmt"
	"github.com/lestrrat/go-file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"strings"
	"time"
)

type Logger struct {
	logger  zap.Logger
	logPath string
}

var l *Logger

// Initlogger 初始化日志类工具，path代表主日志保存路径
func Initlogger(path string) {
	n := path
	var p string
	s := strings.Split(n, "/")
	for i := range s {
		if i != len(s)-1 {
			p = p + s[i] + "/"
			continue
		}
		p = p + "/"
		break
	}
	l = new(Logger)
	l.SetPath(p)
	if !Exists(l.logPath) {
		err := os.MkdirAll(l.logPath, 777)
		file, err := os.Create(path)
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {

			}
		}(file)
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println("mkdir logPath err!")
			panic("exit")
		}
	}
	encoder := initEncoder()

	// 实现多个日志等级接口,将不同的信息分隔开
	// Info:系统正常的工作日志
	// Debug:与实验相关的日志记录
	// Warn:系统工作异常日志（例如：无对应解析记录）
	// Error:错误日志，导致系统最终退出
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.InfoLevel
	})
	debugLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.DebugLevel
	})
	warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.WarnLevel
	})
	errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.ErrorLevel
	})

	// 获取 info、warn日志文件的io.Writer
	infoIoWriter := getWriter(path, "INFO")
	warnIoWriter := getWriter(path, "WARN")
	debugIoWriter := getWriter(path, "DEBUG")
	errorIoWriter := getWriter(path, "ERROR")

	// 创建Logger
	core := zapcore.NewTee(
		//zapcore.NewCore(encoder, zapcore.AddSync(infoIoWriter), infoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(warnIoWriter), warnLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(infoIoWriter), infoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(debugIoWriter), debugLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(errorIoWriter), errorLevel),
	)
	logger := zap.New(core, zap.AddCaller()) // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数
	l.logger = *logger
	l.logger.Info("日志记录工具启动")
}

func Info(a string) {
	defer l.logger.Sync()
	fmt.Println("Info:" + a)
	l.logger.Info(a)
}

func Error(a string) {
	defer l.logger.Sync()
	fmt.Println("Error:" + a)
	l.logger.Error(a)
}

func Warn(a string) {
	defer l.logger.Sync()
	fmt.Println("Warn:" + a)
	l.logger.Warn("Warn:" + a)
}

func Debug(a string) {
	defer l.logger.Sync()
	fmt.Println(a)
	l.logger.Debug(a)
}

//初始化Encoder
func initEncoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		TimeKey:     "time",
		CallerKey:   "file",
		EncodeLevel: zapcore.CapitalLevelEncoder, //基本zapcore.LowercaseLevelEncoder。将日志级别字符串转化为小写
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeCaller: zapcore.ShortCallerEncoder, //一般zapcore.ShortCallerEncoder，以包/文件:行号 格式化调用堆栈
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) { //一般zapcore.SecondsDurationEncoder,执行消耗的时间转化成浮点型的秒
			enc.AppendInt64(int64(d) / 1000000)
		},
	})
}

//查看文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

//日志文件切割
func getWriter(filename, lvl string) io.Writer {
	// 保存30天内的日志，每24小时(整点)分割一次日志
	hook, err := rotatelogs.New(
		filename[:len(filename)-4]+"--"+lvl+"-"+"-%Y%m%d%H.log",
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Hour*24*30),
		rotatelogs.WithRotationTime(time.Hour*24),
	)

	if err != nil {
		panic(err)
	}
	return hook
}

// 	配置日志位置
func (l *Logger) SetPath(s string) {
	l.logPath = s
}
