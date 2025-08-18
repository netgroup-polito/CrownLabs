# Crownlabs Email Notifications
Crownlabs includes a mail notification system that allows operators to send emails when specific events are triggered.

The system is template-based: email templates are organized in the samples/email directory and categorized by operator (e.g., instance-automation, tenant-operator, etc.). A template defines the following components:

* Subject of the email
* Recipient of the message
* Body of the email, which can be written in plain text or HTML format for richer content

In addition to the individual templates, header and footer sections have been designed and are stored in `samples/email/defaults`. These provide general Crownlabs branding and are reused across all messages to maintain consistency.

The logic for email generation and delivery is implemented in the `pkg/utils/mail/` package, which contains helper functions for:
* Loading templates and defaults
* Populating them with data
* Assembling the final email
* Sending it to the specified recipient

## Email Functions
- **NewMailClientFromFilesystem**: Creates a new `Client` instance by reading SMTP configuration and templates from a filesystem directory (typically a mounted ConfigMap volume). It expects a YAML file named `smtp-config.yaml` to be present in the template directory.

- **getPlaceholderMap**: Internal utility that converts a `Placeholders` struct into a map of placeholder names and values. It uses struct tags to determine placeholder keys.

- **replacePlaceholders**: Replaces all placeholder variables (e.g., `{tenantName}`) in the email content with their actual values.

- **SendCrownLabsMail**: Sends an email using the provided content template and placeholder values. It loads the base template and optional headers/footers, substitutes placeholders, and sends the email.

- **readTemplateFile**: Reads a template file from the filesystem.

- **processEmailContentTemplate**: Parses the email content template (in YAML format), replaces all placeholders, and returns a map of email fields (e.g., subject, body).

- **prepareFinalEmail**: Merges header/footer templates with the email content and base email template. It substitutes placeholders to generate the final formatted email.

- **sendEmail**: Sends the fully formatted email to the specified recipient using the configured SMTP client.