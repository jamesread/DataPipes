package api

type jobState struct {
	globalIndex int
	columnOrder []string
	dateLayouts map[string]string
	dataRows    []Row
}

type EtlApi struct {
	jobs map[string]*jobState
}

func NewEtlApi() *EtlApi {
	return &EtlApi{
		jobs: make(map[string]*jobState),
	}
}

func (api *EtlApi) state(jobID string) *jobState {
	st, ok := api.jobs[jobID]
	if !ok {
		st = &jobState{
			dataRows: make([]Row, 0),
		}
		api.jobs[jobID] = st
	}
	return st
}
