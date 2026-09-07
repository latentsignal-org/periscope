package server

import "time"

// defaultDateRange returns (from, to) defaulting to the last
// 30 days if not provided.
func defaultDateRange(
	from, to string,
) (string, string) {
	now := time.Now().UTC()
	if to == "" {
		to = now.Format("2006-01-02")
	}
	if from == "" {
		t, err := time.Parse("2006-01-02", to)
		if err != nil {
			t = now
		}
		from = t.AddDate(0, 0, -30).Format("2006-01-02")
	}
	return from, to
}
