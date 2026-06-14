package config

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	DefaultJobID       = "default"
	ConnectionTypeCSV         = "csv"
	ConnectionTypeMySQL       = "mysql"
	ConnectionTypeDownloadCSV = "download_csv"
	ConnectionTypeFireflyIII  = "firefly_iii"
)

// Connection is a named extract source or load destination.
// type: csv — CSV files from import_directory; type: mysql — SQL load target;
// type: download_csv — implicit load target (no config); CSV download via HTTP.
// type: firefly_iii — Firefly III API load target (creates transactions).
type Connection struct {
	Type string `yaml:"type"`

	ImportDirectory string           `yaml:"import_directory,omitempty"`
	Columns         ExtractColumnMap `yaml:"columns,omitempty"`

	Host     string `yaml:"host,omitempty"`
	Port     string `yaml:"port,omitempty"`
	User     string `yaml:"user,omitempty"`
	Pass     string `yaml:"pass,omitempty"`
	Database string `yaml:"database,omitempty"`
	Table    string `yaml:"table,omitempty"`
	Truncate bool   `yaml:"truncate,omitempty"`

	URL             string            `yaml:"url,omitempty"`
	Token           string            `yaml:"token,omitempty"`
	SourceAccount   string            `yaml:"source_account,omitempty"`
	TransactionType string            `yaml:"transaction_type,omitempty"` // pipeline column name for type hints (e.g. TFR → transfer)
	ApplyRules              *bool             `yaml:"apply_rules,omitempty"`
	LoadColumns             FireflyColumnMap  `yaml:"load_columns,omitempty"`
	DefaultExpenseAccount   string            `yaml:"default_expense_account,omitempty"`
	DefaultTransferAccount  string            `yaml:"default_transfer_account,omitempty"`
}

func (c *Connection) Properties() map[string]string {
	if c == nil {
		return nil
	}
	props := map[string]string{"type": c.Type}
	if c.Host != "" {
		props["host"] = c.Host
	}
	if c.Port != "" {
		props["port"] = c.Port
	}
	if c.User != "" {
		props["user"] = c.User
	}
	if c.Pass != "" {
		props["pass"] = c.Pass
	}
	if c.Database != "" {
		props["database"] = c.Database
	}
	if c.Table != "" {
		props["table"] = c.Table
	}
	if c.URL != "" {
		props["url"] = c.URL
	}
	if c.Token != "" {
		props["token"] = c.Token
	}
	if c.SourceAccount != "" {
		props["source_account"] = c.SourceAccount
	}
	if c.TransactionType != "" {
		props["transaction_type"] = c.TransactionType
	}
	return props
}

func (c *Connection) AsExtractConfig() *ExtractConfig {
	if c == nil || c.Type != ConnectionTypeCSV {
		return nil
	}
	return &ExtractConfig{
		ImportDirectory: c.ImportDirectory,
		Columns:         c.Columns,
	}
}

// JobConfig selects extract and load connections and defines per-job transform + column mapping.
type JobConfig struct {
	Extract   string           `yaml:"extract"`
	Load      string           `yaml:"load"`
	Transform TransformPipeline `yaml:"transform,omitempty"`
}

type Config struct {
	Csv *CsvConfig

	Connections map[string]*Connection `yaml:"connections"`
	Jobs        map[string]*JobConfig  `yaml:"jobs,omitempty"`

	// Legacy single-job config (top-level extract/transform/load).
	Extract   *ExtractConfig
	Transform TransformPipeline
	Load      *LoadConfig

	Network *NetworkConfig
}

type ExtractConfig struct {
	ImportDirectory string           `yaml:"import_directory,omitempty"`
	Columns         ExtractColumnMap `yaml:"columns,omitempty"`
}

type LoadConfig struct {
	Destination string
}

type AddCategoryConfig struct {
	SourceColumn  string            `yaml:"source_column"`
	TargetColumn  string            `yaml:"target_column,omitempty"`
	Values        map[string]string `yaml:"values,omitempty"`
	Regex         map[string]string `yaml:"regex,omitempty"`
	FromFile      string            `yaml:"from_file,omitempty"`
	FromFileCamel string            `yaml:"fromFile,omitempty"`
}

// RollingTotalConfig validates that previous balance + value equals current balance per row.
type RollingTotalConfig struct {
	ValueColumn   string   `yaml:"value_column,omitempty"`
	BalanceColumn string   `yaml:"balance_column,omitempty"`
	Tolerance     *float64 `yaml:"tolerance,omitempty"`
}

func (c *RollingTotalConfig) ValueColumnOrDefault() string {
	if c != nil && c.ValueColumn != "" {
		return c.ValueColumn
	}
	return "value"
}

func (c *RollingTotalConfig) BalanceColumnOrDefault() string {
	if c != nil && c.BalanceColumn != "" {
		return c.BalanceColumn
	}
	return "balance"
}

func (c *RollingTotalConfig) ToleranceOrDefault() float64 {
	if c != nil && c.Tolerance != nil {
		return *c.Tolerance
	}
	return 0.001
}


type CsvConfig struct {
	Header bool
}

type ReplacementsConfig struct {
	SourceColumn string            `yaml:"source_column,omitempty"`
	TargetColumn string            `yaml:"target_column,omitempty"`
	Exact        map[string]string `yaml:"exact,omitempty"`
	Regex        map[string]string `yaml:"regex,omitempty"`
}

func (r *ReplacementsConfig) SourceColumnOrDefault() string {
	if r != nil && r.SourceColumn != "" {
		return r.SourceColumn
	}
	return "description"
}

func (r *ReplacementsConfig) TargetColumnOrDefault() string {
	if r != nil && r.TargetColumn != "" {
		return r.TargetColumn
	}
	return "category"
}

type NetworkConfig struct {
	BindGrpc  string `yaml:"bindgrpc,omitempty"`
	BindRest  string `yaml:"bindrest,omitempty"`
	BindProxy string `yaml:"bindproxy,omitempty"`
}

var config *Config
var configPath string
var configLoadErr error

func GetConfig() *Config {
	if config == nil {
		config = ReloadConfig()
	}
	return config
}

// LoadStatus reports the config file path and any error from the last load attempt.
func LoadStatus() (path string, err error) {
	if config == nil {
		config = ReloadConfig()
	}
	return configPath, configLoadErr
}

func (c *Config) JobNames() []string {
	if c == nil {
		return []string{DefaultJobID}
	}
	if len(c.Jobs) > 0 {
		names := make([]string, 0, len(c.Jobs))
		for name := range c.Jobs {
			names = append(names, name)
		}
		sort.Strings(names)
		return names
	}
	if c.hasLegacyJob() {
		return []string{DefaultJobID}
	}
	return []string{DefaultJobID}
}

func (c *Config) ConnectionNames() []string {
	if c == nil || len(c.Connections) == 0 {
		return nil
	}
	names := make([]string, 0, len(c.Connections))
	for name := range c.Connections {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c *Config) hasLegacyJob() bool {
	return c.Extract != nil || c.Transform != nil || c.Load != nil
}

func (c *Config) Job(jobID string) *JobConfig {
	if c == nil {
		return nil
	}
	if jobID == "" {
		jobID = DefaultJobID
	}
	if len(c.Jobs) > 0 {
		return c.Jobs[jobID]
	}
	if jobID != DefaultJobID || !c.hasLegacyJob() {
		return nil
	}
	job := &JobConfig{Transform: c.Transform}
	if c.Load != nil {
		job.Load = c.Load.Destination
	}
	return job
}

// EffectiveConfigForJob resolves a job's extract/load connections into a runnable config.
func (c *Config) EffectiveConfigForJob(jobID string) *Config {
	if jobID == "" {
		jobID = DefaultJobID
	}

	if len(c.Jobs) == 0 && c.hasLegacyJob() && jobID == DefaultJobID {
		return &Config{
			Csv:         c.Csv,
			Connections: c.Connections,
			Extract:     c.Extract,
			Transform:   c.Transform,
			Load:        c.Load,
		}
	}

	job := c.Job(jobID)
	if job == nil {
		return nil
	}

	var extract *ExtractConfig
	if job.Extract != "" {
		if conn := c.Connections[job.Extract]; conn != nil {
			extract = conn.AsExtractConfig()
		}
	}

	var load *LoadConfig
	if job.Load != "" {
		load = &LoadConfig{
			Destination: job.Load,
		}
	}

	return &Config{
		Csv:         c.Csv,
		Connections: c.Connections,
		Extract:     extract,
		Transform:   job.Transform,
		Load:        load,
	}
}

func (c *Config) ExtractConnectionName(jobID string) string {
	job := c.Job(jobID)
	if job == nil {
		return ""
	}
	return job.Extract
}

func (c *Config) LoadConnectionName(jobID string) string {
	job := c.Job(jobID)
	if job == nil {
		return ""
	}
	return job.Load
}

func newDefaultConfig() *Config {
	return &Config{
		Connections: map[string]*Connection{},
		Network: &NetworkConfig{
			BindGrpc:  "127.0.0.1:50051",
			BindRest:  "127.0.0.1:8081",
			BindProxy: ":8080",
		},
		Csv: &CsvConfig{
			Header: true,
		},
		Transform: TransformPipeline{
			{
				Replacements: &ReplacementsConfig{
					Exact: map[string]string{},
					Regex: map[string]string{},
				},
			},
		},
	}
}

func ReloadConfig() *Config {
	config = newDefaultConfig()
	configLoadErr = nil

	envconf := os.Getenv("DATAPIPES_CONFIG")
	if envconf == "" {
		envconf = os.Getenv("DATACLEANER_CONFIG")
	}
	if envconf == "" {
		envconf = os.Getenv("DATA_CLEANER_CONFIG") // legacy
	}
	filename := ""
	if envconf != "" {
		filename = envconf
	} else {
		home := os.Getenv("HOME")
		filename = path.Join(home, ".datapipes-config.yaml")
		if _, err := os.Stat(filename); err != nil {
			legacy := path.Join(home, ".datacleaner-config.yaml")
			if _, err := os.Stat(legacy); err == nil {
				filename = legacy
			}
		}
	}
	configPath = filename

	log.Infof("Loading config from %s", filename)

	file, err := os.ReadFile(filename)
	if err != nil {
		configLoadErr = fmt.Errorf("could not read config file %s: %w", filename, err)
		log.Warnf("Could not load config file: %v", configLoadErr)
		return config
	}

	err = yaml.UnmarshalStrict(file, &config)
	if err != nil {
		configLoadErr = fmt.Errorf("could not parse config file %s: %w", filename, err)
		config = newDefaultConfig()
		applyNetworkOverlay(file, config)
		log.Warnf("Could not load config file: %v", configLoadErr)
		return config
	}

	if errs := validateConfig(config); len(errs) > 0 {
		configLoadErr = fmt.Errorf("%s", strings.Join(errs, "; "))
		log.Warnf("Config validation failed: %v", configLoadErr)
	}

	return config
}
