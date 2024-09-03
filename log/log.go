package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	Logger = logrus.New()

	// 设置日志格式为 JSON
	Logger.SetFormatter(&logrus.JSONFormatter{})

	// 设置日志级别为 Info
	Logger.SetLevel(logrus.InfoLevel)

	// 添加一个输出目标，比如 os.Stdout
	file, err := os.OpenFile("log/logfile.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		if err = file.Close(); err != nil {
			panic(err)
		}
	}(file)

	Logger.SetOutput(file) // 将日志输出到文件

	Logger.Info("log initialized")
}
