package email

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/jorbush/jorbites-notifier/internal/i18n"
	"github.com/jorbush/jorbites-notifier/internal/models"
)

const BaseTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Jorbites Notification</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Nunito:wght@400;600;700&display=swap" rel="stylesheet">
    <style>
        body {
            font-family: 'Nunito', Arial, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 0;
            background-color: #f8f9fa; /* Lighter background */
            color: #2d3748; /* Dark gray for better readability */
        }
        .container {
            max-width: 600px;
            margin: 40px auto;
            background-color: #ffffff;
            border-radius: 12px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.05);
            padding: 40px;
        }
        .header {
            text-align: center;
            padding-bottom: 30px;
            border-bottom: 1px solid #edf2f7;
            margin-bottom: 30px;
        }
        .logo {
            max-width: 140px;
            height: auto;
        }
        .content {
            padding: 0;
            color: #4a5568;
            font-size: 16px;
        }
        h2 {
            color: #1a202c;
            font-weight: 700;
            margin-top: 0;
        }
        p {
            margin-bottom: 1.5em;
        }
        .footer {
            text-align: center;
            padding-top: 30px;
            color: #a0aec0;
            font-size: 12px;
            border-top: 1px solid #edf2f7;
            margin-top: 40px;
        }
        .footer a {
            color: #718096;
            text-decoration: underline;
        }
        .button {
            display: inline-block;
            padding: 12px 24px;
            background-color: #C5F0A4; /* App main color */
            color: #1a202c !important; /* Dark text for contrast */
            text-decoration: none;
            border-radius: 6px;
            font-weight: 700;
            margin: 20px 0;
            transition: background-color 0.2s;
            box-shadow: 0 2px 4px rgba(0,0,0,0.05);
        }
        .button:hover {
            background-color: #b2d892; /* Slightly darker for hover */
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <img src="{{.LogoURL}}" alt="Jorbites Logo" class="logo">
        </div>
        <div class="content">
            {{.Content}}
        </div>
        <div class="footer">
            {{.Footer}}
        </div>
    </div>
</body>
</html>
`

type TemplateData struct {
	SiteURL     string
	LogoURL     string
	CurrentYear int
	Content     string
	Footer      string
	Metadata    map[string]string
}

func GetEmailTemplate(notificationType models.NotificationType, metadata map[string]string, language string) (string, string, error) {
	const siteURL = "https://jorbites.com"
	logoURL := siteURL + "/images/logo-nobg.webp"

	contentTemplate := i18n.GetEmailTemplateContent(notificationType, language)
	if contentTemplate == "" {
		return "", "", fmt.Errorf("no template defined for notification type: %s", notificationType)
	}

	contentTmpl, err := template.New("content").Parse(contentTemplate)
	if err != nil {
		return "", "", err
	}

	data := TemplateData{
		SiteURL:     siteURL,
		LogoURL:     logoURL,
		CurrentYear: time.Now().Year(),
		Metadata:    metadata,
	}

	var contentBuf bytes.Buffer
	if err := contentTmpl.Execute(&contentBuf, data); err != nil {
		return "", "", err
	}

	data.Content = contentBuf.String()

	footerTemplate := i18n.GetBaseTemplateFooter(language)
	footerTmpl, err := template.New("footer").Parse(footerTemplate)
	if err != nil {
		return "", "", err
	}

	var footerBuf bytes.Buffer
	if err := footerTmpl.Execute(&footerBuf, data); err != nil {
		return "", "", err
	}

	data.Footer = footerBuf.String()

	baseTmpl, err := template.New("base").Parse(BaseTemplate)
	if err != nil {
		return "", "", err
	}

	var htmlBuf bytes.Buffer
	if err := baseTmpl.Execute(&htmlBuf, data); err != nil {
		return "", "", err
	}

	subject := i18n.GetEmailSubject(notificationType, language)

	return subject, htmlBuf.String(), nil
}
