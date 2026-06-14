package api

import (
	"context"
	"fmt"

	pb "github.com/jamesread/data-cleaner/gen/data_cleaner/api/v1"
	"github.com/jamesread/data-cleaner/internal/config"
	log "github.com/sirupsen/logrus"
)

func (api *EtlApi) FullRun(jobID string) *pb.FullRunResponse {
	if jobID == "" {
		jobID = config.DefaultJobID
	}
	log.Infof("Starting full run for job %q", jobID)

	preview := api.extractJob(jobID, 0, 0)
	res := &pb.FullRunResponse{
		Preview: preview,
	}

	rootCfg := config.GetConfig()
	jobCfg := rootCfg.EffectiveConfigForJob(jobID)
	if jobCfg == nil || jobCfg.Load == nil || jobCfg.Load.Destination == "" {
		return res
	}
	if len(preview.Issues) > 0 {
		res.LoadError = "Load skipped: resolve preview issues first"
		return res
	}

	res.LoadAttempted = true
	succeeded, failed, err := api.loadExtractedWithProgress(context.Background(), jobID, nil)
	if err != nil {
		res.LoadError = err.Error()
		log.Errorf("Full run load failed for job %q: %v", jobID, err)
		return res
	}
	if failed > 0 {
		res.LoadError = pbProgressSummary(succeeded, failed)
		return res
	}

	res.LoadSucceeded = true
	return res
}

func pbProgressSummary(succeeded, failed int32) string {
	return fmt.Sprintf("Load finished with %d failed row(s) (%d succeeded)", failed, succeeded)
}

func errJobNotFound(jobID string) error {
	return &loadError{msg: "Unknown job: " + jobID}
}

func errLoadNotConfigured(jobID string) error {
	return &loadError{msg: "Load not configured for job: " + jobID}
}

var errUnknownConnector = &loadError{msg: "Unknown or missing load connection"}

type loadError struct {
	msg string
}

func (e *loadError) Error() string {
	return e.msg
}
