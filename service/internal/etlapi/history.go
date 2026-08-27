package api

import (
	"fmt"
	"strings"
	"time"

	pb "github.com/jamesread/data-cleaner/gen/data_cleaner/api/v1"
	"github.com/jamesread/data-cleaner/internal/config"
)

func (api *EtlApi) ListJobExecutions(jobID string) *pb.ListJobExecutionsResponse {
	configPath, configErr := config.LoadStatus()
	res := &pb.ListJobExecutionsResponse{
		ConfigPath: configPath,
		Executions: []*pb.JobExecution{},
	}
	if configErr != nil {
		res.ConfigError = configErr.Error()
		return res
	}
	if jobID == "" {
		return res
	}

	entries, err := api.history.ListJob(jobID)
	if err != nil {
		res.ConfigError = err.Error()
		return res
	}
	for _, e := range entries {
		res.Executions = append(res.Executions, &pb.JobExecution{
			Id:          e.ID,
			JobId:       e.JobID,
			StartedAt:   e.StartedAt,
			CompletedAt: e.CompletedAt,
			RunKind:     e.RunKind,
			StepDetail:  e.StepDetail,
			Success:     e.Success,
			Message:     e.Message,
			TriggeredBy: e.TriggeredBy,
		})
	}
	return res
}

func (api *EtlApi) recordPreviewExecution(jobID string, stepOrdinal int32, started time.Time, res *pb.PreviewResponse) {
	if res == nil {
		return
	}
	runKind, stepDetail := previewRunMeta(jobID, stepOrdinal)
	success := len(res.Issues) == 0
	message := issuesSummary(res.Issues)
	_ = api.history.Record(jobID, runKind, stepDetail, started, success, message)
}

func (api *EtlApi) recordPipelineExecution(jobID string, started time.Time, res *pb.FullRunResponse) {
	if res == nil {
		return
	}
	success := len(res.Preview.GetIssues()) == 0
	message := issuesSummary(res.Preview.GetIssues())
	if res.LoadAttempted {
		if res.LoadSucceeded {
			if message != "" {
				message += "; "
			}
			message += "Load succeeded"
		} else {
			success = false
			if res.LoadError != "" {
				if message != "" {
					message += "; "
				}
				message += res.LoadError
			} else {
				message += "Load failed"
			}
		}
	} else if res.LoadError != "" {
		if message != "" {
			message += "; "
		}
		message += res.LoadError
		if strings.Contains(strings.ToLower(res.LoadError), "skipped") {
			success = false
		}
	}
	_ = api.history.Record(jobID, "pipeline", "Pipeline (extract, transform, load)", started, success, message)
}

func (api *EtlApi) recordLoadExecution(jobID string, started time.Time, succeeded, failed int32, err error) {
	success := err == nil && failed == 0
	message := ""
	if err != nil {
		message = err.Error()
	} else if failed > 0 {
		message = fmt.Sprintf("Load finished with %d failed row(s) (%d succeeded)", failed, succeeded)
	} else {
		message = fmt.Sprintf("Load finished: %d succeeded", succeeded)
	}
	_ = api.history.Record(jobID, "load", "Load only", started, success, message)
}

func previewRunMeta(jobID string, stepOrdinal int32) (runKind, stepDetail string) {
	switch {
	case stepOrdinal < 0:
		return "extract", "Extract only"
	case stepOrdinal > 0:
		return "step", stepDetailForOrdinal(jobID, stepOrdinal)
	default:
		return "preview", "Preview (all transformations)"
	}
}

func stepDetailForOrdinal(jobID string, ordinal int32) string {
	rootCfg := config.GetConfig()
	jobCfg := rootCfg.EffectiveConfigForJob(jobID)
	if jobCfg == nil {
		return fmt.Sprintf("Step %d", ordinal)
	}
	for _, t := range JobTransformationsFromConfig(jobCfg) {
		if t.Ordinal == ordinal {
			name := t.Name
			if name == "" {
				name = "transformation"
			}
			return fmt.Sprintf("Step %d: %s", ordinal, name)
		}
	}
	return fmt.Sprintf("Step %d", ordinal)
}

func issuesSummary(issues []*pb.Issue) string {
	if len(issues) == 0 {
		return ""
	}
	parts := make([]string, 0, len(issues))
	for _, issue := range issues {
		if issue.GetDescription() != "" {
			parts = append(parts, issue.GetDescription())
		}
	}
	return strings.Join(parts, "; ")
}
