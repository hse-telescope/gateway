package main

import (
	"os"

	"github.com/hse-telescope/gateway/internal/app"
	"github.com/hse-telescope/gateway/internal/config"
)

func main() {
	configPath := os.Args[1]
	conf, err := config.Parse(configPath)
	if err != nil {
		panic(err)
	}

	app := app.New(conf)
	panic(app.Start())
}
