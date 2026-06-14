package api

import (
	"encoding/csv"
	"os"
	"path"
	"strconv"
	"strings"

	pb "github.com/jamesread/data-cleaner/gen/data_cleaner/api/v1"
	"github.com/jamesread/data-cleaner/internal/config"
	log "github.com/sirupsen/logrus"
)

type Row struct {
	Columns    map[string]string
	Index      int
	LineNumber int64
	Filename   string
}

func csvUsesHeader(cfg *config.Config) bool {
	return cfg == nil || cfg.Csv == nil || cfg.Csv.Header
}

func resolveExtractSchema(cfg *config.Config, headerRow []string, width int) (order []string, indexByName map[string]int) {
	indexByName = make(map[string]int)

	if cfg != nil && cfg.Extract != nil && len(cfg.Extract.Columns) > 0 {
		order = config.SortedExtractColumnNames(cfg.Extract.Columns)
		for name, idx := range cfg.Extract.Columns {
			if idx < 0 {
				continue
			}
			indexByName[name] = idx
		}
		return order, indexByName
	}

	if csvUsesHeader(cfg) && len(headerRow) > 0 {
		order = make([]string, 0, len(headerRow))
		for i, h := range headerRow {
			name := strings.TrimSpace(h)
			if name == "" {
				name = strconv.Itoa(i)
			}
			order = append(order, name)
			indexByName[name] = i
		}
		return order, indexByName
	}

	if width <= 0 {
		return nil, indexByName
	}
	order = make([]string, width)
	for i := 0; i < width; i++ {
		name := strconv.Itoa(i)
		order[i] = name
		indexByName[name] = i
	}
	return order, indexByName
}

func (api *EtlApi) parseLines(cfg *config.Config, st *jobState, lines [][]string, filename string, maxRows int) bool {
	if len(lines) == 0 {
		return false
	}

	headerRow := []string(nil)
	dataStart := 0
	if csvUsesHeader(cfg) {
		headerRow = lines[0]
		dataStart = 1
	}

	width := 0
	for i := dataStart; i < len(lines); i++ {
		if len(lines[i]) > width {
			width = len(lines[i])
		}
	}

	columnOrder, indexByName := resolveExtractSchema(cfg, headerRow, width)
	if len(columnOrder) == 0 && width > 0 {
		columnOrder, indexByName = resolveExtractSchema(cfg, nil, width)
	}
	if len(st.columnOrder) == 0 && len(columnOrder) > 0 {
		st.columnOrder = append([]string(nil), columnOrder...)
	}

	for lineNumber := len(lines) - 1; lineNumber >= dataStart; lineNumber-- {
		if maxRows > 0 && len(st.dataRows) >= maxRows {
			return true
		}

		line := lines[lineNumber]
		rec := Row{
			Columns: make(map[string]string, len(columnOrder)),
		}

		for _, name := range columnOrder {
			idx, ok := indexByName[name]
			if !ok || idx < 0 || idx >= len(line) {
				rec.Columns[name] = ""
				continue
			}
			rec.Columns[name] = line[idx]
		}

		rec.Index = st.globalIndex
		rec.LineNumber = int64(lineNumber + 1)
		rec.Filename = filename

		st.dataRows = append(st.dataRows, rec)
		st.globalIndex++
	}
	return false
}

type ByGlobalIndex []Row

func (a ByGlobalIndex) Len() int           { return len(a) }
func (a ByGlobalIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByGlobalIndex) Less(i, j int) bool { return a[i].Index < a[j].Index }

func (api *EtlApi) parseFile(cfg *config.Config, st *jobState, directory string, filename string, maxRows int) (*pb.SourceFile, bool) {
	filepath := path.Join(directory, filename)

	log.Infof("Parsing file: %s", filepath)

	contents, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	csvReader := csv.NewReader(contents)
	lines, err := csvReader.ReadAll()
	if err != nil {
		log.Fatalf("failed to read csv: %v", err)
	}

	err = contents.Close()
	if err != nil {
		log.Fatalf("failed to close file: %v", err)
	}

	truncated := api.parseLines(cfg, st, lines, filepath, maxRows)

	return &pb.SourceFile{
		Filename:  filename,
		LineCount: int64(len(lines)),
	}, truncated
}
