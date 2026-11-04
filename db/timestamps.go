package db

import "time"

func Timestamp(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}
