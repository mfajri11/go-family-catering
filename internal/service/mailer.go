package service

import (
	"bytes"
	"family-catering/config"
	"family-catering/pkg/logger"
	"fmt"
	"net/smtp"
	"path/filepath"
	"strings"
	"text/template"
)

type Mailer interface {
	SendEMailForgotPassword(to []string, cc, subject, name, authlink string) error
	SendEmailNotifyLogin(to []string, cc, subject, name string) error
}

type mailer struct {
	host                       string
	email                      string
	password                   string
	supportEmail               string
	appName                    string
	templateForgotPasswordName string
	identity                   string
	port                       int
	address                    string
	temp                       *template.Template
}

func NewMailer(opts MailerOption) Mailer {

	m := &mailer{
		appName:                    opts.AppName,
		email:                      opts.Email,
		supportEmail:               opts.SupportEmail,
		host:                       opts.Host,
		templateForgotPasswordName: opts.TemplateForgotPasswordName,
		port:                       opts.Port,
		identity:                   opts.Identity,
	}
	path := filepath.Join(config.Path(), m.templateForgotPasswordName)

	m.temp = template.Must(template.ParseFiles(path))
	m.address = fmt.Sprintf("%s:%d", m.host, m.port)
	// ? does necessary to use tls.Config to give option skipVerify through config?

	return m
}

type MailerOption struct {
	Host                       string
	Email                      string
	Password                   string
	SupportEmail               string
	AppName                    string
	TemplateForgotPasswordName string
	Identity                   string
	Port                       int
}

type forgotPasswordEmailTemplate struct {
	AppName      string
	Sender       string
	To           string
	Cc           string
	Subject      string
	ToName       string
	Link         string
	SupportEmail string
}

func (m *mailer) SendEMailForgotPassword(to []string, cc, subject, name, link string) error {
	tos := strings.Join(to, ",")
	templateBody := forgotPasswordEmailTemplate{
		SupportEmail: m.supportEmail,
		AppName:      m.appName,
		Sender:       m.email,
		To:           tos,
		Cc:           cc,
		ToName:       name,
		Subject:      subject,
		Link:         link,
	}

	body := bytes.Buffer{}
	m.temp.Execute(&body, templateBody)
	emailAuth := smtp.CRAMMD5Auth(m.email, m.password)
	err := smtp.SendMail(m.address, emailAuth, fmt.Sprintf("<%s>", m.email), to, body.Bytes())
	if err != nil {
		err = fmt.Errorf("service.mailer.SendEmailForgotPassword: %w", err)
		logger.Error(err, "error sending email for forgot password")
	}

	return err
}

func (m *mailer) SendEmailNotifyLogin(to []string, cc, subject, name string) error {
	panic("not yet implemented")
}
