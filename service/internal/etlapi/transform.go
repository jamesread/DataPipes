package api

import (
	"regexp"
	"strings"

	"github.com/jamesread/data-cleaner/internal/config"
	log "github.com/sirupsen/logrus"
)

type DataRow struct {
	contents map[string]string
}

func (r *DataRow) Set(key, value string) {
	r.contents[key] = value
}

func (r *DataRow) Get(key string) string {
	if val, ok := r.contents[key]; ok {
		return val
	}
	return ""
}

func (r *DataRow) ToSlice(columns []string) []string {
	ret := make([]string, len(columns))
	for i, col := range columns {
		ret[i] = r.Get(col)
	}
	return ret
}

func copyColumnValues(row Row) map[string]string {
	out := make(map[string]string, len(row.Columns))
	for k, v := range row.Columns {
		out[k] = v
	}
	return out
}

func ensureColumnInOrder(order []string, col string) []string {
	for _, c := range order {
		if c == col {
			return order
		}
	}
	return append(order, col)
}

func applyDropColumn(drop []string, row map[string]string, columnOrder []string) (map[string]string, []string) {
	if len(drop) == 0 {
		return row, columnOrder
	}

	dropSet := make(map[string]bool, len(drop))
	for _, col := range drop {
		dropSet[col] = true
	}

	out := make(map[string]string, len(row))
	outOrder := make([]string, 0, len(columnOrder))
	for _, col := range columnOrder {
		if dropSet[col] {
			continue
		}
		if val, ok := row[col]; ok {
			out[col] = val
			outOrder = append(outOrder, col)
		}
	}
	return out, outOrder
}

func droppedColumnOrder(drop []string, columnOrder []string) []string {
	if len(drop) == 0 {
		return columnOrder
	}
	dropSet := make(map[string]bool, len(drop))
	for _, col := range drop {
		dropSet[col] = true
	}
	out := make([]string, 0, len(columnOrder))
	for _, col := range columnOrder {
		if !dropSet[col] {
			out = append(out, col)
		}
	}
	return out
}

func applyRenameColumn(rename map[string]string, row map[string]string, columnOrder []string) (map[string]string, []string) {
	if len(rename) == 0 {
		return row, columnOrder
	}

	out := make(map[string]string, len(row))
	outOrder := make([]string, 0, len(columnOrder))
	for _, src := range columnOrder {
		val, ok := row[src]
		if !ok {
			continue
		}
		tgt := src
		if renamed, ok := rename[src]; ok {
			tgt = renamed
		}
		out[tgt] = val
		outOrder = append(outOrder, tgt)
	}
	return out, outOrder
}

func renamedColumnOrder(rename map[string]string, columnOrder []string) []string {
	if len(rename) == 0 {
		return columnOrder
	}
	out := make([]string, 0, len(columnOrder))
	for _, src := range columnOrder {
		tgt := src
		if renamed, ok := rename[src]; ok {
			tgt = renamed
		}
		out = append(out, tgt)
	}
	return out
}

func (api *EtlApi) Transform(jobID string) ([]DataRow, []string) {
	if jobID == "" {
		jobID = config.DefaultJobID
	}
	rootCfg := config.GetConfig()
	jobCfg := rootCfg.EffectiveConfigForJob(jobID)
	if jobCfg == nil {
		return nil, nil
	}
	st := api.state(jobID)
	steps := jobCfg.TransformSteps()
	ensureDateLayouts(jobCfg, st, st.dataRows)
	return transformAllRows(jobCfg, st.dataRows, st.columnOrder, steps, 0, st.dateLayouts)
}

func applyReplacementsStep(replacements *config.ReplacementsConfig, row *DataRow, order *[]string) {
	if replacements == nil {
		return
	}
	source := replacements.SourceColumnOrDefault()
	target := replacements.TargetColumnOrDefault()
	row.Set(target, findCategory(replacements, row.Get(source)))
	*order = ensureColumnInOrder(*order, target)
}

func applyAddCategoryStep(cfg *config.Config, addCategory *config.AddCategoryConfig, row *DataRow, order *[]string) {
	if addCategory == nil {
		return
	}

	resolved, err := addCategory.Resolve(config.ConfigDirectory())
	if err != nil {
		log.Warnf("add_category: %v", err)
		return
	}
	if resolved == nil || resolved.SourceColumn == "" || !resolved.HasMappings() {
		return
	}

	key := strings.TrimSpace(row.Get(resolved.SourceColumn))
	if key == "" {
		return
	}

	if v, ok := resolved.Values[key]; ok {
		row.Set(resolved.TargetColumn, v)
		*order = ensureColumnInOrder(*order, resolved.TargetColumn)
		return
	}

	for pattern, v := range resolved.Regex {
		if match, _ := regexp.MatchString(pattern, key); match {
			row.Set(resolved.TargetColumn, v)
			*order = ensureColumnInOrder(*order, resolved.TargetColumn)
			return
		}
	}
}

func findCategory(replacements *config.ReplacementsConfig, description string) string {
	if replacements == nil {
		return ""
	}
	if val, ok := replacements.Exact[description]; ok {
		return val
	}

	for pattern, val := range replacements.Regex {
		if match, _ := regexp.MatchString(pattern, description); match {
			return val
		}
	}

	return ""
}
