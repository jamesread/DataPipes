package api

import (
	"strconv"
	"strings"
	"time"

	"github.com/jamesread/data-cleaner/internal/config"
	log "github.com/sirupsen/logrus"
)

const isoDateTimeLayout = "2006-01-02 15:04:05"

func ensureDateLayouts(cfg *config.Config, st *jobState, rows []Row) {
	if cfg == nil {
		return
	}
	if st.dateLayouts == nil {
		st.dateLayouts = make(map[string]string)
	}
	for _, step := range cfg.TransformSteps() {
		if step.Kind() != "date_normalize" {
			continue
		}
		for _, col := range step.DateNormalize.TargetColumns() {
			if _, ok := st.dateLayouts[col]; ok {
				continue
			}
			values := columnValues(rows, col)
			layout := detectDateLayout(values)
			st.dateLayouts[col] = layout
			log.Infof("date_normalize: detected layout %q for column %q", layoutDescription(layout), col)
		}
	}
}

func columnValues(rows []Row, column string) []string {
	values := make([]string, 0, len(rows))
	for _, row := range rows {
		if v, ok := row.Columns[column]; ok {
			values = append(values, v)
		}
	}
	return values
}

func detectDateLayout(values []string) string {
	nonEmpty := filterNonEmptyTrimmed(values)
	if len(nonEmpty) == 0 {
		return time.DateOnly
	}

	slashLayout := ""
	slashValues := make([]string, 0, len(nonEmpty))
	for _, v := range nonEmpty {
		if isSlashDate(v) {
			slashValues = append(slashValues, v)
		}
	}
	if len(slashValues) > 0 {
		slashLayout = resolveSlashDateLayout(slashValues)
	}

	best := time.DateOnly
	bestScore := -1
	for _, layout := range candidateLayouts(slashLayout) {
		score := 0
		for _, v := range nonEmpty {
			if parsesDate(v, layout) {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			best = layout
		}
	}
	return best
}

func candidateLayouts(slashLayout string) []string {
	layouts := []string{
		time.DateOnly,
		isoDateTimeLayout,
		"2006-01-02T15:04:05",
		time.RFC3339,
		"2006/01/02",
		"20060102",
		"02-01-2006",
		"01-02-2006",
		"02.01.2006",
		"2006.01.02",
	}
	if slashLayout != "" {
		layouts = append(layouts, slashLayout, slashLayout+" 15:04:05")
	} else {
		layouts = append(layouts,
			"02/01/2006", "01/02/2006", "2/1/2006", "02/01/06", "01/02/06",
			"02/01/2006 15:04:05", "01/02/2006 15:04:05",
		)
	}
	return layouts
}

func filterNonEmptyTrimmed(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func isSlashDate(value string) bool {
	parts := strings.Split(datePart(value), "/")
	return len(parts) == 3
}

func datePart(value string) string {
	value = strings.TrimSpace(value)
	if idx := strings.IndexAny(value, " T"); idx >= 0 {
		return strings.TrimSpace(value[:idx])
	}
	return value
}

func resolveSlashDateLayout(values []string) string {
	ddmmVotes := 0
	mmddVotes := 0
	for _, v := range values {
		parts := strings.Split(datePart(v), "/")
		if len(parts) != 3 {
			continue
		}
		a, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		b, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil {
			continue
		}
		if a > 12 {
			ddmmVotes++
		}
		if b > 12 {
			mmddVotes++
		}
	}
	if mmddVotes > ddmmVotes {
		return slashLayoutForYearWidth(values, "01/02/2006", "01/02/06")
	}
	return slashLayoutForYearWidth(values, "02/01/2006", "02/01/06")
}

func slashLayoutForYearWidth(values []string, fourDigit, twoDigit string) string {
	for _, v := range values {
		parts := strings.Split(datePart(v), "/")
		if len(parts) != 3 {
			continue
		}
		if len(strings.TrimSpace(parts[2])) <= 2 {
			return twoDigit
		}
	}
	return fourDigit
}

func parsesDate(value, layout string) bool {
	_, err := time.Parse(layout, strings.TrimSpace(value))
	return err == nil
}

func parseDate(value, layout string) (time.Time, bool) {
	t, err := time.Parse(layout, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func parseDateWithFallback(value, primaryLayout string) (time.Time, bool) {
	if t, ok := parseDate(value, primaryLayout); ok {
		return t, true
	}
	for _, layout := range candidateLayouts("") {
		if layout == primaryLayout {
			continue
		}
		if t, ok := parseDate(value, layout); ok {
			return t, true
		}
	}
	return time.Time{}, false
}

func normalizeDateValue(value, layout string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	t, ok := parseDateWithFallback(value, layout)
	if !ok {
		return value
	}
	return t.Format(isoDateTimeLayout)
}

func applyDateNormalizeStep(dateNormalize *config.DateNormalizeConfig, row *DataRow, dateLayouts map[string]string) {
	if dateNormalize == nil || !dateNormalize.Configured() {
		return
	}
	for _, col := range dateNormalize.TargetColumns() {
		layout, ok := dateLayouts[col]
		if !ok {
			layout = time.DateOnly
		}
		row.Set(col, normalizeDateValue(row.Get(col), layout))
	}
}

func layoutDescription(layout string) string {
	switch layout {
	case time.DateOnly:
		return "ISO date (YYYY-MM-DD)"
	case isoDateTimeLayout, "2006-01-02T15:04:05":
		return "ISO datetime"
	case time.RFC3339:
		return "RFC3339"
	case "02/01/2006", "2/1/2006", "02/01/06":
		return "DD/MM/YYYY"
	case "01/02/2006", "01/02/06":
		return "MM/DD/YYYY"
	case "2006/01/02":
		return "YYYY/MM/DD"
	case "20060102":
		return "YYYYMMDD"
	case "02-01-2006", "01-02-2006":
		return "DD-MM-YYYY or MM-DD-YYYY"
	case "02.01.2006", "2006.01.02":
		return "DD.MM.YYYY"
	default:
		return layout
	}
}
