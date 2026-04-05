package domain

import "time"

type TogglEntry struct {
	Description string
	Start       time.Time
	End         time.Time
	ProjectID   ProjectID
}
