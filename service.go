package main

import (
	"gateway/simon/api"
	"gateway/simon/server"
)

func main() {
	serivce := server.Init()
	app := api.Api{serivce}
	app.Run()
}
