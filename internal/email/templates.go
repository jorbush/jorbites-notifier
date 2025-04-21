package email

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/jorbush/jorbites-notifier/internal/models"
)

const BaseTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Jorbites Notification</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 0;
            background-color: #f4f4f4;
        }
        .container {
            max-width: 600px;
            margin: 20px auto;
            background-color: #ffffff;
            border-radius: 10px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
            padding: 20px;
        }
        .header {
            text-align: center;
            padding: 20px 0;
        }
        .logo {
            max-width: 150px;
            height: auto;
        }
        .content {
            padding: 20px;
            color: #333333;
        }
        .footer {
            text-align: center;
            padding: 20px;
            color: #666666;
            font-size: 12px;
            border-top: 1px solid #eeeeee;
            margin-top: 20px;
        }
        .button {
            display: inline-block;
            padding: 10px 20px;
            background-color: #4CAF50;
            color: white;
            text-decoration: none;
            border-radius: 5px;
            margin: 20px 0;
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
            <p>You're receiving this email because you have notifications enabled on Jorbites.</p>
            <p>To manage your email preferences, go to <a href="{{.SiteURL}}">Settings → Email Notifications</a></p>
            <p>© {{.CurrentYear}} Jorbites. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

var templateContent = map[models.NotificationType]string{
	models.TypeNewComment: `
        <h2>You have a new comment!</h2>
        <p>Hi there,</p>
        <p><strong>{{.Metadata.authorName}}</strong> has left a comment on your recipe.</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button">View Comment</a>
    `,
	models.TypeNewLike: `
        <h2>Someone liked your recipe!</h2>
        <p>Hi there,</p>
        <p><strong>{{.Metadata.likedBy}}</strong> has liked your recipe.</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button">View Recipe</a>
    `,
	models.TypeNewRecipe: `
        <h2>New Recipe Alert! 🍳</h2>
        <p>Hi there,</p>
        <p>A new recipe has been posted on Jorbites!</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button">Check it out</a>
    `,
	models.TypeNotificationsActivated: `
        <h2>Notifications Activated! 🎉</h2>
        <p>Hi there,</p>
        <p>You've successfully activated email notifications for Jorbites.</p>
        <p>You'll now receive updates about:</p>
        <ul>
            <li>New comments on your recipes</li>
            <li>Likes on your recipes</li>
            <li>New recipes from your favorite chefs</li>
        </ul>
    `,
}

var emailSubjects = map[models.NotificationType]string{
	models.TypeNewComment:             "New Comment on Your Recipe - Jorbites",
	models.TypeNewLike:                "New Like on Your Recipe - Jorbites",
	models.TypeNewRecipe:              "New Recipe Available - Jorbites",
	models.TypeNotificationsActivated: "Welcome to Jorbites Notifications",
}

type TemplateData struct {
	SiteURL     string
	LogoURL     string
	CurrentYear int
	Content     string
	Metadata    map[string]string
}

func GetEmailTemplate(notificationType models.NotificationType, metadata map[string]string) (string, string, error) {
	const siteURL = "https://jorbites.com"
	logoURL := siteURL + "/images/logo-nobg.webp"

	contentTemplate, exists := templateContent[notificationType]
	if !exists {
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

	baseTmpl, err := template.New("base").Parse(BaseTemplate)
	if err != nil {
		return "", "", err
	}

	var htmlBuf bytes.Buffer
	if err := baseTmpl.Execute(&htmlBuf, data); err != nil {
		return "", "", err
	}

	subject, exists := emailSubjects[notificationType]
	if !exists {
		subject = "Notification from Jorbites"
	}

	return subject, htmlBuf.String(), nil
}
