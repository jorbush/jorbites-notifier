package push

import (
	"testing"
)

func TestValidatePushEndpoint(t *testing.T) {
	validEndpoints := []string{
		"https://fcm.googleapis.com/fcm/send/abc123",
		"https://updates.push.services.mozilla.com/wpush/v2/token",
		"https://web.push.apple.com/QABCD123",
		"https://wns2-bl2p.notify.windows.com/w/?token=abc",
	}

	for _, ep := range validEndpoints {
		if err := ValidatePushEndpoint(ep); err != nil {
			t.Errorf("expected valid endpoint %q, got error: %v", ep, err)
		}
	}

	invalidEndpoints := []string{
		"",
		"http://fcm.googleapis.com/fcm/send/insecure",
		"http://169.254.169.254/latest/meta-data/",
		"https://169.254.169.254/latest/meta-data/",
		"http://localhost:8080/internal",
		"https://service.internal/admin",
		"https://attacker.example.com/hook",
		"https://user:pass@fcm.googleapis.com/fcm/send/abc",
		"https://fcm.googleapis.com:8443/fcm/send/abc",
		"https://not-fcm.googleapis.com/evil",
	}

	for _, ep := range invalidEndpoints {
		if err := ValidatePushEndpoint(ep); err == nil {
			t.Errorf("expected invalid endpoint %q to be rejected, got nil", ep)
		}
	}
}
