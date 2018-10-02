package servers

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/leoloobeek/keyserver/logger"
	"github.com/miekg/dns"
)

//
// HTTP Handling
//

// ServeHTTP allows SubHTTPServer to handle http requests
// The requested URL path needs to match a key's Data["URL"].Value to
// evaluate the Key
func (h *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// add cache control headers regardless of the response
	h.cacheHTTPHeaders(w)

	remoteAddr := parseProxyHeaders(r)

	// Log all requests
	logger.Log.Infof("[HTTP] - %s  \"%s %s\" \"%s\"", remoteAddr, r.Method, r.URL.Path, r.Header.Get("User-Agent"))
	// loop through all keys and see if any URL matches
	for name, key := range h.Keys {
		if r.URL.Path == key.Data["URL"].Value {
			// IsActive() will consider both manually setting the key and constraints
			if active, _ := key.IsActive(r, nil); active {
				fileBytes, err := ReadFile(key.Data["FilePath"].Value)
				if err != nil {
					logger.Log.Warningf("[ERROR] - Error reading HTML file: %s", err)
				} else {
					key.UpdateHits()
					msg := fmt.Sprintf("[HTTPKEY:ON] - Responding with active HTTP Key '%s'", name)
					logger.Log.Noticef(msg)
					if key.SendAlerts {
						logger.Alerts.SendAlerts(msg)
					}
					w.Write(fileBytes)
					return
				}
			} else {
				key.UpdateHits()
				msg := fmt.Sprintf("[HTTPKEY:OFF] - Access attempt for inactive HTTP Key '%s'", name)
				logger.Log.Warningf(msg)
				if key.SendAlerts {
					logger.Alerts.SendAlerts(msg)
				}
			}
		}
	}
	w.Write(h.getDefaultPage())
}

// getDefaultPage returns the default page bytes or '404 Not Found'
func (h *HttpServer) getDefaultPage() []byte {
	if h.State["DefaultPage"].Value != "" {
		fileBytes, err := ReadFile(h.State["DefaultPage"].Value)
		if err == nil {
			return fileBytes
		}
	}
	return []byte("404 Not Found")
}

// cacheHTTPHeaders adds some headers to prevent caching on the user side
func (h *HttpServer) cacheHTTPHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

// parseProxyHeaders helps keyserver work with redirectors
func parseProxyHeaders(r *http.Request) string {
	hdr := r.Header.Get("X-Forwarded-For")
	if hdr == "" {
		return strings.Split(r.RemoteAddr, ":")[0]
	}
	return hdr + " (P)"
}

//
// DNS Handling
//

// ServeDNS handles the DNS queries
func (d *DnsServer) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	for _, q := range r.Question {
		switch q.Qtype {
		case dns.TypeA:
			logger.Log.Infof("[DNS] - Received A query for %s", q.Name)
			resp, ttl, keyName := d.getActiveDNSKeys(&q)
			if resp != "" {
				ipResp := net.ParseIP(resp)
				if ipResp != nil {
					d.AppendResult(q, m, &dns.A{A: ipResp}, d.getTTL(ttl))
					w.WriteMsg(m)
					return
				}
				logger.Log.Warningf("[ERROR] - DNS Key '%s' response is not a valid IP, no response returned", keyName)
			}
			m.SetRcode(r, 3) // 3 - NXDomain  - Non-Existent Domain
		case dns.TypeTXT:
			logger.Log.Infof("[DNS] - Received TXT query for %s", q.Name)
			resp, ttl, _ := d.getActiveDNSKeys(&q)
			if resp != "" {
				d.AppendResult(q, m, &dns.TXT{Txt: []string{resp}}, d.getTTL(ttl))
				w.WriteMsg(m)
				return
			}
			m.SetRcode(r, 3) // 3 - NXDomain  - Non-Existent Domain
		}
	}
	w.WriteMsg(m)
}

// getActiveDNSKeys is leveraged by ServeDNS to get any active key responses back
// returns: DNS response, TTL, and key name
func (d *DnsServer) getActiveDNSKeys(q *dns.Question) (string, string, string) {
	hostname := strings.Split(q.Name, ".")[0]
	// loop through all keys and see if any record and hostname matches
	for name, key := range d.Keys {
		if hostname == key.Data["Hostname"].Value && q.Qtype == recordStringToUint(key.Data["RecordType"].Value) {
			// IsActive() will consider both manually setting the key and constraints
			if active, _ := key.IsActive(nil, q); active {
				key.UpdateHits()
				msg := fmt.Sprintf("[DNSKEY:ON] - Responding with active DNS Key '%s'", name)
				logger.Log.Noticef(msg)
				if key.SendAlerts {
					logger.Alerts.SendAlerts(msg)
				}
				return key.Data["Response"].Value, key.Data["TTL"].Value, name
			} else {
				key.UpdateHits()
				msg := fmt.Sprintf("[DNSKEY:OFF] - Access attempt for inactive DNS Key '%s'", name)
				logger.Log.Warningf(msg)
				if key.SendAlerts {
					logger.Alerts.SendAlerts(msg)
				}
			}
		}
	}
	return "", "", ""
}

// AppendResult prepares response for ServeDNS
// Taken directly from OJ's code
func (is *DnsServer) AppendResult(q dns.Question, m *dns.Msg, rr dns.RR, ttl uint) {
	hdr := dns.RR_Header{Name: q.Name, Class: q.Qclass, Ttl: uint32(ttl)}

	if rrS, ok := rr.(*dns.A); ok {
		hdr.Rrtype = dns.TypeA
		rrS.Hdr = hdr
	} else if rrS, ok := rr.(*dns.CNAME); ok {
		hdr.Rrtype = dns.TypeCNAME
		rrS.Hdr = hdr
	} else if rrS, ok := rr.(*dns.NS); ok {
		hdr.Rrtype = dns.TypeNS
		rrS.Hdr = hdr
	} else if rrS, ok := rr.(*dns.TXT); ok {
		hdr.Rrtype = dns.TypeTXT
		rrS.Hdr = hdr
	}

	if q.Qtype == dns.TypeANY || q.Qtype == rr.Header().Rrtype {
		m.Answer = append(m.Answer, rr)
	} else {
		m.Extra = append(m.Extra, rr)
	}

}

func recordStringToUint(record string) uint16 {
	switch record {
	case "TXT":
		return dns.TypeTXT
	case "A":
		return dns.TypeA
	default:
		return dns.TypeNone
	}
}

// If theres a TTL for the key, return that
func (d *DnsServer) getTTL(value string) uint {
	if value != "" {
		ttl, err := strconv.ParseUint(value, 10, 32)
		if err == nil {
			return uint(ttl)
		}
	}
	return d.DefaultTTL
}
