package api

import (
	"strings"
	"time"

	"github.com/jamesread/data-cleaner/internal/config"
)

func applyDateToIncrementalStep(dateToIncremental *config.DateToIncrementalConfig, rows []DataRow) {
	if dateToIncremental == nil || !dateToIncremental.Configured() {
		return
	}
	column := dateToIncremental.Column

	var lastOriginal time.Time
	var lastOutput time.Time
	var haveLast bool

	for i := range rows {
		value := strings.TrimSpace(rows[i].Get(column))
		if value == "" {
			continue
		}

		original, ok := parseDateTimeValue(value)
		if !ok {
			continue
		}

		out := original
		if haveLast && original.Equal(lastOriginal) {
			out = lastOutput.Add(time.Second)
		}

		rows[i].Set(column, out.Format(isoDateTimeLayout))
		lastOriginal = original
		lastOutput = out
		haveLast = true
	}
}

func parseDateTimeValue(value string) (time.Time, bool) {
	return parseDateWithFallback(value, isoDateTimeLayout)
}
