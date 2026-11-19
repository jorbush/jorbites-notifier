package email

import (
	"strings"
	"testing"

	"github.com/jorbush/jorbites-notifier/internal/models"
)

func TestGetEmailTemplateEscapesHTMLInjection(t *testing.T) {
	maliciousAuthor := `<script>alert("xss")</script><a href="https://evil.example">Click here</a>`
	_, htmlBody, err := GetEmailTemplate(
		models.TypeNewComment,
		map[string]string{
			"authorName": maliciousAuthor,
			"recipeId":   "recipe-123",
		},
		"en",
	)
	if err != nil {
		t.Fatalf("GetEmailTemplate returned unexpected error: %v", err)
	}

	if strings.Contains(htmlBody, "<script>alert") {
		t.Fatalf("expected script tag in authorName to be HTML-escaped, got: %s", htmlBody)
	}
	if strings.Contains(htmlBody, `<a href="https://evil.example">`) {
		t.Fatalf("expected injected anchor tag in authorName to be HTML-escaped, got: %s", htmlBody)
	}
	if !strings.Contains(htmlBody, "&lt;script&gt;") {
		t.Fatalf("expected escaped script tag in output, got: %s", htmlBody)
	}
	// Verify legitimate structural HTML from the trusted template is preserved
	if !strings.Contains(htmlBody, `class="button"`) {
		t.Fatalf("expected trusted template markup to remain intact, got: %s", htmlBody)
	}
}

func TestGetEmailTemplateForgotPasswordPreservesValidResetURL(t *testing.T) {
	resetURL := "https://jorbites.com/reset-password/abcdef123456"
	_, htmlBody, err := GetEmailTemplate(
		models.TypeForgotPassword,
		map[string]string{
			"resetUrl": resetURL,
		},
		"en",
	)
	if err != nil {
		t.Fatalf("GetEmailTemplate returned unexpected error: %v", err)
	}
	if !strings.Contains(htmlBody, resetURL) {
		t.Fatalf("expected resetUrl to be preserved in forgot-password email, got: %s", htmlBody)
	}
}
