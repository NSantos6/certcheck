package notify

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
	To       string
}

type AlertItem struct {
	Domain  string
	Kind    string // "SSL" or "Domain Registration"
	Status  string // "EXPIRING SOON" or "EXPIRED"
	Expiry  time.Time
	DaysLeft int
}

func SendAlert(cfg SMTPConfig, items []AlertItem) error {
	if len(items) == 0 {
		return nil
	}

	from := cfg.From
	if from == "" {
		from = cfg.User
	}

	subject := fmt.Sprintf("[certcheck] %d domínio(s) precisam de atenção", len(items))
	body := buildBody(items)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, cfg.To, subject, body)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)

	return smtp.SendMail(addr, auth, from, []string{cfg.To}, []byte(msg))
}

func buildBody(items []AlertItem) string {
	var sb strings.Builder

	sb.WriteString("certcheck — relatório de domínios\n")
	sb.WriteString(fmt.Sprintf("Gerado em: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(strings.Repeat("-", 50) + "\n\n")

	for _, item := range items {
		if item.DaysLeft < 0 {
			sb.WriteString(fmt.Sprintf("❌ %s [%s]\n", item.Domain, item.Kind))
			sb.WriteString(fmt.Sprintf("   Status: EXPIRADO há %d dia(s)\n\n", -item.DaysLeft))
		} else {
			sb.WriteString(fmt.Sprintf("⚠️  %s [%s]\n", item.Domain, item.Kind))
			sb.WriteString(fmt.Sprintf("   Status: expira em %d dia(s) (%s)\n\n", item.DaysLeft, item.Expiry.Format("2006-01-02")))
		}
	}

	sb.WriteString(strings.Repeat("-", 50) + "\n")
	sb.WriteString("Enviado por certcheck — https://github.com/NSantos6/certcheck\n")

	return sb.String()
}
