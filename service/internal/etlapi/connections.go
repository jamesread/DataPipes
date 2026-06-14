package api

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	pb "github.com/jamesread/data-cleaner/gen/data_cleaner/api/v1"
	"github.com/jamesread/data-cleaner/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

func (api *EtlApi) ListConnections() *pb.ListConnectionsResponse {
	cfg := config.GetConfig()
	configPath, configErr := config.LoadStatus()
	res := &pb.ListConnectionsResponse{
		ConfigPath:  configPath,
		Connections: make([]*pb.ConnectionSummary, 0),
	}
	if configErr != nil {
		res.ConfigError = configErr.Error()
		return res
	}

	res.Connections = append(res.Connections, buildConnectionSummary(config.ImplicitDownloadCSVConnectionID, &config.Connection{
		Type: config.ConnectionTypeDownloadCSV,
	}))

	for _, id := range cfg.ConnectionNames() {
		conn := cfg.Connections[id]
		if conn == nil {
			continue
		}
		res.Connections = append(res.Connections, buildConnectionSummary(id, conn))
	}

	return res
}

func (api *EtlApi) GetConnection(id string) *pb.GetConnectionResponse {
	cfg := config.GetConfig()
	configPath, configErr := config.LoadStatus()
	res := &pb.GetConnectionResponse{
		ConfigPath: configPath,
	}
	if configErr != nil {
		res.ConfigError = configErr.Error()
		return res
	}

	conn := cfg.ResolveConnection(id)
	if conn == nil {
		res.Error = "connection not found: " + id
		return res
	}

	res.Connection = buildConnectionSummary(id, conn)
	return res
}

func buildConnectionSummary(id string, conn *config.Connection) *pb.ConnectionSummary {
	summary := &pb.ConnectionSummary{
		Id:   id,
		Type: conn.Type,
	}
	switch conn.Type {
	case config.ConnectionTypeCSV:
		summary.ImportDirectory = conn.ImportDirectory
	case config.ConnectionTypeMySQL:
		summary.Host = conn.Host
		summary.Port = conn.Port
		summary.User = conn.User
		summary.Database = conn.Database
		summary.Table = conn.Table
	case config.ConnectionTypeFireflyIII:
		summary.Host = conn.URL
		summary.User = conn.SourceAccount
		summary.Database = conn.TransactionTypeHintColumn()
	case config.ConnectionTypeDownloadCSV:
		summary.ImportDirectory = "Implicit — download transformed CSV via job load step"
	}
	ok, message := healthCheckConnection(conn)
	summary.HealthOk = ok
	summary.HealthMessage = message
	return summary
}

func healthCheckConnection(conn *config.Connection) (bool, string) {
	switch conn.Type {
	case config.ConnectionTypeCSV:
		return healthCheckCSV(conn)
	case config.ConnectionTypeMySQL:
		return healthCheckMySQL(conn)
	case config.ConnectionTypeFireflyIII:
		return healthCheckFireflyIII(conn)
	case config.ConnectionTypeDownloadCSV:
		return true, "ready"
	default:
		return false, "unknown connection type: " + conn.Type
	}
}

func healthCheckCSV(conn *config.Connection) (bool, string) {
	if conn.ImportDirectory == "" {
		return false, "import_directory not configured"
	}
	info, err := os.Stat(conn.ImportDirectory)
	if err != nil {
		return false, err.Error()
	}
	if !info.IsDir() {
		return false, "import_directory is not a directory"
	}
	return true, "directory accessible"
}

func healthCheckMySQL(conn *config.Connection) (bool, string) {
	props := conn.Properties()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		props["user"], props["pass"], props["host"], props["port"], props["database"])

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return false, err.Error()
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var version string
	err = db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		return false, err.Error()
	}
	return true, version
}

func healthCheckFireflyIII(conn *config.Connection) (bool, string) {
	if err := conn.ValidateFireflyIII(); err != nil {
		return false, err.Error()
	}
	connector := NewFireflyIIIConnector(conn)
	if err := connector.Connect(); err != nil {
		return false, err.Error()
	}
	return true, "API reachable"
}
