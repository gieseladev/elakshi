package iso8601

import "time"

// ParseDatetime parses a ISO 8601 formatted date.
// NOTE: Currently only supports RFC3339 subset!
func ParseDatetime(t string) (time.Time, error) {
	return time.Parse(time.RFC3339, t)
}
