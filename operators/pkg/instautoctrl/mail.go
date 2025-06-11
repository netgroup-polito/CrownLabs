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

package instautoctrl

import (
	"fmt"
	"net/smtp"
)

// MailClient is a simple SMTP client for sending emails.
type MailClient struct {
	SMTPServer string
	SMTPPort   int
	Auth       smtp.Auth
	From       string
}

// SendMail sends an email using the SMTP server configured in the MailClient.
func (m *MailClient) SendMail(to []string, subject, body string) error {
	msg := []byte(fmt.Sprintf("Subject: %s\n\n%s", subject, body))
	address := fmt.Sprintf("%s:%d", m.SMTPServer, m.SMTPPort)
	return smtp.SendMail(address, m.Auth, m.From, to, msg)
}
