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
</head>
<body style="font-family: 'Nunito', Arial, sans-serif; line-height: 1.6; margin: 0; padding: 0; background-color: #f8f9fa; color: #2d3748;">
    <div class="container" style="max-width: 600px; margin: 40px auto; background-color: #ffffff; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); padding: 40px;">
        <div class="header" style="text-align: center; padding-bottom: 30px; border-bottom: 1px solid #edf2f7; margin-bottom: 30px;">
            <img src="{{.LogoURL}}" alt="Jorbites Logo" class="logo" style="max-width: 140px; height: auto;">
        </div>
        <div class="content" style="padding: 0; color: #4a5568; font-size: 16px;">
            {{.Content}}
        </div>
        <div class="footer" style="text-align: center; padding-top: 30px; color: #a0aec0; font-size: 12px; border-top: 1px solid #edf2f7; margin-top: 40px;">
            <p style="margin-bottom: 1.5em;">You're receiving this email because you have notifications enabled on Jorbites.</p>
            <p style="margin-bottom: 1.5em;">To manage your email preferences, go to <a href="{{.SiteURL}}" style="color: #718096; text-decoration: underline;">Settings → Email Notifications</a></p>
            <p style="margin-bottom: 1.5em;">© {{.CurrentYear}} Jorbites. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

var templateContent = map[models.NotificationType]string{
	models.TypeNewComment: `
        <h2 style="color: #1a202c; font-weight: 700; margin-top: 0;">You have a new comment!</h2>
        <p style="margin-bottom: 1.5em;">Hi there,</p>
        <p style="margin-bottom: 1.5em;"><strong>{{.Metadata.authorName}}</strong> has left a comment on your recipe.</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button" style="display: inline-block; padding: 12px 24px; background-color: #C5F0A4; color: #1a202c !important; text-decoration: none; border-radius: 6px; font-weight: 700; margin: 20px 0; transition: background-color 0.2s; box-shadow: 0 2px 4px rgba(0,0,0,0.05);">View Comment</a>
    `,
	models.TypeNewLike: `
        <h2 style="color: #1a202c; font-weight: 700; margin-top: 0;">Someone liked your recipe!</h2>
        <p style="margin-bottom: 1.5em;">Hi there,</p>
        <p style="margin-bottom: 1.5em;"><strong>{{.Metadata.likedBy}}</strong> has liked your recipe.</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button" style="display: inline-block; padding: 12px 24px; background-color: #C5F0A4; color: #1a202c !important; text-decoration: none; border-radius: 6px; font-weight: 700; margin: 20px 0; transition: background-color 0.2s; box-shadow: 0 2px 4px rgba(0,0,0,0.05);">View Recipe</a>
    `,
	models.TypeNewRecipe: `
        <h2 style="color: #1a202c; font-weight: 700; margin-top: 0;">New Recipe Alert! 🍳</h2>
        <p style="margin-bottom: 1.5em;">Hi there,</p>
        <p style="margin-bottom: 1.5em;">A new recipe has been posted on Jorbites!</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button" style="display: inline-block; padding: 12px 24px; background-color: #C5F0A4; color: #1a202c !important; text-decoration: none; border-radius: 6px; font-weight: 700; margin: 20px 0; transition: background-color 0.2s; box-shadow: 0 2px 4px rgba(0,0,0,0.05);">Check it out</a>
    `,
	models.TypeNotificationsActivated: `
        <h2 style="color: #1a202c; font-weight: 700; margin-top: 0;">Notifications Activated! 🎉</h2>
        <p style="margin-bottom: 1.5em;">Hi there,</p>
        <p style="margin-bottom: 1.5em;">You've successfully activated email notifications for Jorbites.</p>
        <p style="margin-bottom: 1.5em;">You'll now receive updates about:</p>
        <ul>
            <li>New comments on your recipes</li>
            <li>Likes on your recipes</li>
            <li>New recipes from your favorite chefs</li>
        </ul>
    `,
	models.TypeForgotPassword: `
    	<h2 style="color: #1a202c; font-weight: 700; margin-top: 0;">Password Reset</h2>
  		<p style="margin-bottom: 1.5em;">Hi there,</p>
  		<p style="margin-bottom: 1.5em;">You have requested to reset your password. Click on the following link to create a new password:</p>
  		<a href="{{.Metadata.resetUrl}}" class="button" style="display: inline-block; padding: 12px 24px; background-color: #C5F0A4; color: #1a202c !important; text-decoration: none; border-radius: 6px; font-weight: 700; margin: 20px 0; transition: background-color 0.2s; box-shadow: 0 2px 4px rgba(0,0,0,0.05);">Reset Password</a>
    	<p style="margin-bottom: 1.5em;">This link will expire in 1 hour.</p>
        <p style="margin-bottom: 1.5em;">If you did not request this change, you can ignore this email.</p>
    `,
	models.TypeMentionInComment: `
    	<h2 style="color: #1a202c; font-weight: 700; margin-top: 0;">You were mentioned in a comment!</h2>
  		<p style="margin-bottom: 1.5em;">Hi there,</p>
    	<p style="margin-bottom: 1.5em;"><strong>{{.Metadata.authorName}}</strong> mentioned you in a comment on a recipe.</p>
    	<p style="margin-bottom: 1.5em;">Click the button below to view the recipe:</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button" style="display: inline-block; padding: 12px 24px; background-color: #C5F0A4; color: #1a202c !important; text-decoration: none; border-radius: 6px; font-weight: 700; margin: 20px 0; transition: background-color 0.2s; box-shadow: 0 2px 4px rgba(0,0,0,0.05);">View Recipe</a>
    `,
	models.TypeNewBlog: `
		<h2 style="color: #1a202c; font-weight: 700; margin-top: 0;">New Blog Post! 📝</h2>
		<p style="margin-bottom: 1.5em;">Hi there,</p>
		<p style="margin-bottom: 1.5em;">A new blog post has been published on Jorbites!</p>
		<a href="{{.SiteURL}}/blog/{{.Metadata.blog_id}}" class="button" style="display: inline-block; padding: 12px 24px; background-color: #C5F0A4; color: #1a202c !important; text-decoration: none; border-radius: 6px; font-weight: 700; margin: 20px 0; transition: background-color 0.2s; box-shadow: 0 2px 4px rgba(0,0,0,0.05);">Read it now</a>
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
