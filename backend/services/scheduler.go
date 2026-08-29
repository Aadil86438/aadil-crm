package services

import (
	"log"
	"time"

	"crm/config"
)

// StartDailyHealthReport launches a background goroutine that triggers every night at the specified time (e.g. 21:00 / 9:00 PM)
func StartDailyHealthReport(cfg *config.Config) {
	if !cfg.HealthReport.Enabled {
		log.Println("[SCHEDULER] Daily health report is disabled by configuration")
		return
	}

	go func() {
		log.Printf("[SCHEDULER] Daily health report background goroutine initialized. Scheduled time: %s daily", cfg.HealthReport.Time)

		for {
			now := time.Now()
			targetTimeStr := cfg.HealthReport.Time
			if targetTimeStr == "" {
				targetTimeStr = "21:00"
			}

			targetHour, targetMin, err := parseTimeHM(targetTimeStr)
			if err != nil {
				log.Printf("[SCHEDULER] Invalid HEALTH_REPORT_TIME format '%s', falling back to 21:00: %v", targetTimeStr, err)
				targetHour, targetMin = 21, 0
			}

			// Calculate next occurrence of 9:00 PM
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), targetHour, targetMin, 0, 0, now.Location())
			if now.After(nextRun) {
				// Target time has already passed today, schedule for tomorrow
				nextRun = nextRun.Add(24 * time.Hour)
			}

			durationUntilNextRun := time.Until(nextRun)
			log.Printf("[SCHEDULER] Next Daily Health Report scheduled in %v (at %s)", durationUntilNextRun.Round(time.Second), nextRun.Format("2006-01-02 15:04:05 MST"))

			// Wait until 9:00 PM
			timer := time.NewTimer(durationUntilNextRun)
			<-timer.C

			// Dispatch the report in a worker goroutine so scheduler loop remains responsive
			log.Println("[SCHEDULER] Triggering 9:00 PM Daily Health Status Report dispatch...")
			go func() {
				if err := SendHealthReportEmail(cfg); err != nil {
					log.Printf("[SCHEDULER] Failed to send scheduled daily health report: %v", err)
				}
			}()

			// Sleep for 2 minutes to prevent double trigger within the same minute
			time.Sleep(2 * time.Minute)
		}
	}()
}

func parseTimeHM(hm string) (int, int, error) {
	t, err := time.Parse("15:04", hm)
	if err != nil {
		return 0, 0, err
	}
	return t.Hour(), t.Minute(), nil
}
