package cmd

// Thanks to @evilsocket's Bettercap2 (https://github.com/bettercap/bettercap)
// for pointing me to github.com/chzyer/readline and having a good example to work off of

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/chzyer/readline"
	"github.com/leoloobeek/keyserver/servers"
)

// MenuItems holds all commands for a menu
type MenuItems struct {
	MenuType  string
	Items     map[string]*MenuItem
	Completer *readline.PrefixCompleter
}

// MenuItem describes each command, help and usage, and completion
type MenuItem struct {
	Help      string
	Example   string
	Error     string
	Completer *readline.PrefixCompleter
}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func InitializeCompleters() map[string]*readline.Instance {
	// initialize tab completers

	// MainMenu
	mmInst, err := readline.NewEx(&readline.Config{
		Prompt:          "keyserver > ",
		AutoComplete:    nil,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})

	// HttpMenu
	hmInst, err := readline.NewEx(&readline.Config{
		Prompt:          "keyserver (http) > ",
		AutoComplete:    nil,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	// DnsMenu
	dmInst, err := readline.NewEx(&readline.Config{
		Prompt:          "keyserver (dns) > ",
		AutoComplete:    nil,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})

	// HttpKeyMenu
	hkmInst, err := readline.NewEx(&readline.Config{
		Prompt:          "keyserver (httpkey) > ",
		AutoComplete:    nil,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})

	// DnsKeyMenu
	dkmInst, err := readline.NewEx(&readline.Config{
		Prompt:          "keyserver (dnskey) > ",
		AutoComplete:    nil,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})

	if err != nil {
		panic(err)
	}

	return map[string]*readline.Instance{
		"Main":    mmInst,
		"Http":    hmInst,
		"Dns":     dmInst,
		"HttpKey": hkmInst,
		"DnsKey":  dkmInst,
	}
}

func (c *CmdInfo) getMainMenuItems() *MenuItems {

	items := defaultItems()

	items["config"] = &MenuItem{
		Help:    "Configure a server (http or dns)",
		Example: "config http",
		Completer: readline.NewPrefixCompleter(
			readline.PcItem("http"),
			readline.PcItem("dns"),
		),
	}

	items["start"] = &MenuItem{
		Help:    "Start a server (http or dns)",
		Example: "start http",
		Completer: readline.NewPrefixCompleter(
			readline.PcItem("http"),
			readline.PcItem("dns"),
		),
	}

	items["stop"] = &MenuItem{
		Help:    "Stop a server (http or dns)",
		Example: "stop http",
		Completer: readline.NewPrefixCompleter(
			readline.PcItem("http"),
			readline.PcItem("dns"),
		),
	}

	items["restart"] = &MenuItem{
		Help:    "Restart a server (http or dns)",
		Example: "restart http",
		Completer: readline.NewPrefixCompleter(
			readline.PcItem("http"),
			readline.PcItem("dns"),
		),
	}

	items["new"] = &MenuItem{
		Help:    "Create a new dnskey or httpkey",
		Example: "new httpkey",
		Completer: readline.NewPrefixCompleter(
			readline.PcItem("httpkey"),
			readline.PcItem("dnskey"),
		),
	}

	items["status"] = &MenuItem{
		Help:      "Show status of servers and keys",
		Example:   "status",
		Completer: readline.NewPrefixCompleter(),
	}

	items["info"] = &MenuItem{
		Help:      "Show detailed information for a specific key",
		Example:   "info <keyname>",
		Completer: readline.NewPrefixCompleter(readline.PcItemDynamic(c.getAllKeys())),
	}

	items["on"] = &MenuItem{
		Help:      "Manually turn on a specific key",
		Example:   "on <keyname>",
		Completer: readline.NewPrefixCompleter(readline.PcItemDynamic(c.getAllKeys())),
	}

	items["off"] = &MenuItem{
		Help:      "Manually turn off a specific key",
		Example:   "off <keyname>",
		Completer: readline.NewPrefixCompleter(readline.PcItemDynamic(c.getAllKeys())),
	}

	items["disable"] = &MenuItem{
		Help:      "Disable a key and never respond with active key, even if a constraint matches",
		Example:   "off <keyname>",
		Completer: readline.NewPrefixCompleter(readline.PcItemDynamic(c.getAllKeys())),
	}

	items["alert"] = &MenuItem{
		Help:      "Enable alerting for a key",
		Example:   "alert <keyname>",
		Completer: readline.NewPrefixCompleter(readline.PcItemDynamic(c.getAllKeys())),
	}

	items["noalert"] = &MenuItem{
		Help:      "Disable alerting for a key",
		Example:   "noalert <keyname>",
		Completer: readline.NewPrefixCompleter(readline.PcItemDynamic(c.getAllKeys())),
	}

	items["remove"] = &MenuItem{
		Help:      "Remove a key",
		Example:   "remove <keyname>",
		Completer: readline.NewPrefixCompleter(readline.PcItemDynamic(c.getAllKeys())),
	}

	items["clearhits"] = &MenuItem{
		Help:      "Disable alerting for a key",
		Example:   "noalert <keyname>",
		Completer: readline.NewPrefixCompleter(readline.PcItemDynamic(c.getAllKeys())),
	}

	items["time"] = &MenuItem{
		Help:      "Display current time on keyserver (useful when setting time constraints)",
		Example:   "time",
		Completer: readline.NewPrefixCompleter(),
	}

	completer := []readline.PrefixCompleterInterface{}
	for name, mi := range items {
		item := readline.PcItem(name)
		item.Children = mi.Completer.Children
		completer = append(completer, item)
	}

	return &MenuItems{
		MenuType:  "Main",
		Items:     items,
		Completer: readline.NewPrefixCompleter(completer...),
	}
}

func getHttpMenuItems(h *servers.HttpServer) *MenuItems {

	items := getConfigMenuItems(h.State)

	// Update help's completer with DNS settings
	items["help"].Completer = readline.NewPrefixCompleter(readline.PcItemDynamic(getSettings(h.State)))

	completer := []readline.PrefixCompleterInterface{}
	for name, mi := range items {
		item := readline.PcItem(name)
		item.Children = mi.Completer.Children
		completer = append(completer, item)
	}

	return &MenuItems{
		MenuType:  "Http",
		Items:     items,
		Completer: readline.NewPrefixCompleter(completer...),
	}
}

func getDnsMenuItems(d *servers.DnsServer) *MenuItems {

	items := getConfigMenuItems(d.State)

	// Update help's completer with DNS settings
	items["help"].Completer = readline.NewPrefixCompleter(readline.PcItemDynamic(getSettings(d.State)))

	completer := []readline.PrefixCompleterInterface{}
	for name, mi := range items {
		item := readline.PcItem(name)
		item.Children = mi.Completer.Children
		completer = append(completer, item)
	}

	return &MenuItems{
		MenuType:  "Dns",
		Items:     items,
		Completer: readline.NewPrefixCompleter(completer...),
	}
}

// These menu items will consist between both http and dns config menus
func getConfigMenuItems(ss map[string]*servers.ServerSetting) map[string]*MenuItem {

	items := defaultItems()

	items["info"] = &MenuItem{
		Help:      "Show all settings",
		Example:   "info",
		Completer: readline.NewPrefixCompleter(),
	}

	items["set"] = &MenuItem{
		Help:      "Choose settings",
		Example:   "set Listen 0.0.0.0",
		Completer: readline.NewPrefixCompleter(readline.PcItemDynamic(getSettings(ss))),
	}

	items["unset"] = &MenuItem{
		Help:      "Remote settings",
		Example:   "unset Listen",
		Completer: readline.NewPrefixCompleter(readline.PcItemDynamic(getSettings(ss))),
	}

	items["start"] = &MenuItem{
		Help:      "Start the server",
		Example:   "start",
		Completer: readline.NewPrefixCompleter(readline.PcItemDynamic(getSettings(ss))),
	}

	items["restart"] = &MenuItem{
		Help:      "Restart the server",
		Example:   "restart",
		Completer: readline.NewPrefixCompleter(readline.PcItemDynamic(getSettings(ss))),
	}

	items["stop"] = &MenuItem{
		Help:      "Stop the server",
		Example:   "stop",
		Completer: readline.NewPrefixCompleter(readline.PcItemDynamic(getSettings(ss))),
	}

	return items
}

func getHttpKeyMenuItems(k *servers.Key) *MenuItems {

	items := getKeyMenuItems(k)

	completer := []readline.PrefixCompleterInterface{}
	for name, mi := range items {
		item := readline.PcItem(name)
		item.Children = mi.Completer.Children
		completer = append(completer, item)
	}

	return &MenuItems{
		MenuType:  "HttpKey",
		Items:     items,
		Completer: readline.NewPrefixCompleter(completer...),
	}
}

func getDnsKeyMenuItems(k *servers.Key) *MenuItems {

	items := getKeyMenuItems(k)

	completer := []readline.PrefixCompleterInterface{}
	for name, mi := range items {
		item := readline.PcItem(name)
		item.Children = mi.Completer.Children
		completer = append(completer, item)
	}

	return &MenuItems{
		MenuType:  "DnsKey",
		Items:     items,
		Completer: readline.NewPrefixCompleter(completer...),
	}
}

// These menu items will consist between both http and dns key config menus
func getKeyMenuItems(k *servers.Key) map[string]*MenuItem {

	items := defaultItems()

	items["done"] = &MenuItem{
		Help:      "Once key is set, finish and add it to the server's keys",
		Example:   "done",
		Error:     "",
		Completer: readline.NewPrefixCompleter(),
	}

	items["info"] = &MenuItem{
		Help:      "Show all settings",
		Example:   "info",
		Error:     "",
		Completer: readline.NewPrefixCompleter(),
	}

	items["set"] = &MenuItem{
		Help:      "Choose settings or constraints",
		Example:   "set Name MyKey",
		Error:     "Did not choose a valid setting",
		Completer: readline.NewPrefixCompleter(readline.PcItemDynamic(getSettingsAndConstraints(k))),
	}

	items["unset"] = &MenuItem{
		Help:      "Remove settings or constraints",
		Example:   "unset Time",
		Error:     "Did not choose a valid setting",
		Completer: readline.NewPrefixCompleter(readline.PcItemDynamic(getSettingsAndConstraints(k))),
	}

	return items
}

/*
func (m *MenuInformation) getDropperMenuItems() *MenuItems {
	items := defaultItems()

	dropperSettings := readline.NewPrefixCompleter(
		readline.PcItem("SonarName", readline.PcItemDynamic(m.getRunningSonarNames())),
		readline.PcItem("Filename"),
		readline.PcItem("Lang"),
		readline.PcItem("Format"),
		readline.PcItem("StagingURL"),
		readline.PcItem("UserAgent"),
		readline.PcItem("CustomHeaders"),
	)

	items["help"].Completer = dropperSettings

	items["info"] = &MenuItem{
		Help:      "Show sonar info and settings (alias for 'show sonarinfo'",
		Example:   "info",
		Completer: readline.NewPrefixCompleter(),
	}

	items["show"] = &MenuItem{
		Help:    "Show langs or formats that can be set",
		Example: "show langs",
		Error:   "Specify an option to show (langs,formats)",
		Completer: readline.NewPrefixCompleter(
			readline.PcItem("langs"),
			readline.PcItem("formats"),
		),
	}

	items["set"] = &MenuItem{
		Help:      "Set a language or dropper",
		Example:   "set Format mshta",
		Error:     "Options to set: Filename, Lang, Format, StagingURL, UserAgent, CustomHeaders",
		Completer: dropperSettings,
	}

	items["generate"] = &MenuItem{
		Help:      "Generate a dropper based on the set options",
		Example:   "generate",
		Error:     "Error generating dropper file, are you missing required options?",
		Completer: readline.NewPrefixCompleter(),
	}

	completer := []readline.PrefixCompleterInterface{}
	for name, mi := range items {
		item := readline.PcItem(name)
		item.Children = mi.Completer.Children
		completer = append(completer, item)
	}

	return &MenuItems{
		MenuType:  "MainMenu",
		Items:     items,
		Completer: readline.NewPrefixCompleter(completer...),
	}
}
*/
//
// Helpers
//

// TODO: Make this better, tab complete directories/etc.
func listFiles() func(string) []string {
	return func(line string) []string {
		path := "./"
		names := make([]string, 0)
		files, _ := ioutil.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
	}
}

/*
func  getModulePaths() func(string) []string {
	return func(line string) []string {
		var result []string
		for _, modulePath := range m.SubInfo.ModulePaths {
			result = append(result, strings.TrimSuffix(strings.TrimPrefix(modulePath, "modules/"), ".json"))
		}
		return result
	}
}
*/

func getSettingsAndConstraints(k *servers.Key) func(string) []string {
	return func(line string) []string {
		result := []string{"Name"}
		for name, _ := range k.Data {
			result = append(result, name)
		}
		for name, _ := range k.Constraints {
			result = append(result, name)
		}
		return result
	}
}

func (c *CmdInfo) getAllKeys() func(string) []string {
	return func(line string) []string {
		var result []string
		for name, _ := range c.HttpServer.Keys {
			result = append(result, name)
		}
		for name, _ := range c.DnsServer.Keys {
			result = append(result, name)
		}
		return result
	}
}

func getSettings(settings map[string]*servers.ServerSetting) func(string) []string {
	return func(line string) []string {
		var result []string
		for name, _ := range settings {
			result = append(result, name)
		}
		return result
	}
}

func defaultItems() map[string]*MenuItem {
	items := make(map[string]*MenuItem)

	items["help"] = &MenuItem{
		Help:      "Displays all commands with help description",
		Example:   "help",
		Completer: readline.NewPrefixCompleter(),
	}

	items["back"] = &MenuItem{
		Help:      "Go back a menu",
		Example:   "back",
		Completer: readline.NewPrefixCompleter(),
	}

	items["exit"] = &MenuItem{
		Help:      "Exits keyserver. The alias 'quit' also works",
		Example:   "exit",
		Completer: readline.NewPrefixCompleter(),
	}

	return items
}

func (mis *MenuItems) printHelp() {
	// Get map into alphabetical order
	keys := make([]string, len(mis.Items))
	i := 0
	// fill temp array with keys of mis.Items
	for k := range mis.Items {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	// Print out
	fmt.Println("Commands:")
	for _, key := range keys {
		fmt.Printf("\t%s%s\n", columnString(key), mis.Items[key].Help)
	}
}

func columnString(str string) string {
	if len(str) > 18 {
		return str[:18] + "  "
	}
	return str + (strings.Repeat(" ", (19 - len(str))))
}
