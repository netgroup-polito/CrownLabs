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

var (
	// HTML CrownLabs email header and footer
	defaultHeaderHTML = `<!DOCTYPE html>
<html>
<body>
<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
    <div style="background-color: #f8f9fa; padding: 20px; text-align: center;">
        <h2>CrownLabs Notification</h2>
    </div>
    <div style="padding: 20px;">
        <p>Dear user,</p>`
	defaultFooterHTML = `
        <p>Best regards,<br>
        CrownLabs Team</p>
    </div>
    <div style="background-color: #f8f9fa; padding: 15px; font-size: 12px; text-align: center;">
        <p>This is an automated message from CrownLabs.</p>
        <p>If you need assistance, please contact support.</p>
    </div>
</div>
</body>
</html>`
	// Plaintext CrownLabs email Header and Footer
	defaultPlainHeader = "=== CROWNLABS NOTIFICATION ===\n\nDear user,\n\n"
	defaultPlainFooter = "\n\nBest regards,\nCrownLabs Team\n\n---\nThis is an automated message from CrownLabs.\nIf you need assistance, please contact support."
)

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
		HeaderHTML:  defaultHeaderHTML,
		FooterHTML:  defaultFooterHTML,
		PlainHeader: defaultPlainHeader,
		PlainFooter: defaultPlainFooter,
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
// If htmlBody is provided, sends a multipart email with both plain text and HTML versions.
// If htmlBody is empty, sends just a plain text email.
func (m *MailClient) SendMail(to []string, subject, plainBody string, htmlBody ...string) error {
	address := fmt.Sprintf("%s:%d", m.SMTPServer, m.SMTPPort)
	// Common headers
	headers := []string{
		fmt.Sprintf("From: %s", m.From),
		fmt.Sprintf("To: %s", strings.Join(to, ", ")),
		fmt.Sprintf("Subject: %s", subject),
	}
	// Send as HTML if htmlBody is provided, otherwise send as plain text
	if len(htmlBody) > 0 && htmlBody[0] != "" {
		// Create MIME multipart email
		boundary := "CrownLabsEmailBoundary"
		mimeHeaders := append(headers,
			"MIME-Version: 1.0",
			fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s", boundary),
			"")
		// Build plain text part
		plainTextPart := []string{
			fmt.Sprintf("--%s", boundary),
			"Content-Type: text/plain; charset=UTF-8",
			"",
			plainBody,
			""}
		// Build HTML part
		htmlPart := []string{
			fmt.Sprintf("--%s", boundary),
			"Content-Type: text/html; charset=UTF-8",
			"",
			htmlBody[0],
			""}
		// Add closing boundary
		closing := []string{
			fmt.Sprintf("--%s--", boundary)}
		// Combine all parts
		message := strings.Join(
			append(
				append(
					append(mimeHeaders, plainTextPart...),
					htmlPart...),
				closing...),
			"\r\n")
		msg := []byte(message)
		return smtp.SendMail(address, m.Auth, m.From, to, msg)
	} else {
		// Plain text email only
		plainHeaders := append(headers,
			"",
			plainBody)
		msg := []byte(strings.Join(plainHeaders, "\r\n"))
		return smtp.SendMail(address, m.Auth, m.From, to, msg)
	}
}
