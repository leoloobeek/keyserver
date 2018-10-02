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

/*

	// Setup DNS server settings
	is := &lib.DNSServer{
		State: &lib.ServerState{
			Listen:     "0.0.0.0:53",
			Domain:     "domain.com",
			DefaultTTL: 10800,
			KeyTTL:     180,
		},
		SendingKey: false,
	}
	// Setup HTTP server settings
	h := &lib.HTTPServer{
		Listen: ":80",
	}

	status(is)

	var completer = readline.NewPrefixCompleter(
		readline.PcItem("start",
			readline.PcItem("http"),
			readline.PcItem("dns"),
		),
		readline.PcItem("dnskey",
			readline.PcItem("on"),
			readline.PcItem("off"),
		),
		readline.PcItem("show"),
		readline.PcItem("exit"),
		readline.PcItem("add",
			readline.PcItem("dnskey"),
		),
		readline.PcItem("set",
			readline.PcItem("dnsdomain"),
		),
	)

	// setup readline
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[31mkeyserver >\033[0m ",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})

	if err != nil {
		panic(err)
	}
	defer l.Close()

Endless:
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		words := strings.Split(strings.TrimSpace(line), " ")

		switch words[0] {
		case "start":
			if len(words) == 2 {
				if words[1] == "http" {
					h.StartHTTP()
					log.Println("[+] HTTP server started!")
				} else if words[1] == "dns" {
					is.StartDNS()
					log.Println("[+] DNS server started!")
				}
			}
		case "stop":
			if len(words) == 2 {
				if words[1] == "http" {
					if err := h.Server.Shutdown(nil); err != nil {
						log.Printf("Error shutting down HTTP server: %s", err)
					} else {
						log.Println("HTTP server shutdown successfully")
					}
				} else if words[1] == "dns" {
					if err := is.Server.Shutdown(); err != nil {
						log.Printf("Error shutting down DNS server: %s", err)
					} else {
						log.Println("DNS server shutdown successfully")
					}
				}
			}
		case "add":
			if len(words) > 3 {
				if words[1] == "dnskey" {
					hostname := words[2]
					value := words[3]
					if len(words) > 4 {
						value = strings.Join(words[3:], " ")
					}
					sha256 := lib.GenerateSHA256(value)
					sha512 := lib.GenerateSHA512(value)
					is.Key = &lib.DnsKey{
						Hostname: hostname,
						Value:    value,
						Sha256:   sha256,
						Sha512:   sha512,
					}
					fmt.Println("[+] DnsKey added!")
				}
			} else {
				fmt.Println("[!] Syntax: add dnskey <hostname> <values >")
			}
		case "dnskey":
			if len(words) == 2 {
				if words[1] == "on" {
					log.Println("[*] Responding with DnsKey to valid queries")
					is.SendingKey = true
				} else if words[1] == "off" {
					log.Println("[*] No longer responding with DnsKey")
					is.SendingKey = false
				}
			}
		case "show":
			status(is)
		case "set":
			if len(words) == 3 {
				if words[1] == "dnsdomain" {
					is.State.Domain = words[2]
				}
			}
		case "quit", "exit":
			break Endless
		case "":
			continue
		default:
			fmt.Printf("[!] Unknown menu option: %s. Try 'help'\n", words[0])
		}
	}
}

func status(is *lib.DNSServer) {
	fmt.Println()
	fmt.Println("DNS Domain: " + is.State.Domain)
	if is.Key != nil {
		fmt.Println("DnsKey Hostname: " + is.Key.Hostname)
		fmt.Println("DnsKey Value: " + is.Key.Value)
		fmt.Println("DnsKey Sha256: " + is.Key.Sha256)
		fmt.Println("DnsKey Sha512: " + is.Key.Sha512)
		fmt.Println("Active: " + strconv.FormatBool(is.SendingKey))
	}
	fmt.Println()
}
*/
