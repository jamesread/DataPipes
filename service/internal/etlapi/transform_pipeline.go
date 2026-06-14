package api

import (
	"github.com/jamesread/data-cleaner/internal/config"
)

func transformAllRows(cfg *config.Config, rows []Row, columnOrder []string, steps config.TransformPipeline, stepOrdinal int32, dateLayouts map[string]string) ([]DataRow, []string) {
	if len(rows) == 0 {
		return nil, nil
	}

	activeSteps := config.StepsThroughOrdinal(steps, stepOrdinal)
	out := make([]DataRow, len(rows))
	sourceRows := make([]map[string]string, len(rows))
	for i, rec := range rows {
		sourceRows[i] = copyColumnValues(rec)
		out[i] = DataRow{contents: copyColumnValues(rec)}
	}

	order := append([]string(nil), columnOrder...)
	for _, step := range activeSteps {
		switch step.Kind() {
		case "replacements":
			for i := range out {
				applyReplacementsStep(step.Replacements, &out[i], &order)
			}
		case "add_category":
			for i := range out {
				applyAddCategoryStep(cfg, step.AddCategory, &out[i], &order)
			}
		case "date_normalize":
			for i := range out {
				applyDateNormalizeStep(step.DateNormalize, &out[i], dateLayouts)
			}
		case "date_to_incremental":
			applyDateToIncrementalStep(step.DateToIncremental, out)
		case "drop_column":
			for i := range out {
				out[i].contents, order = applyDropColumn(step.DropColumn, out[i].contents, order)
			}
		case "rename_column":
			for i := range out {
				out[i].contents, order = applyRenameColumn(step.RenameColumn, out[i].contents, order)
			}
		case "append_hash":
			for i := range out {
				applyAppendHashStep(step.AppendHash, &out[i], sourceRows[i], columnOrder)
			}
		case "rolling_total":
			// validation only
		}
	}

	return out, outputColumnOrder(order, activeSteps)
}

func outputColumnOrder(baseOrder []string, steps config.TransformPipeline) []string {
	order := append([]string(nil), baseOrder...)
	for _, step := range steps {
		switch step.Kind() {
		case "replacements":
			if step.Replacements != nil {
				order = ensureColumnInOrder(order, step.Replacements.TargetColumnOrDefault())
			}
		case "add_category":
			if step.AddCategory != nil && step.AddCategory.TargetColumn != "" {
				order = ensureColumnInOrder(order, step.AddCategory.TargetColumn)
			}
		case "drop_column":
			order = droppedColumnOrder(step.DropColumn, order)
		case "rename_column":
			order = renamedColumnOrder(step.RenameColumn, order)
		}
	}
	return order
}

func renameMapThroughSteps(steps config.TransformPipeline) map[string]string {
	rename := make(map[string]string)
	for _, step := range steps {
		if step.Kind() != "rename_column" {
			continue
		}
		for src, tgt := range step.RenameColumn {
			rename[src] = tgt
		}
	}
	return rename
}
