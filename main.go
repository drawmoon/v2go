package main

import (
	"embed"

	app "github.com/yorukot/superfile/src/app"
)

var (
	//go:embed src/superfile_config/*
	content embed.FS
)

func main() {
	app.Run(content)
}
