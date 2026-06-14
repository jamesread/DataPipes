package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	pb "github.com/jamesread/data-cleaner/gen/data_cleaner/api/v1"
	dcapiv1connect "github.com/jamesread/data-cleaner/gen/data_cleaner/api/v1/dcapiv1connect"
	etlapi "github.com/jamesread/data-cleaner/internal/etlapi"
	"github.com/jamesread/data-cleaner/internal/config"
	"connectrpc.com/connect"
)

type Server struct {
	etl *etlapi.EtlApi
}

func NewServer() *Server {
	return &Server{
		etl: etlapi.NewEtlApi(),
	}
}

func (s *Server) Init(ctx context.Context, in *connect.Request[pb.InitRequest]) (*connect.Response[pb.InitResponse], error) {
	res := s.etl.Init()
	return connect.NewResponse(res), nil
}

func (s *Server) ListJobs(ctx context.Context, in *connect.Request[pb.ListJobsRequest]) (*connect.Response[pb.ListJobsResponse], error) {
	res := s.etl.ListJobs()
	return connect.NewResponse(res), nil
}

func (s *Server) ListConnections(ctx context.Context, in *connect.Request[pb.ListConnectionsRequest]) (*connect.Response[pb.ListConnectionsResponse], error) {
	res := s.etl.ListConnections()
	return connect.NewResponse(res), nil
}

func (s *Server) ListTransformationTypes(ctx context.Context, in *connect.Request[pb.ListTransformationTypesRequest]) (*connect.Response[pb.ListTransformationTypesResponse], error) {
	res := s.etl.ListTransformationTypes()
	return connect.NewResponse(res), nil
}

func (s *Server) GetConnection(ctx context.Context, in *connect.Request[pb.GetConnectionRequest]) (*connect.Response[pb.GetConnectionResponse], error) {
	res := s.etl.GetConnection(in.Msg.Id)
	return connect.NewResponse(res), nil
}

func (s *Server) Preview(ctx context.Context, in *connect.Request[pb.PreviewRequest]) (*connect.Response[pb.PreviewResponse], error) {
	res := s.etl.Preview(in.Msg.JobId, in.Msg.RowLimit, in.Msg.StepOrdinal)
	return connect.NewResponse(res), nil
}

func (s *Server) FullRun(ctx context.Context, in *connect.Request[pb.FullRunRequest]) (*connect.Response[pb.FullRunResponse], error) {
	res := s.etl.FullRun(in.Msg.JobId)
	return connect.NewResponse(res), nil
}

func (s *Server) Import(ctx context.Context, in *connect.Request[pb.ImportRequest]) (*connect.Response[pb.ImportResponse], error) {
	res := s.etl.Import(in.Msg.JobId)
	return connect.NewResponse(res), nil
}

func (s *Server) Export(ctx context.Context, in *connect.Request[pb.ExportRequest]) (*connect.Response[pb.ExportResponse], error) {
	if in.Msg.RunImport {
		s.etl.Import("")
	}
	return connect.NewResponse(&pb.ExportResponse{}), nil
}

func (s *Server) Reload(ctx context.Context, in *connect.Request[pb.ReloadRequest]) (*connect.Response[pb.ReloadResponse], error) {
	config.ReloadConfig()
	return connect.NewResponse(&pb.ReloadResponse{}), nil
}

func (s *Server) Load(ctx context.Context, in *connect.Request[pb.LoadRequest]) (*connect.Response[pb.LoadResponse], error) {
	s.etl.Load(in.Msg.JobId)
	return connect.NewResponse(&pb.LoadResponse{}), nil
}

func (s *Server) StreamLoad(ctx context.Context, in *connect.Request[pb.LoadRequest], stream *connect.ServerStream[pb.LoadProgress]) error {
	return s.etl.StreamLoad(ctx, in.Msg.JobId, func(p *pb.LoadProgress) error {
		return stream.Send(p)
	})
}

func (s *Server) ConnectHandler() (string, http.Handler) {
	return dcapiv1connect.NewDataCleanerServiceHandler(s)
}

func (s *Server) DownloadCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/download/")
	jobID = strings.TrimSuffix(jobID, ".csv")
	jobID = strings.Trim(jobID, "/")
	if jobID == "" {
		http.Error(w, "job id required", http.StatusBadRequest)
		return
	}

	data := s.etl.ExportCSV(jobID)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.csv"`, jobID))
	_, _ = w.Write(data)
}
