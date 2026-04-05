package domain

import "time"

type TimeEntry struct {
	Description string
	Start       time.Time
	End         time.Time
	ProjectID   int64
}
