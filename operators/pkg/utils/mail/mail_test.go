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

package mail_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/mail"
)

func TestMail(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mail Test Suite")
}

var _ = Describe("Mail", func() {
	var tmpDir string

	var placeholders = mail.Placeholders{
		TenantName:   "Alice",
		TenantEmail:  "alice@example.com",
		PrettyName:   "Instance-1 Pretty",
		InstanceName: "instance-1",
	}

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "mailtest")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	writeSMTPConfig := func(configDir string, content string) {
		Expect(os.MkdirAll(configDir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(configDir, "smtp-config.yaml"), []byte(content), 0o644)).To(Succeed())
	}

	writeTemplates := func(templateDir string) {
		defaultsDir := filepath.Join(templateDir, "defaults")

		Expect(os.MkdirAll(defaultsDir, 0o755)).To(Succeed())

		crownlabsTemplate := "From: {from}\nTo: {tenantEmail}\nSubject: Test\n\n{body}\n{footer}"
		headers := "footer: |\n  --\n  CrownLabs Team"

		Expect(os.WriteFile(filepath.Join(templateDir, "defaults_crownlabs_mail.eml"), []byte(crownlabsTemplate), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(templateDir, "defaults_crownlabs_headers.yaml"), []byte(headers), 0o644)).To(Succeed())
	}

	It("fails if SMTP config file is missing", func() {
		configDir := filepath.Join(tmpDir, "config")
		templateDir := filepath.Join(tmpDir, "templates")
		Expect(os.MkdirAll(templateDir, 0o755)).To(Succeed())

		client, err := mail.NewMailClientFromFilesystem(configDir, templateDir)
		Expect(client).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to open SMTP config file"))
	})

	It("fails if SMTP config has missing fields", func() {
		configDir := filepath.Join(tmpDir, "config")
		templateDir := filepath.Join(tmpDir, "templates")
		writeSMTPConfig(configDir, `
smtpServer: "smtp.example.com"
smtpPort: "587"
smtpUsername: "user"
smtpPassword: "pass"
`)
		Expect(os.MkdirAll(templateDir, 0o755)).To(Succeed())

		client, err := mail.NewMailClientFromFilesystem(configDir, templateDir)
		Expect(client).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("one or more required SMTP configuration parameters are missing"))
	})

	Context("with correct SMTP config and templates", func() {
		var client *mail.Client

		BeforeEach(func() {
			configDir := filepath.Join(tmpDir, "config")
			templateDir := filepath.Join(tmpDir, "templates")

			writeSMTPConfig(configDir, `
smtpServer: "smtp.example.com"
smtpPort: "587"
smtpIdentity: "identity"
smtpUsername: "user"
smtpPassword: "pass"
smtpFrom: "sender@example.com"
`)
			writeTemplates(templateDir)

			var err error
			client, err = mail.NewMailClientFromFilesystem(configDir, templateDir)
			Expect(err).ToNot(HaveOccurred())
		})

		It("creates a client with proper SMTP settings", func() {
			Expect(client.SMTPServer).To(Equal("smtp.example.com"))
			Expect(client.SMTPPort).To(Equal(587))
			Expect(client.From).To(Equal("sender@example.com"))
		})

		It("combines template, headers, and content", func() {

			content := map[string]string{
				"tenantEmail": placeholders.TenantEmail,
				"body":        "Hello world",
			}
			email, err := client.PrepareFinalEmail(content)
			Expect(err).ToNot(HaveOccurred())
			Expect(email).To(ContainSubstring("From: sender@example.com"))
			Expect(email).To(ContainSubstring("Hello world"))
			Expect(email).To(ContainSubstring("CrownLabs Team"))
		})

		It("fails PrepareFinalEmail if template file is missing", func() {
			clientNoTemplate := &mail.Client{
				From:        client.From,
				TemplateDir: filepath.Join(tmpDir, "empty_templates"),
			}
			email, err := clientNoTemplate.PrepareFinalEmail(map[string]string{})
			Expect(email).To(BeEmpty())
			Expect(err).To(HaveOccurred())
		})

		It("handles empty placeholders correctly", func() {
			contentMap := map[string]string{
				"tenantName":   "",
				"tenantEmail":  "",
				"prettyName":   "",
				"instanceName": "",
				"footer":       "--\nCrownLabs Team",
				"body":         "Hello world",
				"from":         client.From,
			}
			email, err := client.PrepareFinalEmail(contentMap)
			Expect(err).ToNot(HaveOccurred())
			Expect(email).To(ContainSubstring("From: sender@example.com"))
			Expect(email).To(ContainSubstring("--\nCrownLabs Team"))
		})

		It("replaces all placeholders correctly in the email", func() {
			contentMap := map[string]string{
				"tenantName":   placeholders.TenantName,
				"tenantEmail":  placeholders.TenantEmail,
				"prettyName":   placeholders.PrettyName,
				"instanceName": placeholders.InstanceName,
				"footer":       "--\nCrownLabs Team",
				"body":         "Hello Alice, your instance Instance-1 Pretty (Instance-1) is ready!",
				"from":         client.From,
			}

			email, err := client.PrepareFinalEmail(contentMap)
			Expect(err).ToNot(HaveOccurred())

			// Check that placeholders are correctly replaced
			Expect(email).To(ContainSubstring("From: sender@example.com"))
			Expect(email).To(ContainSubstring("To: " + placeholders.TenantEmail))
			Expect(email).To(ContainSubstring("Hello Alice, your instance Instance-1 Pretty (Instance-1) is ready!"))
			Expect(email).To(ContainSubstring("--\nCrownLabs Team"))
		})
	})
})
