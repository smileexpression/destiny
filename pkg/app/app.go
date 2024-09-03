package app

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"smile.expression/destiny/log"
	"smile.expression/destiny/pkg/database"
)

type App struct {
	options *Options
}

type Options struct {
	DBOptions *database.Options `json:"dbOptions"`
}

func (a *App) Init() {
	workDir, _ := os.Getwd()
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(workDir + "/config")

	a.initOrUpdateConfig()

	viper.WatchConfig() // 监视配置文件的变化
	viper.OnConfigChange(func(e fsnotify.Event) {
		// 在配置文件发生更改时重新加载配置
		a.initOrUpdateConfig()
	})

	a.serve()
}

func (a *App) serve() {
	db := database.NewDB(a.options.DBOptions)
	fmt.Println(db)
}

func (a *App) initOrUpdateConfig() {
	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Logger.WithError(err).Error("read config failed")
	}

	if err := viper.Unmarshal(&a.options); err != nil {
		log.Logger.WithError(err).Error("unmarshal config file error")
	}
}
