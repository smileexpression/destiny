package app

import (
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/gorm"

	"smile.expression/destiny/logger"
	"smile.expression/destiny/pkg/database"
	"smile.expression/destiny/pkg/http/controller"
	"smile.expression/destiny/pkg/http/middleware"
	"smile.expression/destiny/pkg/http/routes"
	"smile.expression/destiny/pkg/storage"
)

type App struct {
	options           *Options
	r                 *gin.Engine
	db                *gorm.DB
	storageClient     *storage.Client
	userController    *controller.UserController
	storageController *controller.StorageController
	bannerController  *controller.BannerController
}

type Options struct {
	DBOptions      *database.Options `json:"dbOptions"`
	StorageOptions *storage.Options  `json:"storageOptions"`
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
	a.db = database.NewDB(a.options.DBOptions)

	a.storageClient = storage.NewClient(a.options.StorageOptions)

	// controller
	a.r = gin.Default()
	a.r.Use(middleware.CORSMiddleware(), middleware.RecoveryMiddleware())
	a.r.Use(middleware.GenerateRequestID(), middleware.SetRequestID())

	a.userController = controller.NewUserController(a.r, a.db)
	a.userController.Register()
	a.storageController = controller.NewStorageController(a.r, a.db, a.storageClient)
	a.storageController.Register()
	a.bannerController = controller.NewBannerController(a.r, a.db)
	a.bannerController.Register()

	a.r = routes.CollectRoute(a.r)
	panic(a.r.Run(":" + viper.GetString("server.port")))
}

func (a *App) initOrUpdateConfig() {
	var (
		log = logger.SmileLog.Logger
	)
	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.WithError(err).Error("read config failed")
	}

	if err := viper.Unmarshal(&a.options); err != nil {
		log.WithError(err).Error("unmarshal config file error")
	}
}
