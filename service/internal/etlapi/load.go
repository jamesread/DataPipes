package api

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	pb "github.com/jamesread/data-cleaner/gen/data_cleaner/api/v1"
	"github.com/jamesread/data-cleaner/internal/config"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	log "github.com/sirupsen/logrus"
)

type RowLoadReporter func(rowNum int, row DataRow, err error) error

type Connector interface {
	Connect() error
	Load(dataRows []DataRow, columns []string, report RowLoadReporter) error
}

type MySQLConnector struct {
	Properties map[string]string
	truncate   bool

	conn *sql.DB
}

func (c *MySQLConnector) Connect() error {
	var err error

	c.conn, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", c.Properties["user"], c.Properties["pass"], c.Properties["host"], c.Properties["port"], c.Properties["database"]))

	if err != nil {
		log.Errorf("Failed to connect to MySQL: %v", err)
	}

	return err
}

func (c *MySQLConnector) Load(dataRows []DataRow, columns []string, report RowLoadReporter) error {
	tableName := c.Properties["table"]
	if c.truncate {
		if err := c.truncateTable(tableName); err != nil {
			return err
		}
	}

	if len(columns) == 0 {
		return fmt.Errorf("no load columns defined")
	}

	insertSQL := buildInsertSQL(tableName, columns)
	log.Infof("Executing SQL: %v", insertSQL)

	stmt, err := c.conn.Prepare(insertSQL)
	if err != nil {
		log.Errorf("Failed to prepare statement: %v", err)
		return err
	}

	var firstErr error
	for i, row := range dataRows {
		args := rowExecArgs(row, columns)
		_, err = stmt.Exec(args...)
		if report != nil {
			if repErr := report(i+1, row, err); repErr != nil {
				return repErr
			}
		}
		if err != nil {
			log.Errorf("Failed to execute row %d: %v", i+1, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		log.Infof("Loaded row %d", i+1)
	}

	_ = firstErr
	return nil
}

func (c *MySQLConnector) truncateTable(tableName string) error {
	stmt, err := c.conn.Prepare("TRUNCATE TABLE " + tableName)
	if err != nil {
		return fmt.Errorf("failed to prepare truncate statement for table %s: %v", tableName, err)
	}
	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("failed to truncate table %s: %v", tableName, err)
	}
	log.Infof("Truncated table %s before load", tableName)
	return nil
}

func buildInsertSQL(tableName string, columns []string) string {
	placeholders := strings.TrimSuffix(strings.Repeat("?, ", len(columns)), ", ")
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(columns, ", "), placeholders)
}

func rowExecArgs(row DataRow, columns []string) []any {
	args := make([]any, len(columns))
	for i, col := range columns {
		args[i] = row.Get(col)
	}
	return args
}

func (api *EtlApi) Load(jobID string) *pb.LoadResponse {
	res := &pb.LoadResponse{}
	if jobID == "" {
		jobID = config.DefaultJobID
	}

	log.Infof("Full extract before load for job %q", jobID)
	api.extractJob(jobID, 0, 0)

	if _, _, err := api.loadExtractedWithProgress(context.Background(), jobID, nil); err != nil {
		log.Errorf("Failed to load data for job %q: %v", jobID, err)
	}

	return res
}

func initConnector(conn *config.Connection) Connector {
	if conn == nil {
		return nil
	}
	props := conn.Properties()

	switch props["type"] {
	case config.ConnectionTypeMySQL:
		return NewMySQLConnector(conn)
	case config.ConnectionTypeFireflyIII:
		return NewFireflyIIIConnector(conn)
	case config.ConnectionTypeDownloadCSV:
		return nil
	default:
		log.Warnf("Unknown connector type: %s", props["type"])
		return nil
	}
}

func NewMySQLConnector(conn *config.Connection) *MySQLConnector {
	return &MySQLConnector{
		Properties: conn.Properties(),
		truncate:   conn != nil && conn.Truncate,
	}
}

func isDownloadCSVLoad(conn *config.Connection) bool {
	return conn != nil && conn.Type == config.ConnectionTypeDownloadCSV
}
