package main

import (
	_ "smile.expression/destiny/logger"
	"smile.expression/destiny/pkg/app"
)

func main() {
	application := &app.App{}
	application.Init()
}
