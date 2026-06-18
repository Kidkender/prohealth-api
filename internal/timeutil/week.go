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

func SplitTimeRange(start, end time.Time, chunkDuration time.Duration) [][2]time.Time {
	var chunks [][2]time.Time
	cur := start
	for cur.Before(end) {
		next := cur.Add(chunkDuration)
		if next.After(end) {
			next = end
		}
		chunks = append(chunks, [2]time.Time{cur, next})
		cur = next
	}
	return chunks
}

func SplitTimeRangeBalanced(start, end time.Time) [][2]time.Time {
	totalHours := int(end.Sub(start).Hours())
	if totalHours <= 2 {
		return [][2]time.Time{{start, end}}
	}

	q := totalHours / 3
	r := totalHours % 3

	var sizes []time.Duration
	switch r {
	case 0:
		for i := 0; i < q; i++ {
			sizes = append(sizes, 3*time.Hour)
		}
	case 1:
		for i := 0; i < q-1; i++ {
			sizes = append(sizes, 3*time.Hour)
		}
		sizes = append(sizes, 2*time.Hour, 2*time.Hour)
	case 2:
		for i := 0; i < q; i++ {
			sizes = append(sizes, 3*time.Hour)
		}
		sizes = append(sizes, 2*time.Hour)
	}

	chunks := make([][2]time.Time, 0, len(sizes))
	cur := start
	for i, size := range sizes {
		next := cur.Add(size)
		if i == len(sizes)-1 {
			next = end
		}
		chunks = append(chunks, [2]time.Time{cur, next})
		cur = next
	}
	return chunks
}

func NextMonday(input *time.Time) time.Time {
	if input != nil {
		return *input
	}
	now := time.Now().UTC().Truncate(24 * time.Hour)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	daysUntil := (8 - weekday) % 7
	if daysUntil == 0 {
		daysUntil = 7
	}
	return now.AddDate(0, 0, daysUntil)
}
