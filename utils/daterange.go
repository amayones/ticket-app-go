package utils

// Shared date-range helper for dashboards (features/organizer,
// features/seller). Presets: today, 7days, 30days, month, custom.
// Output is YYYY-MM-DD strings, inclusive on both ends. Ranges longer
// than MaxDashboardDays are rejected by the caller via ErrInvalidFilter.
import (
	"fmt"
	"time"
)

const (
	// MaxDashboardDays caps dashboard ranges (spec: maximum 1 year).
	MaxDashboardDays = 366
	// DefaultDashboardDays is the default lookback (spec: last 30 days).
	DefaultDashboardDays = 30
)

// DateRange is a validated inclusive [From, To] range (YYYY-MM-DD).
type DateRange struct {
	From string
	To   string
}

// ParseDateRange resolves preset/from/to query params into a DateRange.
// Empty preset defaults to the last 30 days. Custom requires from+to.
func ParseDateRange(preset, from, to string) (DateRange, error) {
	now := time.Now()
	today := now.Format("2006-01-02")
	switch preset {
	case "", "30days":
		return DateRange{
			From: now.AddDate(0, 0, -(DefaultDashboardDays - 1)).Format("2006-01-02"),
			To:   today,
		}, nil
	case "today":
		return DateRange{From: today, To: today}, nil
	case "7days":
		return DateRange{
			From: now.AddDate(0, 0, -6).Format("2006-01-02"),
			To:   today,
		}, nil
	case "month":
		return DateRange{
			From: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02"),
			To:   today,
		}, nil
	case "custom":
		if from == "" || to == "" {
			return DateRange{}, fmt.Errorf("custom range requires from and to (YYYY-MM-DD)")
		}
		if !IsValidDate(from) || !IsValidDate(to) {
			return DateRange{}, fmt.Errorf("dates must be YYYY-MM-DD")
		}
		if from > to {
			return DateRange{}, fmt.Errorf("from must not be after to")
		}
		if daysBetween(from, to) > MaxDashboardDays {
			return DateRange{}, fmt.Errorf("range must not exceed %d days", MaxDashboardDays)
		}
		return DateRange{From: from, To: to}, nil
	default:
		return DateRange{}, fmt.Errorf("unknown preset: %s (today, 7days, 30days, month, custom)", preset)
	}
}

// IsValidDate reports whether s is YYYY-MM-DD.
func IsValidDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func daysBetween(from, to string) int {
	a, err1 := time.Parse("2006-01-02", from)
	b, err2 := time.Parse("2006-01-02", to)
	if err1 != nil || err2 != nil {
		return 0
	}
	return int(b.Sub(a).Hours()/24) + 1
}

// PreviousRange returns the same-length period immediately before r
// (for up/down comparisons with the previous period).
func PreviousRange(r DateRange) DateRange {
	a, err1 := time.Parse("2006-01-02", r.From)
	b, err2 := time.Parse("2006-01-02", r.To)
	if err1 != nil || err2 != nil {
		return r
	}
	length := int(b.Sub(a).Hours()/24) + 1
	prevTo := a.AddDate(0, 0, -1)
	prevFrom := prevTo.AddDate(0, 0, -(length - 1))
	return DateRange{From: prevFrom.Format("2006-01-02"), To: prevTo.Format("2006-01-02")}
}

// DeltaPct computes the percent change from previous to current
// (0 when previous is 0 and current is 0; 100 when growing from 0).
func DeltaPct(current, previous float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return (current - previous) / previous * 100
}
