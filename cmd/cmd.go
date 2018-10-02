package cmd

// Thanks to @evilsocket's Bettercap2 (https://github.com/bettercap/bettercap)
// for pointing me to github.com/chzyer/readline and having a good example to work off of

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/leoloobeek/keyserver/logger"
	"github.com/leoloobeek/keyserver/servers"
)

// CmdInfo holds all commands for a menu
type CmdInfo struct {
	MenuType      string
	Items         map[string]*MenuItem
	TabCompleters map[string]*readline.Instance
	HttpServer    *servers.HttpServer
	DnsServer     *servers.DnsServer
}

func (c *CmdInfo) MainMenu() {
	menuItems := c.getMainMenuItems()
	c.TabCompleters[c.MenuType].Config.AutoComplete = menuItems.Completer

	for {
		line, err := c.TabCompleters[c.MenuType].Readline()
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
		case "config":
			if len(words) != 2 {

			} else {
				if strings.ToLower(words[1]) == "dns" {
					c.MenuType = "Dns"
				} else if strings.ToLower(words[1]) == "http" {
					c.MenuType = "Http"
				} else {
					fmt.Printf("[!] Unknown menu type: %s\n", words[1])
				}
				return
			}
		case "start":
			errorMsg := "[!] Either start 'http' or 'dns'"
			if len(words) == 2 {
				if strings.ToLower(words[1]) == "http" {
					startHttpServer(c.HttpServer)
				} else if strings.ToLower(words[1]) == "dns" {
					startDnsServer(c.DnsServer)
				} else {
					fmt.Println(errorMsg)
				}
			} else {
				fmt.Println(errorMsg)
			}
		case "stop":
			errorMsg := "[!] Either stop 'http' or 'dns'"
			if len(words) == 2 {
				if strings.ToLower(words[1]) == "http" {
					stopHttpServer(c.HttpServer)
				} else if strings.ToLower(words[1]) == "dns" {
					stopDnsServer(c.DnsServer)
				} else {
					fmt.Println(errorMsg)
				}
			} else {
				fmt.Println(errorMsg)
			}
		case "restart":
			errorMsg := "[!] Either restart 'http' or 'dns'"
			if len(words) == 2 {
				if strings.ToLower(words[1]) == "http" {
					stopHttpServer(c.HttpServer)
					if !c.HttpServer.Running {
						startHttpServer(c.HttpServer)
					}
				} else if strings.ToLower(words[1]) == "dns" {
					stopDnsServer(c.DnsServer)
					if !c.DnsServer.Running {
						startDnsServer(c.DnsServer)
					}
				} else {
					fmt.Println(errorMsg)
				}
			} else {
				fmt.Println(errorMsg)
			}
		case "new":
			if len(words) != 2 {
				fmt.Println("[!] Either create an 'httpkey' or 'dnskey'")
			} else {
				if words[1] == "httpkey" {
					c.MenuType = "HttpKey"
					return
				}
				if words[1] == "dnskey" {
					c.MenuType = "DnsKey"
					return
				}
			}
		case "status":
			fmt.Println()
			running := "not running"
			if c.HttpServer.Running {
				running = "running"
			}
			fmt.Printf("HTTP: (%d keys, %s)\n", len(c.HttpServer.Keys), running)
			printKeys(c.HttpServer.Keys)

			running = "not running"
			if c.DnsServer.Running {
				running = "running"
			}
			fmt.Printf("DNS: (%d keys, %s)\n", len(c.DnsServer.Keys), running)
			printKeys(c.DnsServer.Keys)
			fmt.Println()
		case "info":
			if len(words) != 2 {
				fmt.Println("[!] Use `info <keyname>` to view details about a specific key")
			} else {
				httpKeyFound, dnsKeyFound := findKey(words[1], c.HttpServer.Keys, c.DnsServer.Keys)
				if httpKeyFound != "" {
					printKey(c.HttpServer.Keys[httpKeyFound], httpKeyFound)
				}
				if dnsKeyFound != "" {
					printKey(c.DnsServer.Keys[dnsKeyFound], dnsKeyFound)
				}
			}
		case "on":
			if len(words) != 2 {
				fmt.Println("[!] Use `on <keyname>` to manually turn on a key")
			} else {
				httpKeyFound, dnsKeyFound := findKey(words[1], c.HttpServer.Keys, c.DnsServer.Keys)
				if httpKeyFound != "" {
					c.HttpServer.Keys[httpKeyFound].On = true
					logger.Log.Noticef("[KEYCHANGE] - HTTP Key '%s' has been turned on!", httpKeyFound)
					if !c.HttpServer.Running {
						fmt.Println("[-] HTTP Server isn't running...")
					}
				}
				if dnsKeyFound != "" {
					c.DnsServer.Keys[dnsKeyFound].On = true
					logger.Log.Noticef("[KEYCHANGE] - DNS Key '%s' has been turned on!", dnsKeyFound)
					if !c.DnsServer.Running {
						fmt.Println("[-] DNS Server isn't running...")
					}
				}
			}
		case "off":
			if len(words) != 2 {
				fmt.Println("[!] Use `off <keyname>` to manually turn off a key, constraints will still turn it on")
			} else {
				httpKeyFound, dnsKeyFound := findKey(words[1], c.HttpServer.Keys, c.DnsServer.Keys)
				if httpKeyFound != "" {
					c.HttpServer.Keys[httpKeyFound].On = false
					logger.Log.Noticef("[KEYCHANGE] - HTTP Key '%s' has been turned off!", httpKeyFound)
					if !c.HttpServer.Running {
						fmt.Println("[-] HTTP Server isn't running...")
					}
				}
				if dnsKeyFound != "" {
					c.DnsServer.Keys[dnsKeyFound].On = false
					logger.Log.Noticef("[KEYCHANGE] - DNS Key '%s' has been turned off!", dnsKeyFound)
					if !c.DnsServer.Running {
						fmt.Println("[-] DNS Server isn't running...")
					}
				}
			}
		case "disable":
			if len(words) != 2 {
				fmt.Println("[!] Use `disable <keyname>` to disable a key indefinitely")
			} else {
				httpKeyFound, dnsKeyFound := findKey(words[1], c.HttpServer.Keys, c.DnsServer.Keys)
				if httpKeyFound != "" {
					c.HttpServer.Keys[httpKeyFound].On = false
					c.HttpServer.Keys[httpKeyFound].Disabled = true
					logger.Log.Noticef("[KEYCHANGE] - HTTP Key %s has been turned diabled! Constraints will have no effect.", httpKeyFound)
					if !c.HttpServer.Running {
						fmt.Println("[-] HTTP Server isn't running...")
					}
				}
				if dnsKeyFound != "" {
					c.DnsServer.Keys[dnsKeyFound].On = false
					c.DnsServer.Keys[dnsKeyFound].Disabled = true
					logger.Log.Noticef("[KEYCHANGE] - DNS Key %s has been turned disabled! Constraints will have no effect.", dnsKeyFound)
					if !c.DnsServer.Running {
						fmt.Println("[-] DNS Server isn't running...")
					}
				}
			}
		case "alert":
			if len(words) != 2 {
				fmt.Println("[!] Use `alert <keyname>` to turn on alerting for a key")
			} else {
				httpKeyFound, dnsKeyFound := findKey(words[1], c.HttpServer.Keys, c.DnsServer.Keys)
				if httpKeyFound != "" {
					c.HttpServer.Keys[httpKeyFound].SendAlerts = true
					fmt.Printf("[*] Alerting for %s enabled\n", httpKeyFound)
				}
				if dnsKeyFound != "" {
					c.DnsServer.Keys[dnsKeyFound].SendAlerts = true
					fmt.Printf("[*] Alerting for %s enabled\n", dnsKeyFound)
				}
			}
		case "noalert":
			if len(words) != 2 {
				fmt.Println("[!] Use `noalert <keyname>` to turn off alerting for a key")
			} else {
				httpKeyFound, dnsKeyFound := findKey(words[1], c.HttpServer.Keys, c.DnsServer.Keys)
				if httpKeyFound != "" {
					c.HttpServer.Keys[httpKeyFound].SendAlerts = false
					fmt.Printf("[*] Alerting for %s disabled\n", httpKeyFound)
				}
				if dnsKeyFound != "" {
					c.DnsServer.Keys[dnsKeyFound].SendAlerts = false
					fmt.Printf("[*] Alerting for %s disabled\n", dnsKeyFound)
				}
			}
		case "remove":
			if len(words) != 2 {
				fmt.Println("[!] Use `remove <keyname>` to remove a key")
			} else {
				httpKeyFound, dnsKeyFound := findKey(words[1], c.HttpServer.Keys, c.DnsServer.Keys)
				if httpKeyFound != "" {
					if response := askForPermission("[>] Remove this key? [y/N] "); response {
						delete(c.HttpServer.Keys, httpKeyFound)
					}
				}
				if dnsKeyFound != "" {
					if response := askForPermission("[>] Remove this key? [y/N] "); response {
						delete(c.DnsServer.Keys, dnsKeyFound)
					}
				}
			}
		case "clearhits":
			if len(words) != 2 {
				fmt.Println("[!] Use `clearhits <keyname>` to remove a key")
			} else {
				httpKeyFound, dnsKeyFound := findKey(words[1], c.HttpServer.Keys, c.DnsServer.Keys)
				if httpKeyFound != "" {
					c.HttpServer.Keys[httpKeyFound].ClearHits()
				}
				if dnsKeyFound != "" {
					c.DnsServer.Keys[dnsKeyFound].ClearHits()
				}
			}
		case "time":
			printCurrentTime()
		case "help":
			menuItems.printHelp()
		case "exit":
			c.MenuType = "Quit"
			return
		case "":
			continue
		default:
			fmt.Println("[!] Invalid command!")
		}
	}
}

func (c *CmdInfo) HttpMenu() {
	menuItems := getHttpMenuItems(c.HttpServer)
	c.TabCompleters[c.MenuType].Config.AutoComplete = menuItems.Completer

	for {
		line, err := c.TabCompleters[c.MenuType].Readline()
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
			startHttpServer(c.HttpServer)
		case "stop":
			stopHttpServer(c.HttpServer)
		case "restart":
			stopHttpServer(c.HttpServer)
			if !c.HttpServer.Running {
				startHttpServer(c.HttpServer)
			}
		case "info":
			printHttpStatus(c.HttpServer)
		case "unset":
			if len(words) == 2 {
				setting := strings.ToLower(words[1])

				found := false
				for k, v := range c.HttpServer.State {
					if strings.ToLower(k) == setting {
						v.Value = v.Default
						found = true
					}
				}
				if !found {
					fmt.Printf("[!] Server setting does not exist: " + words[1])
				}
			}
		case "set":
			if len(words) > 1 {
				setting := strings.ToLower(words[1])
				found := ""
				for k := range c.HttpServer.State {
					if strings.ToLower(k) == setting {
						found = k
						break
					}
				}
				if found != "" {
					errorMsg := fmt.Sprintf("[!] Invalid command, use 'help %s' for more info", found)
					switch found {
					case "CertPath", "KeyPath":
						if len(words) > 2 {
							filePath := strings.Join(words[2:], " ")
							// Check if file actually exists
							_, err := os.Stat(filePath)
							if err == nil {
								c.HttpServer.State[found].Value = filePath
							} else {
								fmt.Printf("[!] Error reading file: %s\n", err)
							}
						} else {
							fmt.Println(errorMsg)
						}
					default:
						// by default we will blindly set the value to word[2:] (everything after the second word on the line)
						if len(words) > 2 {
							c.HttpServer.State[found].Value = strings.Join(words[2:], " ")
						} else {
							fmt.Println(errorMsg)
						}
					}
				}
			}
		case "help":
			if len(words) != 2 {
				fmt.Println()
				menuItems.printHelp()
				fmt.Println("Use help <setting> to learn more about each setting")
				fmt.Println()
			} else {
				if _, ok := c.HttpServer.State[words[1]]; ok {
					fmt.Println()
					fmt.Println(c.HttpServer.State[words[1]].Help)
					fmt.Println()
				} else {
					fmt.Printf("[!] The DNS server setting %s does not exist!", words[1])
				}
			}
		case "exit", "back":
			c.MenuType = "Main"
			return
		case "":
			continue
		default:
			fmt.Println("[!] Invalid command!")
		}
	}
}

func (c *CmdInfo) DnsMenu() {
	menuItems := getDnsMenuItems(c.DnsServer)
	c.TabCompleters[c.MenuType].Config.AutoComplete = menuItems.Completer

	for {
		line, err := c.TabCompleters[c.MenuType].Readline()
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
			startDnsServer(c.DnsServer)
		case "stop":
			stopDnsServer(c.DnsServer)
		case "restart":
			stopDnsServer(c.DnsServer)
			if !c.DnsServer.Running {
				startDnsServer(c.DnsServer)
			}
		case "info":
			printDnsStatus(c.DnsServer)
		case "unset":
			if len(words) == 2 {
				setting := strings.ToLower(words[1])

				found := false
				for k, v := range c.DnsServer.State {
					if strings.ToLower(k) == setting {
						v.Value = v.Default
						found = true
					}
				}
				if !found {
					fmt.Printf("[!] Server setting does not exist: " + words[1])
				}
			}
		case "set":
			if len(words) > 1 {
				setting := strings.ToLower(words[1])
				found := ""
				for k := range c.DnsServer.State {
					if strings.ToLower(k) == setting {
						found = k
						break
					}
				}
				if found != "" {
					errorMsg := fmt.Sprintf("[!] Invalid command, use 'help %s' for more info", found)
					switch found {
					case "DefaultTTL":
						if len(words) == 3 {
							// Update the DefaultTTL uint attribute
							val, err := strconv.ParseUint(words[2], 10, 32)
							if err != nil {
								fmt.Printf("[!] %s is not a valid number\n", words[2])
							} else {
								c.DnsServer.DefaultTTL = uint(val)
								c.DnsServer.State[found].Value = words[2]
							}
						} else {
							fmt.Println(errorMsg)
						}
					default:
						// by default we will blindly set the value to word[2:] (everything after the second word on the line)
						if len(words) > 2 {
							c.DnsServer.State[found].Value = strings.Join(words[2:], " ")
						} else {
							fmt.Println(errorMsg)
						}
					}
				}
			}
		case "help":
			if len(words) != 2 {
				fmt.Println()
				menuItems.printHelp()
				fmt.Println("Use help <setting> to learn more about each setting")
				fmt.Println()
			} else {
				if _, ok := c.DnsServer.State[words[1]]; ok {
					fmt.Println()
					fmt.Println(c.DnsServer.State[words[1]].Help)
					fmt.Println()
				} else {
					fmt.Printf("[!] The DNS server setting %s does not exist!", words[1])
				}
			}
		case "exit", "back":
			c.MenuType = "Main"
			return
		case "":
			continue
		default:
			fmt.Println("[!] Invalid command!")
		}
	}
}

func (c *CmdInfo) HttpKeyMenu() {
	keyName := "NewHttpKey"
	key := &servers.Key{
		Type:       "http",
		Disabled:   false,
		SendAlerts: false,
		Data:       servers.HttpKeyData(),
		Hashes:     make(map[string]string),
		HitCounter: make(map[string]int),
	}
	key.HitCounter[servers.GetToday()] = 0
	key.Constraints = key.GetHttpKeyConstraints()

	menuItems := getHttpKeyMenuItems(key)
	c.TabCompleters[c.MenuType].Config.AutoComplete = menuItems.Completer

	for {
		line, err := c.TabCompleters[c.MenuType].Readline()
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
		case "info":
			printKeyMenuStatus(key, keyName)
		case "help":
			menuItems.printHelp()
		case "unset":
			if len(words) == 2 {
				setting := strings.ToLower(words[1])

				found := false
				for k, v := range key.Data {
					if strings.ToLower(k) == setting {
						v.Value = ""
						found = true
					}
				}
				if !found {
					for k, v := range key.Constraints {
						if strings.ToLower(k) == setting {
							v.Constraint = ""
							found = true
						}
					}
					if !found {
						fmt.Printf("[!] Setting does not exist: " + words[1])
					}
				}
			}
		case "set":
			if len(words) > 2 {
				setting := strings.ToLower(words[1])
				found := ""
				isConstraint := false
				for k := range key.Data {
					if strings.ToLower(k) == setting {
						found = k
						break
					}
				}
				for k := range key.Constraints {
					if strings.ToLower(k) == setting {
						found = k
						isConstraint = true
						break
					}
				}
				if setting == "name" {
					found = "name"
				}
				if found != "" {
					switch found {
					case "name":
						keyName = words[2]
					default:
						// by default we will blindly set the value to word[2:] (everything after the second word on the line)
						if isConstraint {
							key.Constraints[found].Constraint = strings.Join(words[2:], " ")
						} else {
							key.Data[found].Value = strings.Join(words[2:], " ")
						}
					}
				}
			} else {
				fmt.Println("[!] Invalid command, use `set <setting> <value`")
			}
		case "done":
			response := askForPermission("[>] Add this key? [y/N] ")
			if response {
				err := c.HttpServer.AddKey(key, keyName)
				if err == nil {
					c.MenuType = "Main"
					return
				}
				fmt.Printf("[!] Error adding key: %s\n", err)
			}
		case "exit", "back":
			c.MenuType = "Main"
			return
		case "":
			continue
		default:
			fmt.Println("[!] Invalid command!")
		}
	}
}

func (c *CmdInfo) DnsKeyMenu() {
	keyName := "NewDnsKey"
	key := &servers.Key{
		Type:       "dns",
		Disabled:   false,
		SendAlerts: false,
		Data:       servers.DnsKeyData(),
		Hashes:     make(map[string]string),
		HitCounter: make(map[string]int),
	}
	key.HitCounter[servers.GetToday()] = 0
	key.Constraints = key.GetDnsKeyConstraints()

	menuItems := getDnsKeyMenuItems(key)
	c.TabCompleters[c.MenuType].Config.AutoComplete = menuItems.Completer

	for {
		line, err := c.TabCompleters[c.MenuType].Readline()
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
		case "info":
			printKeyMenuStatus(key, keyName)
		case "help":
			menuItems.printHelp()
		case "unset":
			if len(words) == 2 {
				setting := strings.ToLower(words[1])

				found := false
				for k, v := range key.Data {
					if strings.ToLower(k) == setting {
						v.Value = ""
						found = true
					}
				}
				if !found {
					for k, v := range key.Constraints {
						if strings.ToLower(k) == setting {
							v.Constraint = ""
							found = true
						}
					}
					if !found {
						fmt.Printf("[!] Setting does not exist: " + words[1])
					}
				}
			}
		case "set":
			if len(words) > 2 {
				setting := strings.ToLower(words[1])
				found := ""
				isConstraint := false
				for k := range key.Data {
					if strings.ToLower(k) == setting {
						found = k
						break
					}
				}
				for k := range key.Constraints {
					if strings.ToLower(k) == setting {
						found = k
						isConstraint = true
						break
					}
				}
				if setting == "name" {
					found = "name"
				}
				if found != "" {
					switch found {
					case "name":
						keyName = words[2]
					default:
						// by default we will blindly set the value to word[2:] (everything after the second word on the line)
						if isConstraint {
							key.Constraints[found].Constraint = strings.Join(words[2:], " ")
						} else {
							key.Data[found].Value = strings.Join(words[2:], " ")
						}
					}
				}
			} else {
				fmt.Println("[!] Invalid command, use `set <setting> <value`")
			}
		case "done":
			response := askForPermission("[>] Add this key? [y/N] ")
			if response {
				err := c.DnsServer.AddKey(key, keyName)
				if err == nil {
					c.MenuType = "Main"
					return
				}
				fmt.Printf("[!] Error adding key: %s\n", err)
			}
		case "exit", "back":
			c.MenuType = "Main"
			return
		case "":
			continue
		default:
			fmt.Println("[!] Invalid command!")
		}
	}
}

func startHttpServer(h *servers.HttpServer) {
	if h.Running {
		fmt.Println("[!] HTTP server already running, use 'restart'")
		return
	}
	if h.State["CertPath"].Value != "" && h.State["KeyPath"].Value != "" {
		h.StartHTTPS()
	} else {
		h.StartHTTP()
	}
	time.Sleep(1 * time.Second)
	if !h.Running {
		fmt.Println("[!] Error occurred starting the HTTP server, port already in use?")
	} else {
		fmt.Println("[+] HTTP server successfully started!")
	}
}

func stopHttpServer(h *servers.HttpServer) {
	if !h.Running {
		fmt.Printf("[!] HTTP server isn't running\n")
		return
	}
	err := h.Server.Shutdown(nil)
	if err != nil {
		fmt.Printf("[!] Error shutting down HTTP gracefully: %s\n", err)
		return
	}
	time.Sleep(1 * time.Second)
	fmt.Println("[*] HTTP server stopped")
	h.Running = false
}

func printHttpStatus(h *servers.HttpServer) {
	fmt.Println()
	fmt.Println("HTTP Key Server")
	fmt.Printf("Running: %s\n", isRunning(h.Running))

	// Print modifiable settings
	settings := servers.AlphabetizeSettings(h.State)
	for _, name := range settings {
		fmt.Printf("    %s %s\n", columnString(name+returnAsterisk(h.State[name].Required)), h.State[name].Value)
	}
	fmt.Println()
}

func startDnsServer(d *servers.DnsServer) {
	if d.Running {
		fmt.Println("[!] DNS server already running, use 'restart'")
		return
	}
	d.StartDNS()
	time.Sleep(1 * time.Second)
	if !d.Running {
		fmt.Println("[!] Error occurred starting the DNS server, port already in use?")
	} else {
		fmt.Println("[+] DNS server successfully started!")
	}
}

func stopDnsServer(d *servers.DnsServer) {
	if !d.Running {
		fmt.Println("[!] DNS server isn't running")
		return
	}
	err := d.Server.Shutdown()
	if err != nil {
		fmt.Printf("[!] Error shutting down DNS gracefully: %s\n", err)
		return
	}
	time.Sleep(1 * time.Second)
	fmt.Println("[*] DNS server stopped")
	d.Running = false
}

func printDnsStatus(d *servers.DnsServer) {
	fmt.Println()
	fmt.Println("DNS Key Server")
	fmt.Printf("Running: %s\n", isRunning(d.Running))

	// Print modifiable settings
	settings := servers.AlphabetizeSettings(d.State)
	for _, name := range settings {
		fmt.Printf("    %s %s\n", columnString(name+returnAsterisk(d.State[name].Required)), columnString(d.State[name].Value))
	}
	fmt.Println()
}

// Prints status of menu items when selecting Key attributes
func printKeyMenuStatus(k *servers.Key, name string) {
	fmt.Println()
	fmt.Println("Key: ")
	fmt.Printf("    %s '%s'\n", columnString("Name:"), name)

	keyData := servers.AlphabetizeKeyData(k.Data)
	for _, name := range keyData {
		fmt.Printf("    %s '%s'\n", columnString(name+":"), k.Data[name].Value)
		fmt.Printf("        %s\n", k.Data[name].Description)
	}
	fmt.Println("\nConstraints:")
	constraints := servers.AlphabetizeConstraints(k.Constraints)
	for _, name := range constraints {
		fmt.Printf("    %s '%s'\n", columnString(name+":"), k.Constraints[name].Constraint)
		fmt.Printf("        %s\n", k.Constraints[name].Description)
	}
	fmt.Println()
}

// printKey shows more detail for one specific key by name
func printKey(key *servers.Key, name string) {
	fmt.Println()
	fmt.Printf("Name: %s (%s)\n", name, key.Type)

	if key.Type == "dns" {
		fmt.Printf("Hostname: %s\n", key.Data["Hostname"].Value)
		fmt.Printf("Record Type: %s\n", key.Data["RecordType"].Value)
		fmt.Printf("Response: %s\n", key.Data["Response"].Value)
		fmt.Printf("TTL: %s\n", key.Data["TTL"].Value)
	} else if key.Type == "http" {
		fmt.Printf("URL: %s\n", key.Data["URL"].Value)
		fmt.Printf("FilePath: %s\n", key.Data["FilePath"].Value)
	} else {
		fmt.Printf("[!] Unknown key type: %s\n", key.Type)
		return
	}

	fmt.Println()
	fmt.Println("Hashes of response:")
	for k, v := range key.Hashes {
		fmt.Printf("%s: '%s'\n", k, v)
	}

	fmt.Println()
	constraints := servers.AlphabetizeConstraints(key.Constraints)
	fmt.Println("Constraints:")
	for _, name := range constraints {
		fmt.Printf("    %s '%s'\n", columnString(name), key.Constraints[name].Constraint)
	}
	fmt.Println()
}

func printKeys(keys map[string]*servers.Key) {
	if len(keys) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("Keys ---")
	for name, key := range keys {
		fmt.Printf("    Name: %s\n", name)

		if key.Type == "dns" {
			fmt.Printf("    Hostname: %s\n", key.Data["Hostname"].Value)
		} else if key.Type == "http" {
			fmt.Printf("    URL: %s\n", key.Data["URL"].Value)
		} else {
			fmt.Printf("[!] Unknown key type: %s\n", key.Type)
			continue
		}
		fmt.Printf("    Hits Today: %d\n", key.GetHits())
		fmt.Printf("    Last Hit: %s\n", key.LastHit)
		if key.SendAlerts {
			fmt.Println("    Alerts: Enabled")
		} else {
			fmt.Println("    Alerts: Disabled")
		}

		active, reason := key.IsActive(nil, nil)
		if active {
			fmt.Printf("    Active: YES (%s)\n", reason)
		} else {
			if reason != "" {
				reason = "(" + reason + ")"
			}
			fmt.Printf("    Active: NO %s\n", reason)
		}
		fmt.Println()
	}
}

func printCurrentTime() {
	fmt.Println(time.Now().Format("Jan 2 15:04"))
}

func isRunning(result bool) string {
	if result {
		return "Yes"
	}
	return "No"
}

func returnAsterisk(result bool) string {
	if result {
		return "*:"
	}
	return ": "
}

func askForPermission(q string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(q)
	confirm, err := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)
	if confirm == "n" || confirm == "N" || err != nil {
		return false
	}
	return true
}

// searches by name for a http or dns key, returns (httpKeyName, dnsKeyName), each string is empty if not found
func findKey(input string, httpKeys map[string]*servers.Key, dnsKeys map[string]*servers.Key) (string, string) {
	var dnsKeyName string
	var httpKeyName string
	setting := strings.ToLower(input)

	for k := range httpKeys {
		if strings.ToLower(k) == setting {
			httpKeyName = k
		}
	}
	for k := range dnsKeys {
		if strings.ToLower(k) == setting {
			dnsKeyName = k
		}
	}

	return httpKeyName, dnsKeyName
}
