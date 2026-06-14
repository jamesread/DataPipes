package api

import (
	"bytes"
	"encoding/csv"

	"github.com/jamesread/data-cleaner/internal/config"
	log "github.com/sirupsen/logrus"
)

func (api *EtlApi) ExportCSV(jobID string) []byte {
	if jobID == "" {
		jobID = config.DefaultJobID
	}

	st := api.state(jobID)
	if len(st.dataRows) == 0 {
		api.extractJob(jobID, 0, 0)
	}

	dataRows, columns := api.Transform(jobID)
	log.Infof("Exporting CSV for job %q, rows: %v", jobID, len(dataRows))

	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)

	if len(columns) > 0 {
		if err := writer.Write(columns); err != nil {
			log.Errorf("Error writing header to CSV: %v", err)
		}
	}

	for _, row := range dataRows {
		err := writer.Write(row.ToSlice(columns))
		if err != nil {
			log.Errorf("Error writing row to CSV: %v", err)
			continue
		}
	}

	writer.Flush()

	return buf.Bytes()
}
