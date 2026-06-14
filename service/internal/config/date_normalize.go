package config

import "sort"

type DateNormalizeConfig struct {
	Column  string   `yaml:"column,omitempty"`
	Columns []string `yaml:"columns,omitempty"`
}

func (d *DateNormalizeConfig) Configured() bool {
	return d != nil && len(d.TargetColumns()) > 0
}

func (d *DateNormalizeConfig) TargetColumns() []string {
	if d == nil {
		return nil
	}
	if d.Column != "" {
		return []string{d.Column}
	}
	out := make([]string, 0, len(d.Columns))
	for _, col := range d.Columns {
		if col != "" {
			out = append(out, col)
		}
	}
	return out
}

func SortedDateNormalizeColumns(cols []string) []string {
	out := append([]string(nil), cols...)
	sort.Strings(out)
	return out
}
