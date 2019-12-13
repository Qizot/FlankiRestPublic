package services

import (
	"FlankiRest/logger"
	"crypto/tls"
	"github.com/joeshaw/envdecode"
	"gopkg.in/gomail.v2"
	"io"
	"io/ioutil"
)

const SMTPServer  = "smtp.gmail.com"
const SMTPPort    = 587

var appEmailInstance *MailAuth

type EmailSender struct {
	From string
	To []string
	Subject string
	Body io.Reader
}

type MailAuth struct {
	Username string `env:"APP_EMAIL,required"`
	Password string `env:"APP_EMAIL_PASSWORD,required"`
}

func init() {
	appEmailInstance = &MailAuth{}
	err := envdecode.Decode(appEmailInstance)
	if err != nil {
		logger.GetGlobalLogger().WithField("prefix", "[EMAIL SERVICE]").Fatal(err.Error())
	}
}

func GetAppMailAuth() *MailAuth {
	return appEmailInstance
}

func (auth MailAuth) SendEmail(sender EmailSender, useDefaultEmail bool) error {

	m := gomail.NewMessage()
	if useDefaultEmail {
		m.SetHeader("From", auth.Username)
	} else {
		m.SetHeader("From", sender.From)
	}

	addresses := make([]string, len(sender.To))
	for i := range addresses {
		addresses[i] = m.FormatAddress(sender.To[i], "")
	}
	m.SetHeader("To", addresses...)

	m.SetHeader("Subject", sender.Subject)
	b, err := ioutil.ReadAll(sender.Body)
	if err != nil {
		return err
	}
	m.SetBody("text/html", string(b))

	d := gomail.NewDialer(SMTPServer, SMTPPort, auth.Username, auth.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // temporary because it doesnt work on linux
	return d.DialAndSend(m)
}
