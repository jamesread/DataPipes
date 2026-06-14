package api

import (
	pb "github.com/jamesread/data-cleaner/gen/data_cleaner/api/v1"
	"github.com/jamesread/data-cleaner/internal/config"
)

func (api *EtlApi) ListJobs() *pb.ListJobsResponse {
	cfg := config.GetConfig()
	configPath, configErr := config.LoadStatus()
	res := &pb.ListJobsResponse{
		ConfigPath: configPath,
		Jobs:       make([]*pb.JobSummary, 0),
	}
	if configErr != nil {
		res.ConfigError = configErr.Error()
		return res
	}

	res.Jobs = make([]*pb.JobSummary, 0, len(cfg.JobNames()))

	for _, id := range cfg.JobNames() {
		job := cfg.Job(id)
		if job == nil {
			continue
		}
		eff := cfg.EffectiveConfigForJob(id)
		summary := &pb.JobSummary{
			Id:               id,
			ExtractConnection: cfg.ExtractConnectionName(id),
			LoadConnection:   cfg.LoadConnectionName(id),
		}
		if eff != nil && eff.Extract != nil {
			summary.ImportDirectory = eff.Extract.ImportDirectory
		}
		if eff != nil && eff.Load != nil {
			summary.LoadConfigured = eff.Load.Destination != ""
		}
		if eff != nil {
			summary.Transformations = JobTransformationsFromConfig(eff)
		}
		res.Jobs = append(res.Jobs, summary)
	}

	return res
}
