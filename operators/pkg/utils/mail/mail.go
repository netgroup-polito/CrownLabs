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

package mail

import (
	"embed"
	"fmt"
	"net/smtp"
	"reflect"
	"regexp"

	"gopkg.in/yaml.v3"
)

// MailClient is a simple SMTP client for sending emails.
type MailClient struct {
	SMTPServer string
	SMTPPort   int
	Auth       smtp.Auth
	From       string
}

const (
	CROWNLABS_MAIL_TEMPLATE_PATH string = "templates/defaults/crownlabs_mail.eml"
	HEADER_FOOTER_TEMPLATE_PATH  string = "templates/defaults/crownlabs_headers.yaml"
)

// Placeholders is a struct that holds the placeholders values for the email content.
// Each field corresponds to a placeholder in the email template.
// The `name` tag specifies the placeholder name to be used in the template.
// e.g. `{tenantName}` will be replaced with the value of `TenantName` field.
type Placeholders struct {
	TenantName   string `name:"tenantName"`
	TenantEmail  string `name:"tenantEmail"`
	PrettyName   string `name:"prettyName"`
	InstanceName string `name:"instanceName"`
}

var (
	//go:embed templates/*
	templatesFS embed.FS
)

// getPlaceholderKeys returns a map of placeholder where the key
// is the placeholder name and the value is the placeholder value.
func getPlaceholderMap(ph Placeholders) map[string]string {
	t := reflect.TypeOf(ph)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	fieldMap := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := field.Tag.Get("name")
		if name == "" {
			name = field.Name
		}
		value := reflect.ValueOf(ph).FieldByName(field.Name).String()
		fieldMap[name] = value
	}
	return fieldMap
}

// replaceTemplateVars replaces template variables in content using a map of replacements.
func replacePlaceholders(content string, emailValues map[string]string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("content cannot be empty")
	}

	// Replace each email value in the content with its corresponding value
	for key, val := range emailValues {
		// Create regex pattern that matches both {key} and { key } formats
		pattern := regexp.MustCompile(`\{\s*` + regexp.QuoteMeta(key) + `\s*\}`)
		content = pattern.ReplaceAllString(content, val)
	}

	return content, nil
}

// SendMail sends an email using the SMTP server configured in the MailClient.
func (m *MailClient) SendCrownLabsMail(email_content_template_path string, ph Placeholders) error {
	if email_content_template_path == "" {
		return fmt.Errorf("email content template path is required")
	}

	emailContent, err := m.processEmailContentTemplate(email_content_template_path, ph)
	if err != nil {
		return err
	}

	formattedEmail, err := m.prepareFinalEmail(emailContent)
	if err != nil {
		return err
	}

	return m.sendEmail(ph.TenantEmail, formattedEmail)
}

// processEmailContentTemplate loads and processes the content template file
func (m *MailClient) processEmailContentTemplate(templatePath string, ph Placeholders) (map[string]string, error) {
	// Get the email content template
	emailContentTemplate, err := templatesFS.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read email content template: %w", err)
	}

	// Parse content template YAML to extract content fields
	var contentYAML map[string]string
	if err := yaml.Unmarshal(emailContentTemplate, &contentYAML); err != nil {
		return nil, fmt.Errorf("failed to parse email content template: %w", err)
	}

	// Convert placeholders struct to a map
	phMap := getPlaceholderMap(ph)

	// Substitute placeholders in the email content template
	formattedContent, err := replacePlaceholders(string(emailContentTemplate), phMap)
	if err != nil {
		return nil, fmt.Errorf("failed to format email content template: %w", err)
	}

	// Parse the formatted content YAML to extract email fields
	var emailValues map[string]string
	if err := yaml.Unmarshal([]byte(formattedContent), &emailValues); err != nil {
		return nil, fmt.Errorf("failed to parse email content template: %w", err)
	}

	return emailValues, nil
}

// prepareFinalEmail prepares the final email by combining the base template with content
func (m *MailClient) prepareFinalEmail(emailContent map[string]string) (string, error) {
	// Get the entire email template
	crownlabsEmailTemplate, err := templatesFS.ReadFile(CROWNLABS_MAIL_TEMPLATE_PATH)
	if err != nil {
		return "", fmt.Errorf("failed to read email template: %w", err)
	}

	// Get headers template
	headerFooterTemplate, err := templatesFS.ReadFile(HEADER_FOOTER_TEMPLATE_PATH)
	if err != nil {
		return "", fmt.Errorf("failed to read header/footer template: %w", err)
	}

	// Parse headers template
	var headerFooter map[string]string
	if err := yaml.Unmarshal(headerFooterTemplate, &headerFooter); err != nil {
		return "", fmt.Errorf("failed to parse header/footer template: %w", err)
	}

	// Add headers and footers to the email content map
	for key, value := range headerFooter {
		emailContent[key] = value
	}

	// Add sender info to email content
	emailContent["from"] = m.From

	// Substitute placeholders with formatted content
	formattedEmail, err := replacePlaceholders(string(crownlabsEmailTemplate), emailContent)
	if err != nil {
		return "", fmt.Errorf("failed to format email template: %w", err)
	}

	return formattedEmail, nil
}

// sendEmail sends the email to the recipient
func (m *MailClient) sendEmail(recipient string, emailContent string) error {
	msg := []byte(emailContent)
	address := fmt.Sprintf("%s:%d", m.SMTPServer, m.SMTPPort)
	to := []string{recipient}

	return smtp.SendMail(address, m.Auth, m.From, to, msg)
}
