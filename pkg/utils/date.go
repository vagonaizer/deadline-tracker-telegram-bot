package utils

import (
	"fmt"
	"strings"
	"time"
)

func ParseDeadline(dateStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"02.01.2006 15:04",
		"02.01.2006",
		"2006-01-02",
		"02/01/2006 15:04",
		"02/01/2006",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			if t.IsZero() {
				return time.Time{}, fmt.Errorf("invalid date")
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func FormatDeadline(t time.Time) string {
	return t.Format("02.01.2006 15:04")
}

func FormatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := t.Sub(now)

	if diff < 0 {
		diff = -diff
		if diff < time.Hour {
			return fmt.Sprintf("%d минут назад", int(diff.Minutes()))
		} else if diff < 24*time.Hour {
			return fmt.Sprintf("%d часов назад", int(diff.Hours()))
		} else {
			return fmt.Sprintf("%d дней назад", int(diff.Hours()/24))
		}
	}

	if diff < time.Hour {
		return fmt.Sprintf("через %d минут", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("через %d часов", int(diff.Hours()))
	} else {
		return fmt.Sprintf("через %d дней", int(diff.Hours()/24))
	}
}

func GetPriorityByDeadline(deadline time.Time) string {
	now := time.Now()
	diff := deadline.Sub(now)

	if diff < 0 {
		return "critical"
	} else if diff < 24*time.Hour {
		return "critical"
	} else if diff < 3*24*time.Hour {
		return "high"
	} else if diff < 7*24*time.Hour {
		return "normal"
	} else {
		return "low"
	}
}

func IsOverdue(deadline time.Time) bool {
	return deadline.Before(time.Now())
}

func GetNotificationTimes(deadline time.Time, reminderDays []int) []time.Time {
	var times []time.Time

	for _, days := range reminderDays {
		notifyTime := deadline.AddDate(0, 0, -days)
		if notifyTime.After(time.Now()) {
			times = append(times, notifyTime)
		}
	}

	return times
}

func FormatDuration(d time.Duration) string {
	if d < 0 {
		return "просрочено"
	}

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	var parts []string

	if days > 0 {
		if days == 1 {
			parts = append(parts, "1 день")
		} else {
			parts = append(parts, fmt.Sprintf("%d дней", days))
		}
	}

	if hours > 0 {
		if hours == 1 {
			parts = append(parts, "1 час")
		} else {
			parts = append(parts, fmt.Sprintf("%d часов", hours))
		}
	}

	if minutes > 0 && days == 0 {
		if minutes == 1 {
			parts = append(parts, "1 минута")
		} else {
			parts = append(parts, fmt.Sprintf("%d минут", minutes))
		}
	}

	if len(parts) == 0 {
		return "менее минуты"
	}

	return strings.Join(parts, " ")
}

func GetWeekStart(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
}

func GetMonthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}
