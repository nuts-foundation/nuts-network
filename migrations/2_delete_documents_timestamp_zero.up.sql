-- Unix nanoseconds value for zero Golang time.Time as SQLite BIGINT is -6795364578871345152
DELETE FROM document WHERE timestamp = -6795364578871345152;