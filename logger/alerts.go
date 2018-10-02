package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"os"
)

var Alerts = parseConfig()

type AlertConfig struct {
	SlackWebhookURL string
	SMTPServer      string
	SMTPPort        int
	MailFrom        string
	Password        string
	MailTo          string
}

func (ac *AlertConfig) SendAlerts(message string) {
	if ac.SlackWebhookURL != "" {
		ac.slackAlert(message)
	}
	if ac.SMTPServer != "" && ac.MailFrom != "" && ac.MailTo != "" && ac.Password != "" {
		ac.emailAlert(message)
	}
}

// slackAlert sends an alert to Slack via web hook
func (ac *AlertConfig) slackAlert(message string) {
	text := map[string]string{
		"text": message,
	}
	postData, err := json.Marshal(text)
	if err != nil {
		return
	}

	_, err = http.Post(ac.SlackWebhookURL, "application/json", bytes.NewBuffer(postData))
	if err != nil {
		fmt.Println("[!] Error sending Slack hook")
	}
}

// emailAlert sends an alert to an email address
func (ac *AlertConfig) emailAlert(message string) {
	subject := "KeyServer Alert"
	email := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s",
		ac.MailFrom, ac.MailTo, subject, message)

	server := fmt.Sprintf("%s:%s", ac.SMTPServer, ac.SMTPPort)
	err := smtp.SendMail(server,
		smtp.PlainAuth("", ac.MailFrom, ac.Password, ac.SMTPServer),
		ac.MailFrom, []string{ac.MailTo}, []byte(email))

	if err != nil {
		return
	}
}

func parseConfig() *AlertConfig {
	configPath := "alerts.config"

	file, err := os.Open(configPath)
	if err != nil {
		fmt.Println("[!] Error reading alerts.config")
		return &AlertConfig{}
	}

	decoder := json.NewDecoder(file)
	config := AlertConfig{}
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("[!] Error parsing config.json:", err)
		return &AlertConfig{}
	}
	return &config
}

func readFile(path string) ([]byte, error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}
