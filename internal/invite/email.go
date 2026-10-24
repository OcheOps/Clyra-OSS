package invite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/oche/diagram/internal/config"
)

const resendAPI = "https://api.resend.com/emails"

// EmailService sends invite emails via Resend.
type EmailService struct {
	apiKey    string
	fromEmail string
	fromName  string
	appURL    string
	client    *http.Client
	logger    *slog.Logger
}

func NewEmailService(cfg config.ResendConfig, logger *slog.Logger) *EmailService {
	return &EmailService{
		apiKey:    cfg.APIKey,
		fromEmail: cfg.FromEmail,
		fromName:  cfg.FromName,
		appURL:    cfg.AppURL,
		client:    &http.Client{Timeout: 10 * time.Second},
		logger:    logger,
	}
}

// Enabled returns true if the Resend API key is configured.
func (s *EmailService) Enabled() bool {
	return s.apiKey != ""
}

// SendInvite sends a beta invite email with the invite code.
func (s *EmailService) SendInvite(toEmail, code string) error {
	return s.send(toEmail, "You're invited to Clyra", s.buildInviteHTML(toEmail, code), "invite")
}

// SendWelcome greets a newly-registered user. Best-effort; callers
// typically fire-and-forget so signup latency doesn't depend on Resend.
func (s *EmailService) SendWelcome(toEmail string) error {
	return s.send(toEmail, "Welcome to Clyra", s.buildWelcomeHTML(toEmail), "welcome")
}

// SendPasswordReset sends a password-reset link. The token must be the
// plaintext value that the recipient needs to click — never send the hash.
func (s *EmailService) SendPasswordReset(toEmail, token string) error {
	return s.send(toEmail, "Reset your Clyra password", s.buildPasswordResetHTML(toEmail, token), "password_reset")
}

// SendAdmin sends a one-off admin message to a user. Body is plain text
// with simple newline → <br> conversion; no HTML allowed from admins to
// keep this safe-by-default.
func (s *EmailService) SendAdmin(toEmail, subject, body string) error {
	return s.send(toEmail, subject, s.buildAdminHTML(toEmail, body), "admin")
}

// SendFeedbackNotification pings the operator that someone submitted a
// support / feedback form. The body itself is escaped — we never trust
// user input in HTML.
func (s *EmailService) SendFeedbackNotification(toEmail, fromName, fromEmail, category, message, refID string) error {
	body := fmt.Sprintf(
		"From: %s <%s>\nCategory: %s\nID: %s\n\n%s",
		fromName, fromEmail, category, refID, message,
	)
	subject := "[Clyra feedback · " + category + "] " + refID
	return s.send(toEmail, subject, s.buildAdminHTML(toEmail, body), "feedback")
}

func (s *EmailService) buildAdminHTML(toEmail, body string) string {
	// Light HTML template — stay on-brand without going overboard.
	safeBody := htmlEscape(body)
	// Convert newlines to <br> so authors don't need to write HTML.
	safeBody = strings.ReplaceAll(safeBody, "\n", "<br>")
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="margin:0;padding:0;background-color:#fbfaf7;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;color:#1c1e23;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#fbfaf7;padding:40px 20px;">
    <tr><td align="center">
      <table width="520" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border:1px solid #ece9e1;border-radius:12px;padding:36px;">
        <tr><td style="font-size:20px;font-weight:600;color:#121316;padding-bottom:16px;">clyra.</td></tr>
        <tr><td style="font-size:15px;line-height:1.55;color:#1c1e23;">%s</td></tr>
        <tr><td style="padding-top:28px;font-size:12px;color:#8d93a0;">
          Sent to %s · <a href="%s" style="color:#0f7d79;text-decoration:none;">useclyra.com</a>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, safeBody, htmlEscape(toEmail), s.appURL)
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// send is the shared Resend request path. A logical `kind` label tags the
// log line so we can filter in ops.
func (s *EmailService) send(toEmail, subject, html, kind string) error {
	if !s.Enabled() {
		s.logger.Warn("resend not configured, skipping email", "to", toEmail, "kind", kind)
		return nil
	}

	payload := map[string]any{
		"from":    fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail),
		"to":      []string{toEmail},
		"subject": subject,
		"html":    html,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling email: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, resendAPI, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp map[string]any
		json.NewDecoder(resp.Body).Decode(&errResp)
		s.logger.Error("resend API error", "status", resp.StatusCode, "response", errResp, "kind", kind)
		return fmt.Errorf("resend returned %d", resp.StatusCode)
	}

	s.logger.Info("email sent", "to", toEmail, "kind", kind)
	return nil
}

func (s *EmailService) buildPasswordResetHTML(toEmail, token string) string {
	link := fmt.Sprintf("%s/reset-password?token=%s", s.appURL, token)
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0;padding:0;background-color:#0a0d14;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#0a0d14;padding:40px 20px;">
    <tr>
      <td align="center">
        <table width="480" cellpadding="0" cellspacing="0" style="background-color:#151a24;border:1px solid #252d3d;border-radius:12px;padding:40px;">
          <tr>
            <td align="center" style="padding-bottom:24px;">
              <img src="%s/clyra-logo.png" alt="Clyra" height="36" style="display:block;" />
            </td>
          </tr>
          <tr>
            <td align="center" style="padding-bottom:16px;">
              <h1 style="margin:0;font-size:20px;font-weight:600;color:#e8ecf4;">Reset your password</h1>
            </td>
          </tr>
          <tr>
            <td align="center" style="padding-bottom:32px;">
              <p style="margin:0;font-size:14px;line-height:22px;color:#9ba4b8;">
                Someone asked to reset the Clyra password for this email. If that was you, click below.
                The link expires in 1 hour and can only be used once.
              </p>
            </td>
          </tr>
          <tr>
            <td align="center" style="padding-bottom:32px;">
              <a href="%s" style="display:inline-block;background-color:#068bff;color:#ffffff;font-size:14px;font-weight:600;text-decoration:none;padding:12px 32px;border-radius:8px;">
                Reset password
              </a>
            </td>
          </tr>
          <tr>
            <td align="center">
              <p style="margin:0;font-size:12px;color:#4a5268;">
                Or paste this into your browser: <a href="%s" style="color:#1eaaff;text-decoration:none;">%s</a>
              </p>
            </td>
          </tr>
        </table>
        <table width="480" cellpadding="0" cellspacing="0" style="padding:24px 0;">
          <tr>
            <td align="center">
              <p style="margin:0;font-size:11px;color:#4a5268;">
                If you didn't request this, you can ignore this email — your password stays the same.
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, s.appURL, link, link, link)
}

func (s *EmailService) buildWelcomeHTML(toEmail string) string {
	dashboard := s.appURL + "/dashboard"
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0;padding:0;background-color:#0a0d14;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#0a0d14;padding:40px 20px;">
    <tr>
      <td align="center">
        <table width="480" cellpadding="0" cellspacing="0" style="background-color:#151a24;border:1px solid #252d3d;border-radius:12px;padding:40px;">
          <tr>
            <td align="center" style="padding-bottom:24px;">
              <img src="%s/clyra-logo.png" alt="Clyra" height="36" style="display:block;" />
            </td>
          </tr>
          <tr>
            <td align="center" style="padding-bottom:16px;">
              <h1 style="margin:0;font-size:20px;font-weight:600;color:#e8ecf4;">Welcome to Clyra</h1>
            </td>
          </tr>
          <tr>
            <td align="center" style="padding-bottom:24px;">
              <p style="margin:0;font-size:14px;line-height:22px;color:#9ba4b8;">
                Glad to have you. Clyra turns plain-language descriptions into
                cloud architecture diagrams you can edit, export, and share.
              </p>
            </td>
          </tr>
          <tr>
            <td style="padding-bottom:24px;">
              <p style="margin:0 0 8px 0;font-size:13px;line-height:20px;color:#9ba4b8;">A few things to try first:</p>
              <ul style="margin:0;padding-left:20px;font-size:13px;line-height:22px;color:#9ba4b8;">
                <li>Generate a diagram from a one-line prompt.</li>
                <li>Drag nodes around — your edits stick.</li>
                <li>Export as PNG or share a read-only link.</li>
              </ul>
            </td>
          </tr>
          <tr>
            <td align="center" style="padding-bottom:24px;">
              <a href="%s" style="display:inline-block;background-color:#068bff;color:#ffffff;font-size:14px;font-weight:600;text-decoration:none;padding:12px 32px;border-radius:8px;">
                Open Clyra
              </a>
            </td>
          </tr>
          <tr>
            <td align="center">
              <p style="margin:0;font-size:12px;color:#4a5268;">
                Got feedback or hit a snag? Just reply to this email — it comes to me directly.
              </p>
            </td>
          </tr>
        </table>
        <table width="480" cellpadding="0" cellspacing="0" style="padding:24px 0;">
          <tr>
            <td align="center">
              <p style="margin:0;font-size:11px;color:#4a5268;">
                Sent to %s · <a href="%s" style="color:#1eaaff;text-decoration:none;">useclyra.com</a>
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, s.appURL, dashboard, htmlEscape(toEmail), s.appURL)
}

func (s *EmailService) buildInviteHTML(toEmail, code string) string {
	link := fmt.Sprintf("%s/signup?invite=%s", s.appURL, code)

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0;padding:0;background-color:#0a0d14;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#0a0d14;padding:40px 20px;">
    <tr>
      <td align="center">
        <table width="480" cellpadding="0" cellspacing="0" style="background-color:#151a24;border:1px solid #252d3d;border-radius:12px;padding:40px;">
          <tr>
            <td align="center" style="padding-bottom:24px;">
              <img src="%s/clyra-logo.png" alt="Clyra" height="36" style="display:block;" />
            </td>
          </tr>
          <tr>
            <td align="center" style="padding-bottom:16px;">
              <h1 style="margin:0;font-size:20px;font-weight:600;color:#e8ecf4;">
                You're invited to Clyra
              </h1>
            </td>
          </tr>
          <tr>
            <td align="center" style="padding-bottom:32px;">
              <p style="margin:0;font-size:14px;line-height:22px;color:#9ba4b8;">
                Oche invited you to try Clyra — an AI-powered architecture
                diagram workbench for solutions architects, network architects,
                and enterprise architects.
              </p>
            </td>
          </tr>
          <tr>
            <td align="center" style="padding-bottom:16px;">
              <p style="margin:0;font-size:12px;color:#6b7488;text-transform:uppercase;letter-spacing:1px;">
                Your invite code
              </p>
            </td>
          </tr>
          <tr>
            <td align="center" style="padding-bottom:32px;">
              <code style="display:inline-block;background-color:#0a0d14;border:1px solid #252d3d;border-radius:8px;padding:12px 24px;font-size:18px;font-family:'JetBrains Mono',monospace;letter-spacing:3px;color:#1eaaff;">
                %s
              </code>
            </td>
          </tr>
          <tr>
            <td align="center" style="padding-bottom:32px;">
              <a href="%s" style="display:inline-block;background-color:#068bff;color:#ffffff;font-size:14px;font-weight:600;text-decoration:none;padding:12px 32px;border-radius:8px;">
                Get Started
              </a>
            </td>
          </tr>
          <tr>
            <td align="center">
              <p style="margin:0;font-size:12px;color:#4a5268;">
                Or go to <a href="%s" style="color:#1eaaff;text-decoration:none;">%s</a> and enter your code manually.
              </p>
            </td>
          </tr>
        </table>
        <table width="480" cellpadding="0" cellspacing="0" style="padding:24px 0;">
          <tr>
            <td align="center">
              <p style="margin:0;font-size:11px;color:#4a5268;">
                This invite was sent to %s. If you didn't expect this, you can ignore it.
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, s.appURL, code, link, s.appURL, s.appURL, toEmail)
}
