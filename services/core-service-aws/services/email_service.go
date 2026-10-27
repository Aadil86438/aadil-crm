package services

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"core-service-aws/config"
)

// SendApprovalEmail sends the CRM access code to the user after admin approval
func SendApprovalEmail(cfg *config.Config, toEmail, userName, crmCode string) error {
	if toEmail == "" {
		return fmt.Errorf("recipient email is required")
	}

	subject := fmt.Sprintf("Subject: 🎉 Your CRM Account Approved - Access Code: %s\r\n", crmCode)
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	
	fromAddr := cfg.SMTP.From
	if fromAddr == "" {
		fromAddr = cfg.SMTP.Username
	}
	if fromAddr == "" {
		fromAddr = "noreply@crm.local"
	}

	fromHeader := fmt.Sprintf("From: Proprietor CRM Admin <%s>\r\n", fromAddr)
	toHeader := fmt.Sprintf("To: %s\r\n", toEmail)

	loginURL := strings.TrimRight(cfg.FrontendURL, "/") + "/login"

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f1f5f9; margin: 0; padding: 20px; color: #1e293b; }
    .card { max-width: 580px; margin: 0 auto; background: #ffffff; border-radius: 16px; overflow: hidden; box-shadow: 0 10px 25px rgba(0,0,0,0.06); border: 1px solid #e2e8f0; }
    .header { background: linear-gradient(135deg, #1e3a8a 0%%, #2563eb 100%%); padding: 36px 30px; text-align: center; color: #ffffff; }
    .header h1 { margin: 0 0 8px 0; font-size: 24px; font-weight: 700; letter-spacing: -0.5px; }
    .header p { margin: 0; opacity: 0.85; font-size: 14px; }
    .content { padding: 32px 30px; }
    .greeting { font-size: 17px; font-weight: 600; margin-bottom: 12px; }
    .lead { font-size: 14px; line-height: 1.6; color: #475569; margin-bottom: 24px; }
    .code-box { background: #f8fafc; border: 2px dashed #3b82f6; border-radius: 12px; padding: 20px; text-align: center; margin-bottom: 28px; }
    .code-label { font-size: 12px; text-transform: uppercase; font-weight: 700; color: #64748b; letter-spacing: 1px; margin-bottom: 6px; }
    .code-value { font-size: 32px; font-weight: 800; color: #1d4ed8; letter-spacing: 3px; font-family: 'Courier New', Courier, monospace; }
    .steps { background: #eff6ff; border-radius: 10px; padding: 18px 20px; margin-bottom: 28px; }
    .steps-title { font-size: 13px; font-weight: 700; color: #1e40af; margin-bottom: 10px; }
    .steps ol { margin: 0; padding-left: 20px; font-size: 13px; line-height: 1.8; color: #1e3a8a; }
    .btn-container { text-align: center; margin: 28px 0 16px 0; }
    .btn { display: inline-block; background: #2563eb; color: #ffffff !important; text-decoration: none; padding: 12px 36px; border-radius: 8px; font-weight: 600; font-size: 14px; box-shadow: 0 4px 12px rgba(37,99,235,0.3); }
    .footer { padding: 20px 30px; background: #f8fafc; border-top: 1px solid #f1f5f9; text-align: center; font-size: 12px; color: #94a3b8; }
  </style>
</head>
<body>
  <div class="card">
    <div class="header">
      <h1>Proprietor CRM</h1>
      <p>Account Approval & Access Activation</p>
    </div>
    <div class="content">
      <div class="greeting">Hello %s,</div>
      <div class="lead">
        Great news! Your registration request has been reviewed and approved by the system administrator. You are now officially granted access to the CRM platform.
      </div>

      <div class="code-box">
        <div class="code-label">Your Unique CRM Access Code</div>
        <div class="code-value">%s</div>
      </div>

      <div class="steps">
        <div class="steps-title">How to Log In:</div>
        <ol>
          <li>Click the button below or open the CRM login page.</li>
          <li>Enter your registered email address and password.</li>
          <li>Enter your CRM Access Code: <strong>%s</strong></li>
        </ol>
      </div>

      <div class="btn-container">
        <a href="%s" class="btn" target="_blank">Sign In to Your CRM Account</a>
      </div>
    </div>
    <div class="footer">
      This is an automated notification. Please do not share your CRM Access Code with anyone.
    </div>
  </div>
</body>
</html>`, userName, crmCode, crmCode, loginURL)

	msg := []byte(fromHeader + toHeader + subject + mime + htmlBody)

	// If SMTP host is configured, attempt sending
	if cfg.SMTP.Host != "" && cfg.SMTP.Username != "" && cfg.SMTP.Password != "" {
		addr := fmt.Sprintf("%s:%d", cfg.SMTP.Host, cfg.SMTP.Port)
		auth := smtp.PlainAuth("", cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Host)

		err := smtp.SendMail(addr, auth, fromAddr, []string{toEmail}, msg)
		if err != nil {
			log.Printf("[APPROVAL EMAIL] Warning: Failed to dispatch email to %s: %v", toEmail, err)
			return err
		}
		log.Printf("[APPROVAL EMAIL] Successfully dispatched CRM Code %s email to %s", crmCode, toEmail)
		return nil
	}

	log.Printf("[APPROVAL EMAIL] (SMTP credentials not fully configured) Simulated email dispatch of CRM Code %s to %s", crmCode, toEmail)
	return nil
}
