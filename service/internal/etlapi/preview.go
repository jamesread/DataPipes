package api

import (
	"os"
	"sort"
	"strings"
	"time"

	pb "github.com/jamesread/data-cleaner/gen/data_cleaner/api/v1"
	"github.com/jamesread/data-cleaner/internal/config"
	log "github.com/sirupsen/logrus"
)

const DefaultPreviewRowLimit = 10

func previewRowLimit(requestLimit int32) int {
	if requestLimit > 0 {
		return int(requestLimit)
	}
	return DefaultPreviewRowLimit
}

func (api *EtlApi) Preview(jobID string, requestLimit int32, stepOrdinal int32) *pb.PreviewResponse {
	limit := previewRowLimit(requestLimit)
	res := api.extractJob(jobID, limit, stepOrdinal)
	res.RowLimit = int32(limit)
	if stepOrdinal > 0 {
		res.AppliedStepOrdinal = stepOrdinal
	}
	return res
}

// extractJob reads CSV extract data into job state. maxRows 0 means no limit.
// stepOrdinal 0 runs all configured transformation phases.
func (api *EtlApi) extractJob(jobID string, maxRows int, stepOrdinal int32) *pb.PreviewResponse {
	if jobID == "" {
		jobID = config.DefaultJobID
	}

	rootCfg := config.GetConfig()
	jobCfg := rootCfg.EffectiveConfigForJob(jobID)
	if jobCfg == nil {
		return &pb.PreviewResponse{
			JobId: jobID,
			Issues: []*pb.Issue{{
				Description: "Unknown job: " + jobID,
			}},
		}
	}

	res := &pb.PreviewResponse{
		JobId:  jobID,
		Issues: make([]*pb.Issue, 0),
	}

	st := api.state(jobID)
	st.dataRows = make([]Row, 0)
	st.globalIndex = 0
	st.columnOrder = nil
	st.dateLayouts = nil

	if jobCfg.Extract == nil {
		res.Issues = append(res.Issues, &pb.Issue{
			Description: "Job has no CSV extract connection configured (jobs.<name>.extract → connections.<name> with type: csv)",
		})
		return res
	}

	dir := "/opt/import/"
	if jobCfg.Extract.ImportDirectory != "" {
		dir = jobCfg.Extract.ImportDirectory
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Warnf("failed to read directory: %v", err)
		res.Issues = append(res.Issues, &pb.Issue{
			Description: "Failed to read directory: " + err.Error(),
		})
	}

	truncated := false
	for i, entry := range entries {
		if maxRows > 0 && len(st.dataRows) >= maxRows {
			truncated = true
			break
		}
		if entry.Name() == "Export.csv" {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".csv") {
			continue
		}
		sourceFile, fileTruncated := api.parseFile(jobCfg, st, dir, entry.Name(), maxRows)
		res.SourceFiles = append(res.SourceFiles, sourceFile)
		if maxRows > 0 && len(st.dataRows) > maxRows {
			st.dataRows = st.dataRows[:maxRows]
		}
		if maxRows > 0 && len(st.dataRows) >= maxRows {
			truncated = fileTruncated || hasRemainingCSVFiles(entries, i+1)
			break
		}
	}

	sort.Sort(ByGlobalIndex(st.dataRows))
	ensureDateLayouts(jobCfg, st, st.dataRows)
	appendExtractColumnsSnapshot(res, st.columnOrder, jobCfg)
	appendExtractPreviewRows(res, st.dataRows)
	appendPreviewRows(res, jobCfg, st.dataRows, st.columnOrder, st.dateLayouts, maxRows, stepOrdinal)
	applyValidationsThroughStep(jobCfg, st.dataRows, st.columnOrder, res, stepOrdinal)

	res.TotalLines = int64(len(st.dataRows))
	res.CompletedDate = time.Now().Format(time.RFC3339)
	res.Transformations = JobTransformationsFromConfig(jobCfg)
	if maxRows == 0 && len(st.dataRows) > DefaultPreviewRowLimit {
		res.Truncated = true
	} else {
		res.Truncated = truncated
	}

	return res
}

func hasRemainingCSVFiles(entries []os.DirEntry, from int) bool {
	for _, entry := range entries[from:] {
		if entry.Name() == "Export.csv" || !strings.HasSuffix(entry.Name(), ".csv") {
			continue
		}
		return true
	}
	return false
}

func applyValidationsThroughStep(cfg *config.Config, rows []Row, columnOrder []string, res *pb.PreviewResponse, stepOrdinal int32) {
	for _, step := range config.StepsThroughOrdinal(cfg.TransformSteps(), stepOrdinal) {
		if step.RollingTotal != nil {
			applyRollingTotalCheck(step.RollingTotal, rows, columnOrder, res)
		}
	}
}

func rowFilenameBase(filename string) string {
	if i := strings.LastIndex(filename, "/"); i >= 0 {
		return filename[i+1:]
	}
	return filename
}

func previewToImport(res *pb.PreviewResponse) *pb.ImportResponse {
	if res == nil {
		return nil
	}
	out := &pb.ImportResponse{
		Issues:             res.Issues,
		CompletedDate:      res.CompletedDate,
		TotalLines:         res.TotalLines,
		SourceFiles:        res.SourceFiles,
		Transformations:    res.Transformations,
		ColumnMap:          res.ColumnMap,
		ImportPreview:      res.PreviewRows,
		JobId:              res.JobId,
		ExtractColumns:     res.ExtractColumns,
		ExtractPreviewRows: res.ExtractPreviewRows,
	}
	return out
}
