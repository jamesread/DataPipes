package api

import (
	pb "github.com/jamesread/data-cleaner/gen/data_cleaner/api/v1"
)

func (api *EtlApi) Import(jobID string) *pb.ImportResponse {
	return previewToImport(api.extractJob(jobID, 0, 0))
}
