package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"runtime"
	"strings"
	"time"

	"crm/config"
	"crm/database"
)

var startTime = time.Now()

// SystemHealthData holds telemetry metrics for report generation
type SystemHealthData struct {
	Timestamp      string
	AppEnv         string
	OverallStatus  string // "HEALTHY" or "DEGRADED" / "CRITICAL"
	Uptime         string
	DBStatus       string
	DBStatusClass  string
	RedisStatus    string
	RedisClass     string
	GoroutineCount int
	MemAllocMB     string
	MemSysMB       string
	ActiveUsers    int64
	PendingRegs    int64
	TotalLeads     int64
	FrontendURL    string
}

// PerformHealthCheck collects real-time system metrics
func PerformHealthCheck(cfg *config.Config) SystemHealthData {
	uptimeDuration := time.Since(startTime)
	days := int(uptimeDuration.Hours()) / 24
	hours := int(uptimeDuration.Hours()) % 24
	mins := int(uptimeDuration.Minutes()) % 60

	uptimeStr := fmt.Sprintf("%dd %dh %dm", days, hours, mins)

	// Check Database
	dbStatus := "CONNECTED"
	dbClass := "status-ok"
	if database.DB == nil {
		dbStatus = "DISCONNECTED"
		dbClass = "status-error"
	} else if err := database.DB.Ping(); err != nil {
		dbStatus = "UNREACHABLE"
		dbClass = "status-error"
	}

	// Check Redis
	redisStatus := "CONNECTED"
	redisClass := "status-ok"
	if database.RedisClient == nil {
		redisStatus = "DISABLED / DISCONNECTED"
		redisClass = "status-warning"
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := database.RedisClient.Ping(ctx).Err(); err != nil {
			redisStatus = "UNREACHABLE"
			redisClass = "status-warning"
		}
	}

	// Runtime metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memAlloc := fmt.Sprintf("%.2f MB", float64(m.Alloc)/(1024*1024))
	memSys := fmt.Sprintf("%.2f MB", float64(m.Sys)/(1024*1024))

	// Database counts
	var activeUsers, pendingRegs, totalLeads int64
	if database.DB != nil {
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&activeUsers)
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM registration_requests WHERE status = 'PENDING'").Scan(&pendingRegs)
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM leads").Scan(&totalLeads)
	}

	overallStatus := "HEALTHY"
	if dbClass == "status-error" {
		overallStatus = "CRITICAL"
	} else if redisClass == "status-warning" {
		overallStatus = "OPTIMAL"
	}

	return SystemHealthData{
		Timestamp:      time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"),
		AppEnv:         strings.ToUpper(cfg.AppEnv),
		OverallStatus:  overallStatus,
		Uptime:         uptimeStr,
		DBStatus:       dbStatus,
		DBStatusClass:  dbClass,
		RedisStatus:    redisStatus,
		RedisClass:     redisClass,
		GoroutineCount: runtime.NumGoroutine(),
		MemAllocMB:     memAlloc,
		MemSysMB:       memSys,
		ActiveUsers:    activeUsers,
		PendingRegs:    pendingRegs,
		TotalLeads:     totalLeads,
		FrontendURL:    cfg.FrontendURL,
	}
}

// SendHealthReportEmail compiles the premium HTML template and sends it via SMTP
func SendHealthReportEmail(cfg *config.Config) error {
	recipient := cfg.HealthReport.Recipient
	if recipient == "" {
		recipient = cfg.SMTP.Username
	}

	if recipient == "" || cfg.SMTP.Host == "" {
		log.Println("[HEALTH CHECK] Skipped email sending: SMTP credentials or recipient not configured")
		return nil
	}

	data := PerformHealthCheck(cfg)
	body, err := renderEmailTemplate(data)
	if err != nil {
		return fmt.Errorf("failed to render HTML email template: %w", err)
	}

	subject := fmt.Sprintf("Subject: 🚀 [Daily Health Report] Status: %s - %s\r\n", data.OverallStatus, time.Now().Format("02 Jan 2006"))
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	fromHeader := fmt.Sprintf("From: CRM Service Monitor <%s>\r\n", cfg.SMTP.From)
	toHeader := fmt.Sprintf("To: %s\r\n", recipient)

	msg := []byte(fromHeader + toHeader + subject + mime + body)

	addr := fmt.Sprintf("%s:%d", cfg.SMTP.Host, cfg.SMTP.Port)
	var auth smtp.Auth
	if cfg.SMTP.Username != "" && cfg.SMTP.Password != "" {
		auth = smtp.PlainAuth("", cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Host)
	}

	err = smtp.SendMail(addr, auth, cfg.SMTP.From, []string{recipient}, msg)
	if err != nil {
		log.Printf("[HEALTH CHECK] Failed to send email: %v", err)
		return err
	}

	log.Printf("[HEALTH CHECK] Successfully dispatched Daily Health Report to %s", recipient)
	return nil
}

func renderEmailTemplate(data SystemHealthData) (string, error) {
	tmpl, err := template.New("healthReport").Parse(emailHTMLTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

const emailHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Daily System Health Status Report</title>
  <style>
    body {
      margin: 0;
      padding: 0;
      background-color: #0b0f19;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
      color: #e2e8f0;
      -webkit-font-smoothing: antialiased;
    }
    .wrapper {
      width: 100%;
      table-layout: fixed;
      background-color: #0b0f19;
      padding: 40px 10px;
    }
    .main-card {
      max-width: 600px;
      margin: 0 auto;
      background: linear-gradient(145deg, #131b2e 0%, #0f172a 100%);
      border: 1px solid #1e293b;
      border-radius: 16px;
      overflow: hidden;
      box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5), 0 8px 10px -6px rgba(0, 0, 0, 0.5);
    }
    .header {
      background: linear-gradient(90deg, #4f46e5 0%, #7c3aed 100%);
      padding: 32px 24px;
      text-align: center;
    }
    .header h1 {
      margin: 0;
      font-size: 24px;
      font-weight: 800;
      color: #ffffff;
      letter-spacing: -0.5px;
    }
    .header p {
      margin: 6px 0 0 0;
      font-size: 13px;
      color: #e0e7ff;
      opacity: 0.9;
    }
    .content {
      padding: 32px 24px;
    }
    .status-banner {
      background: rgba(30, 41, 59, 0.7);
      border: 1px solid #334155;
      border-radius: 12px;
      padding: 16px 20px;
      margin-bottom: 24px;
      display: flex;
      align-items: center;
      justify-content: space-between;
    }
    .badge {
      display: inline-block;
      padding: 6px 14px;
      border-radius: 9999px;
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 0.5px;
      text-transform: uppercase;
    }
    .status-ok {
      background-color: rgba(16, 185, 129, 0.15);
      color: #34d399;
      border: 1px solid rgba(52, 211, 153, 0.3);
    }
    .status-warning {
      background-color: rgba(245, 158, 11, 0.15);
      color: #fbbf24;
      border: 1px solid rgba(251, 191, 36, 0.3);
    }
    .status-error {
      background-color: rgba(239, 68, 68, 0.15);
      color: #f87171;
      border: 1px solid rgba(248, 113, 113, 0.3);
    }
    .section-title {
      font-size: 14px;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 1px;
      color: #94a3b8;
      margin: 24px 0 12px 0;
      padding-bottom: 6px;
      border-bottom: 1px solid #1e293b;
    }
    .grid {
      width: 100%;
      border-collapse: collapse;
      margin-bottom: 16px;
    }
    .grid td {
      padding: 10px 12px;
      border-bottom: 1px solid #1e293b;
      font-size: 14px;
    }
    .grid td.label {
      color: #94a3b8;
      font-weight: 500;
      width: 45%;
    }
    .grid td.val {
      color: #f8fafc;
      font-weight: 600;
      text-align: right;
    }
    .cta-btn {
      display: block;
      width: 100%;
      text-align: center;
      background: linear-gradient(135deg, #6366f1 0%, #4f46e5 100%);
      color: #ffffff;
      text-decoration: none;
      font-weight: 700;
      font-size: 14px;
      padding: 14px 0;
      border-radius: 10px;
      margin-top: 28px;
      box-shadow: 0 4px 14px rgba(99, 102, 241, 0.4);
    }
    .footer {
      padding: 20px 24px;
      background-color: #070a12;
      text-align: center;
      font-size: 12px;
      color: #64748b;
      border-top: 1px solid #1e293b;
    }
  </style>
</head>
<body>
  <div class="wrapper">
    <div class="main-card">
      <!-- Header -->
      <div class="header">
        <h1>📊 Daily Service Health Status</h1>
        <p>AWS EC2 Production Environment Telemetry</p>
      </div>

      <!-- Content -->
      <div class="content">
        <!-- Status Banner -->
        <table width="100%" style="margin-bottom: 24px;">
          <tr>
            <td style="font-size: 15px; font-weight: 600; color: #94a3b8;">System Health Overview</td>
            <td align="right">
              <span class="badge {{if eq .OverallStatus "HEALTHY"}}status-ok{{else if eq .OverallStatus "OPTIMAL"}}status-ok{{else}}status-error{{end}}">
                ● {{.OverallStatus}}
              </span>
            </td>
          </tr>
        </table>

        <!-- Infrastructure & Database Health -->
        <div class="section-title">Core Infrastructure Status</div>
        <table class="grid">
          <tr>
            <td class="label">PostgreSQL Database</td>
            <td class="val"><span class="badge {{.DBStatusClass}}">{{.DBStatus}}</span></td>
          </tr>
          <tr>
            <td class="label">Redis Cache Engine</td>
            <td class="val"><span class="badge {{.RedisClass}}">{{.RedisStatus}}</span></td>
          </tr>
          <tr>
            <td class="label">Backend Uptime</td>
            <td class="val">{{.Uptime}}</td>
          </tr>
          <tr>
            <td class="label">Active Environment</td>
            <td class="val">{{.AppEnv}}</td>
          </tr>
        </table>

        <!-- Go Telemetry -->
        <div class="section-title">Backend Performance Metrics</div>
        <table class="grid">
          <tr>
            <td class="label">Active Goroutines</td>
            <td class="val">{{.GoroutineCount}} workers</td>
          </tr>
          <tr>
            <td class="label">Heap Memory Alloc</td>
            <td class="val">{{.MemAllocMB}}</td>
          </tr>
          <tr>
            <td class="label">Total System Memory</td>
            <td class="val">{{.MemSysMB}}</td>
          </tr>
        </table>

        <!-- Business Telemetry -->
        <div class="section-title">Application Telemetry Summary</div>
        <table class="grid">
          <tr>
            <td class="label">Total Registered Users</td>
            <td class="val">{{.ActiveUsers}}</td>
          </tr>
          <tr>
            <td class="label">Pending User Approvals</td>
            <td class="val" style="color: #fbbf24;">{{.PendingRegs}} pending</td>
          </tr>
          <tr>
            <td class="label">Total CRM Leads</td>
            <td class="val">{{.TotalLeads}}</td>
          </tr>
        </table>

        <a href="{{.FrontendURL}}/admin" class="cta-btn">Open CRM Admin Dashboard →</a>
      </div>

      <!-- Footer -->
      <div class="footer">
        Automated Nightly Report generated on {{.Timestamp}}<br>
        Powered by CRM Automated Goroutine Monitor • AWS EC2 Free Tier
      </div>
    </div>
  </div>
</body>
</html>`
