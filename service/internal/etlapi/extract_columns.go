package api

import (
	"sort"

	pb "github.com/jamesread/data-cleaner/gen/data_cleaner/api/v1"
	"github.com/jamesread/data-cleaner/internal/config"
)

func appendExtractColumnsSnapshot(res *pb.PreviewResponse, columnOrder []string, cfg *config.Config) {
	if res == nil {
		return
	}

	if len(columnOrder) == 0 && cfg != nil && cfg.Extract != nil && len(cfg.Extract.Columns) > 0 {
		columnOrder = config.SortedExtractColumnNames(cfg.Extract.Columns)
	}

	for _, name := range columnOrder {
		idx := int32(-1)
		if cfg != nil && cfg.Extract != nil {
			if i, ok := cfg.Extract.Columns[name]; ok {
				idx = int32(i)
			}
		}
		res.ExtractColumns = append(res.ExtractColumns, &pb.ExtractColumnEntry{
			FieldName:   name,
			ColumnIndex: idx,
		})
	}
}

func appendExtractPreviewRows(res *pb.PreviewResponse, rows []Row) {
	if res == nil || len(rows) == 0 {
		return
	}

	limit := len(rows)
	if limit > DefaultPreviewRowLimit {
		limit = DefaultPreviewRowLimit
	}

	for _, row := range rows[:limit] {
		cells := extractPreviewCells(res.ExtractColumns, row)
		res.ExtractPreviewRows = append(res.ExtractPreviewRows, &pb.ImportedLinePreview{
			Cells:            cells,
			SourceFilename:   rowFilenameBase(row.Filename),
			SourceLineNumber: row.LineNumber,
		})
	}
}

func extractPreviewCells(columns []*pb.ExtractColumnEntry, row Row) []string {
	cells := make([]string, 0, len(columns))
	for _, col := range columns {
		cells = append(cells, row.Columns[col.GetFieldName()])
	}
	return cells
}

func appendOutputColumnsSnapshot(res *pb.PreviewResponse, columns []string, steps config.TransformPipeline) {
	if res == nil || len(columns) == 0 {
		return
	}
	rename := renameMapThroughSteps(steps)
	for _, name := range columns {
		entry := &pb.ColumnMapEntry{LoadColumn: name}
		for src, tgt := range rename {
			if tgt == name {
				entry.SourceColumn = src
				break
			}
		}
		if entry.SourceColumn == "" {
			entry.SourceColumn = name
		}
		res.ColumnMap = append(res.ColumnMap, entry)
	}
}

func appendPreviewRows(res *pb.PreviewResponse, cfg *config.Config, rows []Row, columnOrder []string, dateLayouts map[string]string, maxRows int, stepOrdinal int32) {
	if len(rows) == 0 {
		return
	}
	displayLimit := DefaultPreviewRowLimit
	if maxRows > 0 {
		displayLimit = maxRows
	}

	steps := cfg.TransformSteps()
	activeSteps := config.StepsThroughOrdinal(steps, stepOrdinal)
	columns := outputColumnOrder(columnOrder, activeSteps)
	appendOutputColumnsSnapshot(res, columns, activeSteps)

	transformed, _ := transformAllRows(cfg, rows, columnOrder, steps, stepOrdinal, dateLayouts)
	limit := len(transformed)
	if displayLimit < limit {
		limit = displayLimit
	}

	for i := 0; i < limit; i++ {
		dr := transformed[i]
		row := rows[i]
		cells := make([]string, 0, len(columns))
		for _, col := range columns {
			cells = append(cells, dr.Get(col))
		}
		res.PreviewRows = append(res.PreviewRows, &pb.ImportedLinePreview{
			Cells:            cells,
			SourceFilename:   rowFilenameBase(row.Filename),
			SourceLineNumber: row.LineNumber,
		})
	}
}

func sortedDropColumns(cols []string) []string {
	out := append([]string(nil), cols...)
	sort.Strings(out)
	return out
}

func sortedRenameKeys(rename map[string]string) []string {
	keys := make([]string, 0, len(rename))
	for k := range rename {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
