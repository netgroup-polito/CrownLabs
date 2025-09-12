# Crownlabs Email Notifications

Crownlabs includes a mail notification system that allows operators to send emails when specific events are triggered (e.g., instance expiration, inactivity warnings).

The system is template-based: email templates are organized under `deploy/crownmail/mail-templates` and categorized by operator (e.g., `instance-automation`, `tenant-operator`, …).

Each email template is a YAML file defining:

- subject of the email
- recipient(s)
- body of the message, in either plain text or HTML

In addition to per-event templates, common CrownLabs header and footer fragments as well as MIME mail format are defined under `deploy/crownmail/mail-templates/defaults`.


## Implementation

The logic for email generation and delivery lives in [`mail.go`](mail.go).
This package provides helper functions for:

- Loading SMTP configuration and templates from disk (typically via Kubernetes ConfigMaps and Secrets)
- Parsing YAML templates and substituting placeholders
- Assembling the final email with headers/footers
- Sending it via SMTP with authentication

The main type is:

```go
type Client struct {
    SMTPServer  string
    SMTPPort    int
    Auth        smtp.Auth
    From        string
    TemplateDir string
}
```


## Configuration

- **SMTP settings** are read from a `smtp-config.yaml` file located in the provided config directory (usually mounted from a Secret).
  Required fields:

  ```yaml
  smtpServer: smtp.example.com
  smtpPort: "587"
  smtpIdentity: ""
  smtpUsername: "user@example.com"
  smtpPassword: "supersecret"
  smtpFrom: "noreply@example.com"
  ```

- Email templates, header/footer defaults and MIME mail format are read from the configured `TemplateDir` (usually mounted from a ConfigMap).


## Email Workflow

1. Initialize the client

   ```go
   client, err := NewMailClientFromFilesystem("/etc/crownmail/mail-config", "/etc/crownmail/mail-templates")
   ```

   This loads SMTP config from `/etc/crownmail/mail-configs/smtp-config.yaml` and prepares the template directory.

2. Prepare placeholders

   ```go
   ph := Placeholders{
       TenantName:   "Alice",
       TenantEmail:  "alice@example.com",
       PrettyName:   "Alice B.",
       InstanceName: "lab-instance-1",
   }
   ```

3. Send the email (expiration notification)

   ```go
   err = client.SendCrownLabsMail("instance-automation/expiration_notification.yaml", ph)
   ```


## Functions

- `NewMailClientFromFilesystem(configDir, templateDir)`:
  Creates a `Client` by reading SMTP configuration and templates from the filesystem.

- `getPlaceholderMap(ph Placeholders)`
  Converts a `Placeholders` struct into a map of `{ placeholderName: value }`, based on struct tags.

- `replacePlaceholders(content string, values map[string]string)`
  Replaces all `{placeholder}` occurrences in the input string with actual values.

- `readTemplateFile(templatePath string)`
  Reads a template file from the filesystem under the configured `TemplateDir`.

- `processEmailContentTemplate(templatePath string, ph Placeholders)`
  Loads a YAML content template, substitutes placeholders, and returns a map of fields (`subject`, `body`, etc.).

- `prepareFinalEmail(emailContent map[string]string)`
  Combines base template, header/footer defaults, and substituted content into a final RFC822-style email message.

- `sendEmail(recipient, emailContent string)`
  Sends the fully formatted email through the configured SMTP server.

- `SendCrownLabsMail(templatePath string, ph Placeholders)`
  Main entry point: loads a content template, applies placeholders, merges with base template + headers/footers, and sends via SMTP.
