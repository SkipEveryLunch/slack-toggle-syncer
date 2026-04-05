package domain

import (
	"fmt"
	"strconv"
	"time"
)

var JST = time.FixedZone("Asia/Tokyo", 9*60*60)

type SourceMessage struct {
	Text      string
	Timestamp time.Time
}

func ParseSlackTimestamp(ts string) (time.Time, error) {
	unixTime, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("strconv.ParseFloat: %w", err)
	}
	return time.Unix(int64(unixTime), 0).In(JST), nil
}
