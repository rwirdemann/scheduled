package telegram

import "github.com/rwirdemann/scheduled"

// ParseWeekday delegates to the root package implementation.
func ParseWeekday(text string) (day int, cleanName string) {
	return scheduled.ParseWeekday(text)
}
