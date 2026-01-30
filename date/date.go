package date

import "time"

// GetMondayOfWeek returns the Monday of the given week
func GetMondayOfWeek(week int) time.Time {
	now := time.Now()

	// Get January 1 of the current year
	startOfYear := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, time.Local)

	// Calculate the first Monday on or after January 1
	firstMonday := startOfYear
	for firstMonday.Weekday() != time.Monday {
		firstMonday = firstMonday.AddDate(0, 0, 1)
	}

	// Add the offset for the given week (weeks start from 0)
	daysToAdd := (week - 2) * 7
	mondayOfWeek := firstMonday.AddDate(0, 0, daysToAdd)

	return mondayOfWeek
}
