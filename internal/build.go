package internal

import (
	"strings"
	"time"
)

var BuildCommit string       // nolint:gochecknoglobals // Filled at compilation time.
var BuildVersion string      // nolint:gochecknoglobals // Filled at compilation time.
var BuildTime string         // nolint:gochecknoglobals // Filled at compilation time.
var BuildTimeFmtd *time.Time // nolint:gochecknoglobals // Filled at int the init.

func init() {
	if BuildTime != "" {
		parsedTime, err := time.Parse(time.RFC3339, strings.ReplaceAll(BuildTime, "UTC", "Z"))
		if err != nil {
			panic(err)
		}
		BuildTimeFmtd = &parsedTime
	}
}
