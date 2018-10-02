package servers

// All the DNS server code came from OJ Reeves (@TheColonial)
// See https://www.youtube.com/watch?v=FeH2Yrw68f8

import (
	"net/http"
	"sort"

	"github.com/miekg/dns"
)

//
// Structs
//

// ServerSetting for various settings for DNS and HTTP servers
type ServerSetting struct {
	Value    string
	Default  string
	Required bool
	Help     string
}

// HttpServer struct, uses following map keys for modifiable settings
// "Listen":       listening IP address
// "Port":         listening port
// "DefaultPage":  default page returned with no key matchings
// "ServerHeader": option HTTP response 'Server' header
type HttpServer struct {
	Server  *http.Server
	State   map[string]*ServerSetting
	Keys    map[string]*Key
	Running bool
}

// DnsServer struct, uses following map keys for modifiable settings
// "Listen":     listening IP address
// "Domain":     the root level domain name we're authoratative over
// "DefaultTTL": default Time To Live (TTL) for DNS responses
type DnsServer struct {
	State      map[string]*ServerSetting
	Server     *dns.Server
	Keys       map[string]*Key
	DefaultTTL uint
	Running    bool
	SendingKey bool
}

//
// HTTP functions
//

// GetHttpServer returns a starting point for the HttpServer and
// HttpState structs for use throughout keyserver
func GetHttpServer() *HttpServer {

	state := make(map[string]*ServerSetting)

	state["Listen"] = &ServerSetting{
		Value:    "127.0.0.1",
		Default:  "127.0.0.1",
		Required: true,
		Help:     "The listening IP address. Requires root/admin for 0.0.0.0.",
	}

	state["Port"] = &ServerSetting{
		Value:    "8080",
		Default:  "8080",
		Required: true,
		Help:     "The port to listen on. Requires root/admin for ports < 1024.",
	}

	state["CertPath"] = &ServerSetting{
		Value:    "",
		Default:  "",
		Required: false,
		Help:     "Certificate to run an HTTPS server.",
	}

	state["KeyPath"] = &ServerSetting{
		Value:    "",
		Default:  "",
		Required: false,
		Help:     "Private key to run an HTTPS server.",
	}

	state["DefaultPage"] = &ServerSetting{
		Value:    "wwwroot/error.html",
		Default:  "",
		Required: false,
		Help:     "The default page to send for non-key requests. If empty, '404 Not Found' will be returned.",
	}

	return &HttpServer{
		State:   state,
		Running: false,
		Keys:    make(map[string]*Key),
	}
}

// StartHTTP is the exported function to call and get
// the HTTP server running.
func (h *HttpServer) StartHTTP() {
	mux := http.NewServeMux()
	mux.Handle("/", h)

	addr := h.State["Listen"].Value + ":" + h.State["Port"].Value
	// see net/http docs, this is where to set TLS up as well
	h.Server = &http.Server{Addr: addr, Handler: mux}
	h.Running = true

	go func() {
		if err := h.Server.ListenAndServe(); err != nil {
			h.Running = false
		}
	}()
}

// StartHTTPS is the exported function to call and get
// the HTTPS/TLS server running.
func (h *HttpServer) StartHTTPS() {
	mux := http.NewServeMux()
	mux.Handle("/", h)

	addr := h.State["Listen"].Value + ":" + h.State["Port"].Value
	h.Server = &http.Server{Addr: addr, Handler: mux}
	h.Running = true
	go func() {
		if err := h.Server.ListenAndServeTLS(h.State["CertPath"].Value, h.State["CertPath"].Value); err != nil {
			h.Running = false
		}
	}()
}

// GetDnsServer returns a starting point for the DnsServer and
// DnsState structs for use throughout keyserver
func GetDnsServer() *DnsServer {

	state := make(map[string]*ServerSetting)

	state["Listen"] = &ServerSetting{
		Value:    "127.0.0.1",
		Default:  "127.0.0.1",
		Required: true,
		Help:     "The listening IP address. Requires root/admin for 0.0.0.0.",
	}

	state["Port"] = &ServerSetting{
		Value:    "5333",
		Default:  "5333",
		Required: true,
		Help:     "The port to listen on. Requires root/admin for ports < 1024.",
	}

	state["Domain"] = &ServerSetting{
		Value:    "",
		Default:  "",
		Required: true,
		Help:     "The root domain name for the nameserver. Example: domain.com",
	}

	state["DefaultTTL"] = &ServerSetting{
		Value:    "10800",
		Default:  "10800",
		Required: true,
		Help:     "The default TTL response for DNS queries",
	}

	return &DnsServer{
		DefaultTTL: 10800,
		State:      state,
		Keys:       make(map[string]*Key),
		Running:    false,
	}
}

// StartDNS starts the DNS server in the background
func (d *DnsServer) StartDNS() {
	addr := d.State["Listen"].Value + ":" + d.State["Port"].Value
	d.Server = &dns.Server{Addr: addr, Net: "udp", Handler: d}
	d.Running = true
	go func() {
		if err := d.Server.ListenAndServe(); err != nil {
			d.Running = false
		}
	}()
}

// AlphabetizeSettings takes in a map of ServerSettings and returns
// the keys/names in alphabetical order.
func AlphabetizeSettings(settings map[string]*ServerSetting) []string {
	// Get map into alphabetical order
	keys := make([]string, len(settings))
	i := 0
	// fill temp array with keys of mis.Items
	for k := range settings {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	return keys
}
