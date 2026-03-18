package telegram

import (
	"regexp"
	"strings"
)

var weekdayPrefixes = []struct {
	patterns []string
	day      int
}{
	{[]string{"montag", "mo", "monday", "mon"}, 1},
	{[]string{"dienstag", "di", "tuesday", "tue"}, 2},
	{[]string{"mittwoch", "mi", "wednesday", "wed"}, 3},
	{[]string{"donnerstag", "do", "thursday", "thu"}, 4},
	{[]string{"freitag", "fr", "friday", "fri"}, 5},
	{[]string{"samstag", "sa", "saturday", "sat"}, 6},
	{[]string{"sonntag", "so", "sunday", "sun"}, 7},
}

var prefixRe = regexp.MustCompile(`(?i)^(\S+)\s+(.+)$`)

// ParseWeekday checks if text starts with a weekday prefix (German or English,
// long or abbreviated). If so, it returns the corresponding day number (1–7)
// and the task name with the prefix stripped. Otherwise it returns day 0
// (Inbox) and the original text unchanged.
func ParseWeekday(text string) (day int, cleanName string) {
	m := prefixRe.FindStringSubmatch(text)
	if m == nil {
		return 0, text
	}
	prefix := strings.ToLower(m[1])
	for _, entry := range weekdayPrefixes {
		for _, p := range entry.patterns {
			if prefix == p {
				return entry.day, strings.TrimSpace(m[2])
			}
		}
	}
	return 0, text
}
