package model

import "time"

// MarshalDocumentTime converts a document's time.Time to unix nanoseconds.
func MarshalDocumentTime(ts time.Time) int64 {
	if ts.IsZero() {
		return 0
	} else {
		return ts.UTC().UnixNano()
	}
}

// UnmarshalDocumentTime converts a document's timestamp in unix nanoseconds to time.Time.
func UnmarshalDocumentTime(ns int64) time.Time {
	if ns <= 0 {
		return time.Time{}
	} else {
		return time.Unix(0, ns).UTC()
	}
}
