package api

import (
	pb "github.com/jamesread/data-cleaner/gen/data_cleaner/api/v1"
	"github.com/jamesread/data-cleaner/internal/buildinfo"
	"github.com/jamesread/data-cleaner/internal/config"
)

func (api *EtlApi) Init() *pb.InitResponse {
	config.GetConfig()
	configPath, loadErr := config.LoadStatus()

	res := &pb.InitResponse{
		Version:    buildinfo.Version,
		ConfigPath: configPath,
		Ok:         true,
	}
	if loadErr != nil {
		res.Ok = false
		res.Errors = []string{loadErr.Error()}
	}
	return res
}
