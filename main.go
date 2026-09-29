package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails 用 embed 把 frontend/dist 嵌进二进制,由内置资源服务器对外提供。

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 首次启动初始化:配置文件、默认目录、默认编辑器/终端
	if err := bootstrap(); err != nil {
		log.Fatal(err)
	}

	app := application.New(application.Options{
		Name:        "quickopen",
		Description: "项目快开:一键用编辑器/终端打开项目",
		Services: []application.Service{
			application.NewService(&ProjectService{}),
			application.NewService(&EditorService{}),
			application.NewService(&SettingsService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "项目快开",
		Width:            980,
		Height:           640,
		MinWidth:         720,
		MinHeight:        560,
		BackgroundColour: application.NewRGB(11, 12, 15),
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
