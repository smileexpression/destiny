package main

import (
	"smile.expression/destiny/logger"
	"smile.expression/destiny/pkg/app"
)

func main() {
	logger.Init()

	application := &app.App{}
	application.Init()
}
