// Copyright 2020-2025 Politecnico di Torino
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"fmt"
	"net/smtp"
	"strings"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// MailClient is a simple SMTP client for sending emails.
type MailClient struct {
	SMTPServer string
	SMTPPort   int
	Auth       smtp.Auth
	From       string
}

// EmailTemplate contains the common parts of an email notification.
type EmailTemplate struct {
	Subject     string
	HeaderHTML  string
	FooterHTML  string
	PlainHeader string
	PlainFooter string
}

// DefaultEmailTemplate returns the standard CrownLabs email template.
func DefaultEmailTemplate() EmailTemplate {
	return EmailTemplate{
		HeaderHTML: `<!DOCTYPE html>
<html>
<body>
<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
    <div style="background-color: #f8f9fa; padding: 20px; text-align: center;">
        <h2>CrownLabs Notification</h2>
    </div>
    <div style="padding: 20px;">`,
		FooterHTML: `
    </div>
    <div style="background-color: #f8f9fa; padding: 15px; font-size: 12px; text-align: center;">
        <p>This is an automated message from CrownLabs.</p>
        <p>If you need assistance, please contact support.</p>
    </div>
</div>
</body>
</html>`,
		PlainHeader: "Dear user,\n\n",
		PlainFooter: "\n\nBest regards,\nCrownLabs Team",
	}
}

// FormatEmailContent replaces template variables in message content.
func FormatEmailContent(content string, instance *clv1alpha2.Instance, tenant *clv1alpha2.Tenant) string {
	replacements := map[string]string{
		"{name}":        instance.Name,
		"{prettyName}":  instance.Spec.PrettyName,
		"{tenantName}":  tenant.Name,
		"{tenantEmail}": tenant.Spec.Email,
	}

	for key, value := range replacements {
		content = strings.ReplaceAll(content, key, value)
	}
	return content
}

// SendMail sends an email using the SMTP server configured in the MailClient.
// Only supports plain text emails.
func (m *MailClient) sendMail(to []string, subject, body string) error {
	msg := []byte(fmt.Sprintf("Subject: %s\n\n%s", subject, body))
	address := fmt.Sprintf("%s:%d", m.SMTPServer, m.SMTPPort)
	return smtp.SendMail(address, m.Auth, m.From, to, msg)
}

// SendHTMLMail sends an email with both plain text and HTML content.
func (m *MailClient) sendHTMLMail(to []string, subject, plainBody, htmlBody string) error {
	boundary := "CrownLabsEmailBoundary"

	msg := []byte(fmt.Sprintf("From: %s\r\n", m.From) +
		fmt.Sprintf("To: %s\r\n", to[0]) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"MIME-Version: 1.0\r\n" +
		fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n\r\n", boundary) +

		fmt.Sprintf("--%s\r\n", boundary) +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		plainBody + "\r\n\r\n" +

		fmt.Sprintf("--%s\r\n", boundary) +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		htmlBody + "\r\n\r\n" +

		fmt.Sprintf("--%s--", boundary))

	address := fmt.Sprintf("%s:%d", m.SMTPServer, m.SMTPPort)
	return smtp.SendMail(address, m.Auth, m.From, to, msg)
}

// SendFormattedMail sends either HTML or plain text email depending on if htmlBody is provided.
func (m *MailClient) SendFormattedMail(to []string, subject, plainBody string, htmlBody ...string) error {
	if len(htmlBody) > 0 && htmlBody[0] != "" {
		return m.sendHTMLMail(to, subject, plainBody, htmlBody[0])
	}
	return m.sendMail(to, subject, plainBody)
}
