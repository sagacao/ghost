package main

import (
	_ "ghost/global"
	_ "ghost/models"
	_ "ghost/network"
	"ghost/services"
)

func main() {
	services.RunTCPServer()
	select {}
}
