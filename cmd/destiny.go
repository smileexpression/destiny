package main

import (
	"smile.expression/destiny/pkg/app"
	"smile.expression/destiny/pkg/logger"
)

func main() {
	logger.Init()

	application := &app.App{}
	application.Init()
}
