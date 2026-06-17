package timeutil

import "time"

func WeekRange(t time.Time) (start, end time.Time) {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start = t.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)
	end = start.AddDate(0, 0, 7)
	return start, end
}

func CurrentWeekRange() (start, end time.Time) {
	return WeekRange(time.Now())
}
