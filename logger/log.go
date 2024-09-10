package logger

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"smile.expression/destiny/pkg/constant"
)

var (
	SmileLog *SmileLogger
)

func Init() {
	SmileLog = NewLogger()
}

type SmileLogger struct {
	Logger *logrus.Logger
}

func (s *SmileLogger) WithContext(ctx context.Context) *logrus.Entry {
	requestID, ok := ctx.Value(constant.XRequestID).(string)
	if !ok {
		requestID = "unknown"
	}

	return s.Logger.WithFields(logrus.Fields{
		constant.XRequestID: requestID,
	})
}

func NewLogger() *SmileLogger {
	logger := logrus.New()

	// 设置日志格式为 JSON
	logger.SetFormatter(&logrus.JSONFormatter{})

	// 设置日志级别为 Info
	logger.SetLevel(logrus.InfoLevel)

	// 添加一个输出目标，比如 os.Stdout
	file, err := os.OpenFile("logger/logfile.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	logger.SetOutput(file) // 将日志输出到文件

	logger.Info("log initialized")

	// 监听终止信号，执行清理操作
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// 在收到终止信号时关闭文件
		logger.Info("closing log file")
		if err = file.Close(); err != nil {
			panic(err)
		}
		os.Exit(1) // 退出程序
	}()

	return &SmileLogger{
		Logger: logger,
	}
}
