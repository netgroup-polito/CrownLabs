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

// Package mail provides utilities for sending templated emails via SMTP.
// It supports loading configuration and templates from Kubernetes ConfigMaps,
// formatting email content with dynamic placeholders, and sending messages
// using standard SMTP authentication.
package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/smtp"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"

	ctrl "sigs.k8s.io/controller-runtime"

	"gopkg.in/yaml.v3"
)

// Client is a simple SMTP client for sending emails.
type Client struct {
	SMTPServer  string
	SMTPPort    int
	Auth        smtp.Auth
	From        string
	TemplateDir string
}

const (
	// CrownlabsMailTemplatePath is the default path for the CrownLabs email template.
	CrownlabsMailTemplatePath string = "defaults_crownlabs_mail.eml"
	// HeaderFooterTemplatePath is the default path for the header/footer template.
	HeaderFooterTemplatePath string = "defaults_crownlabs_headers.yaml"
)

// Placeholders is a struct that holds the placeholders values for the email content.
// Each field corresponds to a placeholder in the email template.
// The `name` tag specifies the placeholder name to be used in the template.
// e.g. `{tenantName}` will be replaced with the value of `TenantName` field.
type Placeholders struct {
	Date         string `name:"date"`
	TenantName   string `name:"tenantName"`
	TenantEmail  string `name:"tenantEmail"`
	PrettyName   string `name:"prettyName"`
	InstanceName string `name:"instanceName"`
}

// NewMailClientFromFilesystem creates a new Client instance that reads configs and templates from filesystem paths.
func NewMailClientFromFilesystem(configDir, templateDir string) (*Client, error) {
	// Load SMTP configuration from filesystem
	configPath := filepath.Join(configDir, "smtp-config.yaml")

	configFile, err := os.Open(filepath.Clean(configPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open SMTP config file %s: %w", configPath, err)
	}
	defer configFile.Close()

	configData, err := io.ReadAll(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read SMTP config file: %w", err)
	}

	var smtpConfigData map[string]string
	if err := yaml.Unmarshal(configData, &smtpConfigData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal smtp-config.yaml: %w", err)
	}

	smtpServer := smtpConfigData["smtpServer"]
	smtpPortStr := smtpConfigData["smtpPort"]
	smtpIdentity := smtpConfigData["smtpIdentity"]
	smtpUsername := smtpConfigData["smtpUsername"]
	smtpPassword := smtpConfigData["smtpPassword"]
	smtpFrom := smtpConfigData["smtpFrom"]

	if smtpServer == "" || smtpPortStr == "" ||
		smtpUsername == "" || smtpPassword == "" || smtpFrom == "" {
		return nil, fmt.Errorf("one or more required SMTP configuration parameters are missing")
	}

	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid smtpPort value '%s': %w", smtpPortStr, err)
	}

	auth := smtp.PlainAuth(smtpIdentity, smtpUsername, smtpPassword, smtpServer)
	return &Client{
		SMTPServer:  smtpServer,
		SMTPPort:    smtpPort,
		Auth:        auth,
		From:        smtpFrom,
		TemplateDir: templateDir,
	}, nil
}

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

// SendCrownLabsMail sends an email using the SMTP server configured in the Client.
func (m *Client) SendCrownLabsMail(emailContentTemplatePath string, ph Placeholders) error {
	log := ctrl.LoggerFrom(context.Background())

	if emailContentTemplatePath == "" {
		return fmt.Errorf("email content template path is required")
	}

	emailContent, err := m.processEmailContentTemplate(emailContentTemplatePath, ph)
	if err != nil {
		log.Error(err, "Failed to process email content template")
		return err
	}

	formattedEmail, err := m.PrepareFinalEmail(emailContent)
	if err != nil {
		log.Error(err, "Failed to prepare final email")
		return err
	}

	return m.sendEmail(ph.TenantEmail, formattedEmail)
}

// readTemplateFile reads a template file from the filesystem.
func (m *Client) readTemplateFile(templatePath string) ([]byte, error) {
	fullPath := filepath.Join(m.TemplateDir, templatePath)

	file, err := os.Open(filepath.Clean(fullPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open template file %s: %w", fullPath, err)
	}
	defer file.Close()

	return io.ReadAll(file)
}

// processEmailContentTemplate loads and processes the content template file.
func (m *Client) processEmailContentTemplate(templatePath string, ph Placeholders) (map[string]string, error) {
	// Get the email content template
	emailContentTemplate, err := m.readTemplateFile(templatePath)
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

// PrepareFinalEmail prepares the final email by combining the base template with content.
func (m *Client) PrepareFinalEmail(emailContent map[string]string) (string, error) {
	// Get the entire email template
	crownlabsEmailTemplate, err := m.readTemplateFile(CrownlabsMailTemplatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read email template: %w", err)
	}

	// Get headers template
	headerFooterTemplate, err := m.readTemplateFile(HeaderFooterTemplatePath)
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

// sendEmail sends the email to the recipient using SSL/TLS connection.
func (m *Client) sendEmail(recipient, emailContent string) error {
	msg := []byte(emailContent)
	address := fmt.Sprintf("%s:%d", m.SMTPServer, m.SMTPPort)
	to := []string{recipient}

	tlsConfig := &tls.Config{
		ServerName: m.SMTPServer,
	}
	conn, err := tls.Dial("tcp", address, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to establish TLS connection: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.SMTPServer)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer func() {
		client.Quit()
		client.Close()
	}()

	if err := client.Auth(m.Auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	if err := client.Mail(m.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", addr, err)
		}
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	if _, err := writer.Write(msg); err != nil {
		writer.Close()
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return nil
}
