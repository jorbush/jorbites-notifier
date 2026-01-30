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
	models.TypeForgotPassword: `
    	<h2>Password Reset</h2>
  		<p>Hi there,</p>
  		<p>You have requested to reset your password. Click on the following link to create a new password:</p>
  		<a href="{{.Metadata.resetUrl}}" class="button">Reset Password</a>
    	<p>This link will expire in 1 hour.</p>
        <p>If you did not request this change, you can ignore this email.</p>
    `,
	models.TypeMentionInComment: `
    	<h2>You were mentioned in a comment!</h2>
  		<p>Hi there,</p>
    	<p><strong>{{.Metadata.authorName}}</strong> mentioned you in a comment on a recipe.</p>
    	<p>Click the button below to view the recipe:</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button">View Recipe</a>
    `,
	models.TypeNewBlog: `
		<h2>New Blog Post! 📝</h2>
		<p>Hi there,</p>
		<p>A new blog post has been published on Jorbites!</p>
		<a href="{{.SiteURL}}/blog/{{.Metadata.blog_id}}" class="button">Read it now</a>
	`,
}

var emailSubjects = map[models.NotificationType]string{
	models.TypeNewComment:             "New Comment on Your Recipe - Jorbites",
	models.TypeNewLike:                "New Like on Your Recipe - Jorbites",
	models.TypeNewRecipe:              "New Recipe Available - Jorbites",
	models.TypeNotificationsActivated: "Welcome to Jorbites Notifications",
	models.TypeForgotPassword:         "Password Reset Request - Jorbites",
	models.TypeMentionInComment:       "You Were Mentioned in a Comment - Jorbites",
	models.TypeNewBlog:                "New Blog Post Available - Jorbites",
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
