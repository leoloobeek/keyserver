package main

//
// keyserver - Leo Loobeek 2018
//

import (
	"fmt"

	"github.com/leoloobeek/keyserver/cmd"
	"github.com/leoloobeek/keyserver/logger"
	"github.com/leoloobeek/keyserver/servers"
)

func main() {
	fmt.Println()

	logger.Init()
	logger.Log.Info("Keyserver starting up...")

	c := cmd.CmdInfo{
		MenuType:      "Main",
		TabCompleters: cmd.InitializeCompleters(),
		HttpServer:    servers.GetHttpServer(),
		DnsServer:     servers.GetDnsServer(),
	}

Endless:
	for {
		switch c.MenuType {
		case "Main":
			c.MainMenu()
		case "Http":
			c.HttpMenu()
		case "Dns":
			c.DnsMenu()
		case "HttpKey":
			c.HttpKeyMenu()
		case "DnsKey":
			c.DnsKeyMenu()
		case "Quit":
			break Endless
		default:
			c.MenuType = "Main"
		}
	}
}
