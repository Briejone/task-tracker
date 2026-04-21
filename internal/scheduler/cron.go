package scheduler

import (
	"time"

	"github.com/robfig/cron/v3"
)

	func CalculateNextRun(cronExpr string) (time.Time, error) {
		parser := cron.NewParser(
			cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
		)

		schedule,err := parser.Parse(cronExpr)
		if err != nil {
			return time.Time{}, err
		}
		return schedule.Next(time.Now()), nil
	}


	func ValidateCron(cronExpr string) bool {
		parser := cron.NewParser(
			cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
		)

		_, err := parser.Parse(cronExpr)
		return err == nil
	}