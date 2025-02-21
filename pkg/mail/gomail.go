package mail

import (
	"bytes"

	"gopkg.in/gomail.v2"
)

type Gomail interface {
	SendEmail(request *SendEmail) error
	GetFromEmail() string
}

type ImplGomail struct {
	client    *gomail.Dialer
	fromEmail string
}

type SendEmail struct {
	EmailFrom string
	EmailTo   string
	Subject   string
	Body      bytes.Buffer
}

func NewGomail(client *gomail.Dialer, fromEmail string) *ImplGomail {
	return &ImplGomail{
		client:    client,
		fromEmail: fromEmail,
	}
}

func (g *ImplGomail) SendEmail(request *SendEmail) error {
	m := gomail.NewMessage()
	m.SetHeader("From", g.fromEmail)
	m.SetHeader("To", request.EmailTo)
	m.SetHeader("Subject", request.Subject)
	m.SetBody("text/html", request.Body.String())

	return g.client.DialAndSend(m)
}

func (g *ImplGomail) GetFromEmail() string {
	return g.fromEmail
}
