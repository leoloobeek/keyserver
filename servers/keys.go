package servers

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// Key contains attributes that fit both Http and Dns keys
type Key struct {
	Type        string
	On          bool
	Disabled    bool
	SendAlerts  bool
	HitCounter  map[string]int
	LastHit     string
	Data        map[string]*KeyData
	Constraints map[string]*KeyConstraint
	Hashes      map[string]string
}

type KeyData struct {
	Description string
	Value       string
}

type KeyConstraint struct {
	Description     string
	Constraint      string
	ConstraintRegex *regexp.Regexp
	HttpValidator   func(constraint string, r *http.Request) bool
	DnsValidator    func(constraint string, q *dns.Question) bool
}

//
// HTTP Key
//

func HttpKeyData() map[string]*KeyData {
	data := make(map[string]*KeyData)

	data["FilePath"] = &KeyData{
		Description: "The path to the file to serve (html will be used as key)",
		Value:       "wwwroot/file.html",
	}

	data["URL"] = &KeyData{
		Description: "The URL of the HTTP request",
		Value:       "/content/file.html",
	}

	return data
}

// GetHttpKeyConstraints returns all possible key constraints for an HttpKey
func (k *Key) GetHttpKeyConstraints() map[string]*KeyConstraint {
	constraints := make(map[string]*KeyConstraint)

	constraints["Time"] = &KeyConstraint{
		Description:     "Turn on key within timeframe by minutes: (00:00-23:59)",
		Constraint:      "",
		ConstraintRegex: regexp.MustCompile("^([0-9]|0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]-([0-9]|0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$"),
		HttpValidator:   k.TimeHttpConstraint,
	}

	constraints["HitLimit"] = &KeyConstraint{
		Description:     "Turn off the key after a certain number of hits are received",
		Constraint:      "",
		ConstraintRegex: regexp.MustCompile("^[0-9]+$"),
		HttpValidator:   k.HitLimitHttpConstraint,
	}

	constraints["HitMax"] = &KeyConstraint{
		Description:     "Turn on the key after a certain number of hits are received",
		Constraint:      "",
		ConstraintRegex: regexp.MustCompile("^[0-9]+$"),
		HttpValidator:   k.HitMaxHttpConstraint,
	}

	return constraints
}

// TimeHttpConstraint is a key constraint that returns true if the current time falls
// within a specified timeframe. http.Request data not needed as we just want the current time.
func (k *Key) TimeHttpConstraint(constraint string, r *http.Request) bool {
	return timeConstraint(constraint)
}

// HitLimitHttpConstraint is a key constraint that returns true if the number of hits
// is below the supplied limit
func (k *Key) HitLimitHttpConstraint(constraint string, r *http.Request) bool {
	limit, err := strconv.Atoi(constraint)
	if err == nil {
		if k.GetHits() < limit {
			return true
		}
	}
	return false
}

// HitMaxHttpConstraint is a key constraint that returns true if the number of hits
// is above the supplied value
func (k *Key) HitMaxHttpConstraint(constraint string, r *http.Request) bool {
	value, err := strconv.Atoi(constraint)
	if err == nil {
		if k.GetHits() > value {
			return true
		}
	}
	return false
}

// UserAgentConstraint is a key constraint that returns true if the current time falls
// within a specified timeframe. requestData is null as the data we want is the current time.
func (k *Key) UserAgentHttpConstraint(constraint string, r *http.Request) bool {
	if r == nil {
		return false
	}
	if constraint == r.Header.Get("User-Agent") {
		return true
	}
	return false
}

// AddKey does the fun stuff, takes in the data generates the hasehs and adds
// it to the end of the Keys slice within the HttpServer
func (h *HttpServer) AddKey(k *Key, name string) error {
	if strings.Contains(name, " ") {
		return errors.New("Key name contains spaces")
	}
	if _, exists := h.Keys[name]; exists {
		return errors.New("Key name already exists!")
	}

	if err := validateKeyConstraints(k.Constraints); err != nil {
		return err
	}

	fileContents, err := ReadFile(k.Data["FilePath"].Value)
	if err != nil {
		return err
	}
	k.Hashes = BuildKey(string(fileContents))

	h.Keys[name] = k
	return nil
}

//
// DNS Key
//

func DnsKeyData() map[string]*KeyData {
	data := make(map[string]*KeyData)

	data["Hostname"] = &KeyData{
		Description: "The hostname for the DNS request, do not include FQDN",
		Value:       "mail",
	}

	data["RecordType"] = &KeyData{
		Description: "The record type: A or TXT",
		Value:       "TXT",
	}

	data["Response"] = &KeyData{
		Description: "The response to send back for the request.",
		Value:       "",
	}

	data["TTL"] = &KeyData{
		Description: "The TTL for the DNS response (in seconds). Important to keep smaller than the delay if using retries.",
		Value:       "180",
	}

	return data
}

func (k *Key) GetDnsKeyConstraints() map[string]*KeyConstraint {
	constraints := make(map[string]*KeyConstraint)

	constraints["Time"] = &KeyConstraint{
		Description:     "Turn on key within timeframe by minutes: (00:00-23:59)",
		Constraint:      "",
		ConstraintRegex: regexp.MustCompile("^([0-9]|0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]-([0-9]|0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$"),
		DnsValidator:    k.TimeDnsConstraint,
	}

	constraints["HitLimit"] = &KeyConstraint{
		Description:     "Turn off the key after a certain number of hits are received",
		Constraint:      "",
		ConstraintRegex: regexp.MustCompile("^[0-9]+$"),
		DnsValidator:    k.HitLimitDnsConstraint,
	}

	constraints["HitMax"] = &KeyConstraint{
		Description:     "Turn on the key after a certain number of hits are received",
		Constraint:      "",
		ConstraintRegex: regexp.MustCompile("^[0-9]+$"),
		DnsValidator:    k.HitMaxDnsConstraint,
	}

	return constraints

}

// TimeConstraint is a key constraint that returns true if the current time falls
// within a specified timeframe. DNS request data not needed as we just want the current time.
func (k *Key) TimeDnsConstraint(constraint string, q *dns.Question) bool {
	return timeConstraint(constraint)
}

// HitLimitDnsConstraint is a key constraint that returns true if the number of hits
// is below the supplied limit
func (k *Key) HitLimitDnsConstraint(constraint string, q *dns.Question) bool {
	limit, err := strconv.Atoi(constraint)
	if err == nil {
		if k.GetHits() < limit {
			return true
		}
	}
	return false
}

// HitMaxDnsConstraint is a key constraint that returns true if the number of hits
// is above the supplied value
func (k *Key) HitMaxDnsConstraint(constraint string, q *dns.Question) bool {
	value, err := strconv.Atoi(constraint)
	if err == nil {
		if k.GetHits() > value {
			return true
		}
	}
	return false
}

// AddKey does the fun stuff, takes in the data generates the hasehs and adds
// it to the end of the Keys slice within the HttpServer
func (d *DnsServer) AddKey(k *Key, name string) error {
	if strings.Contains(name, " ") {
		return errors.New("Key name contains spaces")
	}
	if _, exists := d.Keys[name]; exists {
		return errors.New("Key name already exists!")
	}

	if err := validateKeyConstraints(k.Constraints); err != nil {
		return err
	}

	k.Hashes = BuildKey(k.Data["Response"].Value)

	// remove unused constraints
	for name, _ := range k.Constraints {
		if k.Constraints[name].Constraint == "" {
			delete(k.Constraints, name)
		}
	}
	d.Keys[name] = k
	return nil
}

//
// Key functions for HTTP and DNS
//

// IsActive determines whether a key is active for the HttpServer
// The string returned is the "reason" the key is active or inactive, manually turned
// on or due to a constraint
func (k *Key) IsActive(r *http.Request, q *dns.Question) (bool, string) {
	if k.Disabled {
		return false, "disabled"
	}

	var reasons []string
	var active bool
	if k.On {
		active = true
		reasons = append(reasons, "Manual")
	}

	for name, _ := range k.Constraints {
		if k.Type == "http" {
			if k.Constraints[name].HttpValidator(k.Constraints[name].Constraint, r) {
				active = true
				reasons = append(reasons, name)
			}
		} else if k.Type == "dns" {
			if k.Constraints[name].DnsValidator(k.Constraints[name].Constraint, q) {
				active = true
				reasons = append(reasons, name)
			}

		}
	}

	return active, strings.Join(reasons, ", ")
}

// timeConstraint is handled by both DNS and HTTP TimeConstraint methods
func timeConstraint(constraint string) bool {
	layout := "15:04"
	startTime, err := time.Parse(layout, strings.Split(constraint, "-")[0])
	if err != nil {
		return false
	}
	endTime, err := time.Parse(layout, strings.Split(constraint, "-")[1])
	if err != nil {
		return false
	}
	nowStr := time.Now().Format("15:04")
	nowTime, err := time.Parse(layout, nowStr)
	if err != nil {
		return false
	}
	// This allows us to be inclusive of the times
	startTime = startTime.Add(-(time.Millisecond * 1))
	endTime = endTime.Add(time.Millisecond * 1)

	if nowTime.After(startTime) && nowTime.Before(endTime) {
		return true
	}
	return false
}

//
// Hashing stuff
//

// BuildKey takes the string data and generates all supported hashes
// A map with the hash type as key and hash as value is returned
func BuildKey(s string) map[string]string {
	return map[string]string{
		"sha512": GenerateSHA512(s),
		//"sha256": GenerateSHA256(s), - holding off on supporting multiple hashing types for now
	}
}

// GenerateSHA512 takes a string, generates a SHA512 hash
// and sends back as hex string
func GenerateSHA512(s string) string {
	sha := sha512.New()
	sha.Write([]byte(s))

	return hex.EncodeToString(sha.Sum(nil))
}

// GenerateSHA256 takes a string, generates a SHA256 hash
// and sends back as hex string
func GenerateSHA256(s string) string {
	sha := sha256.New()
	sha.Write([]byte(s))

	return hex.EncodeToString(sha.Sum(nil))
}

//
// Helpers
//

// GetHits returns the hit counter but ensures theres an entry for
// the current day.
func (k *Key) GetHits() int {
	today := GetToday()
	if _, exists := k.HitCounter[today]; !exists {
		k.HitCounter[today] = 0
	}
	return k.HitCounter[today]

}

// UpdateHits updates the HitCounter for the current day
func (k *Key) UpdateHits() {
	today := GetToday()
	if _, exists := k.HitCounter[today]; !exists {
		k.HitCounter[today] = 0
	}
	k.HitCounter[today]++
	k.LastHit = time.Now().Format("01/02/2006 15:04:05")
}

// ClearHits sets the current day to 0 hits
func (k *Key) ClearHits() {
	today := GetToday()
	if _, exists := k.HitCounter[today]; !exists {
		k.HitCounter[today] = 0
	}
	k.HitCounter[today] = 0
}

func GetToday() string {
	return time.Now().Format("01/02/2006")
}

// ReadFile returns the contents as a []byte
func ReadFile(path string) ([]byte, error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}

// AlphabetizeKeyData takes in a map of KeyData and returns
// the keys/names in alphabetical order.
func AlphabetizeKeyData(keyData map[string]*KeyData) []string {
	// Get map into alphabetical order
	keys := make([]string, len(keyData))
	i := 0
	// fill temp array with keys of mis.Items
	for k := range keyData {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	return keys
}

// AlphabetizeConstraints takes in a map of KeyData and returns
// the keys/names in alphabetical order.
func AlphabetizeConstraints(constraints map[string]*KeyConstraint) []string {
	// Get map into alphabetical order
	keys := make([]string, len(constraints))
	i := 0
	// fill temp array with keys of mis.Items
	for k := range constraints {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	return keys
}

// validateKeyConstraints loops through all key constraints, and if value is not empty,
// ensures the value matches the constraint's regex. This is used when attempting to add
// a key. If one fails, we just return that error, for now.
func validateKeyConstraints(constraints map[string]*KeyConstraint) error {
	for name, kc := range constraints {
		if kc.Constraint != "" && !kc.ConstraintRegex.MatchString(kc.Constraint) {
			e := name + " value is not valid"
			return errors.New("Key constraint error, " + e)
		}
	}
	return nil
}
