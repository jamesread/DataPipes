package api

import (
	"context"
	"fmt"

	pb "github.com/jamesread/data-cleaner/gen/data_cleaner/api/v1"
	"github.com/jamesread/data-cleaner/internal/config"
	log "github.com/sirupsen/logrus"
)

type LoadProgressReporter func(*pb.LoadProgress) error

func loadStatsFromRow(row DataRow, columns []string) []*pb.LoadStat {
	prefer := []string{"date", "description", "memo", "amount", "value", "category", "type", "destination_name"}
	seen := make(map[string]bool, len(prefer))
	var stats []*pb.LoadStat
	add := func(key, value string) {
		if value == "" || seen[key] {
			return
		}
		seen[key] = true
		stats = append(stats, &pb.LoadStat{Key: key, Value: value})
	}
	for _, key := range prefer {
		add(key, row.Get(key))
	}
	for _, col := range columns {
		add(col, row.Get(col))
	}
	return stats
}

func (api *EtlApi) StreamLoad(ctx context.Context, jobID string, send LoadProgressReporter) error {
	if send == nil {
		send = func(*pb.LoadProgress) error { return nil }
	}
	if jobID == "" {
		jobID = config.DefaultJobID
	}

	log.Infof("Stream load for job %q", jobID)
	api.extractJob(jobID, 0, 0)

	succeeded, failed, err := api.loadExtractedWithProgress(ctx, jobID, send)
	if err != nil {
		_ = send(&pb.LoadProgress{
			Phase:   "failed",
			Message: err.Error(),
			Failed:  failed,
		})
		return err
	}

	msg := fmt.Sprintf("Load finished: %d succeeded", succeeded)
	if failed > 0 {
		msg = fmt.Sprintf("Load finished: %d succeeded, %d failed", succeeded, failed)
	}
	return send(&pb.LoadProgress{
		Phase:     "complete",
		Succeeded: succeeded,
		Failed:    failed,
		Message:   msg,
	})
}

func (api *EtlApi) loadExtractedWithProgress(ctx context.Context, jobID string, send LoadProgressReporter) (int32, int32, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	emit := func(p *pb.LoadProgress) error {
		if send == nil {
			return nil
		}
		return send(p)
	}
	if jobID == "" {
		jobID = config.DefaultJobID
	}

	rootCfg := config.GetConfig()
	jobCfg := rootCfg.EffectiveConfigForJob(jobID)
	if jobCfg == nil {
		return 0, 0, errJobNotFound(jobID)
	}

	ldconfig := jobCfg.Load
	if ldconfig == nil || ldconfig.Destination == "" {
		return 0, 0, errLoadNotConfigured(jobID)
	}

	conn := rootCfg.ResolveConnection(ldconfig.Destination)
	dataRows, columns := api.Transform(jobID)
	total := int32(len(dataRows))

	if isDownloadCSVLoad(conn) {
		if len(dataRows) == 0 {
			return 0, 0, errLoadNotConfigured(jobID)
		}
		if len(columns) == 0 {
			return 0, 0, errLoadNotConfigured(jobID)
		}
		if err := emit(&pb.LoadProgress{
			Phase:      "started",
			TotalRows:  1,
			Message:    "Prepared transformed data for download",
		}); err != nil {
			return 0, 0, err
		}
		if err := emit(&pb.LoadProgress{
			Phase:      "row",
			RowNumber:  1,
			TotalRows:  1,
			RowSuccess: true,
			Succeeded:  1,
			Stats:      loadStatsFromRow(dataRows[0], columns),
			Message:    "Ready for CSV download",
		}); err != nil {
			return 0, 0, err
		}
		return 1, 0, nil
	}

	connector := initConnector(conn)
	if connector == nil {
		return 0, 0, errUnknownConnector
	}
	if err := connector.Connect(); err != nil {
		return 0, 0, err
	}
	if len(columns) == 0 {
		return 0, 0, errLoadNotConfigured(jobID)
	}

	if err := emit(&pb.LoadProgress{
		Phase:     "started",
		TotalRows: total,
		Message:   fmt.Sprintf("Loading %d rows", total),
	}); err != nil {
		return 0, 0, err
	}

	var succeeded, failed int32
	err := connector.Load(dataRows, columns, func(rowNum int, row DataRow, rowErr error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		ev := &pb.LoadProgress{
			Phase:     "row",
			RowNumber: int32(rowNum),
			TotalRows: total,
			Stats:     loadStatsFromRow(row, columns),
		}
		if rowErr != nil {
			failed++
			ev.RowSuccess = false
			ev.Error = rowErr.Error()
			ev.Failed = failed
			ev.Succeeded = succeeded
			ev.Message = fmt.Sprintf("Row %d failed", rowNum)
		} else {
			succeeded++
			ev.RowSuccess = true
			ev.Succeeded = succeeded
			ev.Failed = failed
			ev.Message = fmt.Sprintf("Row %d loaded", rowNum)
		}
		return emit(ev)
	})
	if err != nil {
		return succeeded, failed, err
	}
	return succeeded, failed, nil
}
