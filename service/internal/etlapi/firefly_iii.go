package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jamesread/data-cleaner/internal/config"
	log "github.com/sirupsen/logrus"
)

var fireflyAppendHashSuffix = regexp.MustCompile(` \[[0-9a-f]{8}\]$`)

const fireflyProgressInterval = 25

type FireflyIIIConnector struct {
	conn   *config.Connection
	client *http.Client
	apiURL string
}

type fireflyTransaction struct {
	Type            string `json:"type"`
	Date            string `json:"date"`
	Amount          string `json:"amount"`
	Description     string `json:"description,omitempty"`
	SourceName      string `json:"source_name,omitempty"`
	DestinationName string `json:"destination_name,omitempty"`
	CategoryName    string `json:"category_name,omitempty"`
}

type fireflyCreateRequest struct {
	ApplyRules             bool                 `json:"apply_rules"`
	FireWebhooks           bool                 `json:"fire_webhooks"`
	ErrorIfDuplicateHash   bool                 `json:"error_if_duplicate_hash"`
	Transactions           []fireflyTransaction `json:"transactions"`
}

func NewFireflyIIIConnector(conn *config.Connection) *FireflyIIIConnector {
	return &FireflyIIIConnector{
		conn:   conn,
		client: &http.Client{Timeout: 60 * time.Second},
		apiURL: fireflyAPIURL(conn.URL),
	}
}

func fireflyAPIURL(base string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(base, "/api/v1") {
		return base
	}
	return base + "/api/v1"
}

func (c *FireflyIIIConnector) Connect() error {
	req, err := http.NewRequest(http.MethodGet, c.apiURL+"/about/user", nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("firefly III health check failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *FireflyIIIConnector) Load(dataRows []DataRow, columns []string, report RowLoadReporter) error {
	if c.conn == nil {
		return fmt.Errorf("firefly III connection not configured")
	}

	colMap := c.conn.ResolvedFireflyColumnMap()
	for i, row := range dataRows {
		tx, err := buildFireflyTransaction(c.conn, row, colMap)
		if err != nil {
			if report != nil {
				if repErr := report(i+1, row, err); repErr != nil {
					return repErr
				}
			}
			continue
		}
		rowErr := c.createTransaction(tx)
		if report != nil {
			if repErr := report(i+1, row, rowErr); repErr != nil {
				return repErr
			}
		}
		if rowErr != nil {
			log.Errorf("Firefly III row %d: %v", i+1, rowErr)
			continue
		}
		if (i+1)%fireflyProgressInterval == 0 {
			log.Infof("Firefly III: created %d transactions", i+1)
		}
	}
	log.Infof("Firefly III: created %d transactions", len(dataRows))

	return nil
}

func (c *FireflyIIIConnector) createTransaction(tx fireflyTransaction) error {
	payload := fireflyCreateRequest{
		ApplyRules:           c.conn.FireflyApplyRules(),
		FireWebhooks:         false,
		ErrorIfDuplicateHash: false,
		Transactions:         []fireflyTransaction{tx},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.apiURL+"/transactions", bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}

func (c *FireflyIIIConnector) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.conn.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
}

func buildFireflyTransaction(conn *config.Connection, row DataRow, colMap config.FireflyColumnMap) (fireflyTransaction, error) {
	tx := fireflyTransaction{
		SourceName: conn.SourceAccount,
	}

	rawAmount, amountCol, err := resolveRawAmount(row, colMap)
	if err != nil {
		return fireflyTransaction{}, err
	}

	for fireflyField, pipelineCol := range colMap {
		if fireflyField == "type" || fireflyField == "amount" {
			continue
		}
		value := strings.TrimSpace(row.Get(pipelineCol))
		if value == "" {
			continue
		}
		switch fireflyField {
		case "date":
			tx.Date = formatFireflyDate(value)
		case "description":
			tx.Description = value
		case "source_name":
			tx.SourceName = value
		case "destination_name":
			tx.DestinationName = value
		case "category_name":
			tx.CategoryName = value
		}
	}

	hint := transactionTypeHint(row, conn)
	tx.Type = resolveTransactionType(hint, rawAmount)
	tx.Amount = formatFireflyAmountValue(rawAmount)
	fillMissingDestination(&tx, row, colMap, conn)
	finalizeFireflyAccounts(&tx, conn)

	if tx.SourceName == "" {
		tx.SourceName = conn.SourceAccount
	}
	if tx.Date == "" {
		return fireflyTransaction{}, fmt.Errorf("date is required")
	}
	if amountCol == "" {
		return fireflyTransaction{}, fmt.Errorf("amount is required")
	}
	if tx.DestinationName == "" {
		return fireflyTransaction{}, fmt.Errorf("destination_name is required for %s", tx.Type)
	}

	return tx, nil
}

func fillMissingDestination(tx *fireflyTransaction, row DataRow, colMap config.FireflyColumnMap, conn *config.Connection) {
	if tx.DestinationName != "" {
		return
	}
	if tx.CategoryName != "" {
		tx.DestinationName = tx.CategoryName
		return
	}
	if tx.Type == "deposit" {
		tx.DestinationName = conn.SourceAccount
		return
	}
	if tx.Description != "" && !isMaskedFireflyAccountName(tx.Description) {
		tx.DestinationName = tx.Description
	}
}

func finalizeFireflyAccounts(tx *fireflyTransaction, conn *config.Connection) {
	if tx.Type == "deposit" {
		revenue := sanitizeFireflyAccountName(tx.CategoryName)
		if revenue == "" || isMaskedFireflyAccountName(revenue) {
			revenue = sanitizeFireflyAccountName(tx.Description)
		}
		if revenue != "" && !isMaskedFireflyAccountName(revenue) {
			tx.SourceName = revenue
		}
		tx.DestinationName = conn.SourceAccount
	}

	tx.SourceName = sanitizeFireflyAccountName(tx.SourceName)
	tx.DestinationName = sanitizeFireflyAccountName(tx.DestinationName)

	switch tx.Type {
	case "withdrawal", "transfer":
		if tx.DestinationName == "" || isMaskedFireflyAccountName(tx.DestinationName) {
			if tx.Type == "transfer" && conn.DefaultTransferAccount != "" {
				tx.DestinationName = conn.DefaultTransferAccount
			} else {
				tx.DestinationName = conn.DefaultExpenseAccountOr("Unidentified")
			}
		}
	}
}

func sanitizeFireflyAccountName(name string) string {
	return strings.TrimSpace(fireflyAppendHashSuffix.ReplaceAllString(strings.TrimSpace(name), ""))
}

func isMaskedFireflyAccountName(name string) bool {
	name = sanitizeFireflyAccountName(name)
	if name == "" {
		return true
	}
	masked := 0
	for _, r := range name {
		if r == '*' || r == '•' {
			masked++
		}
	}
	return masked*2 >= len(name)
}

func resolveRawAmount(row DataRow, colMap config.FireflyColumnMap) (float64, string, error) {
	if amountCol := colMap["amount"]; amountCol != "" {
		if raw, err := parseAmount(row.Get(amountCol)); err == nil {
			return raw, amountCol, nil
		}
	}
	if raw, err := parseAmount(row.Get("value")); err == nil {
		return raw, "value", nil
	}
	if amountCol := colMap["amount"]; amountCol != "" {
		return 0, "", fmt.Errorf("amount column %q: invalid value", amountCol)
	}
	return 0, "", fmt.Errorf("amount is required")
}

func transactionTypeHint(row DataRow, conn *config.Connection) string {
	hintCol := conn.TransactionTypeHintColumn()
	if hintCol == "" {
		return ""
	}
	return row.Get(hintCol)
}

func resolveTransactionType(hint string, rawAmount float64) string {
	if txType := transactionTypeFromHint(hint); txType != "" {
		return txType
	}
	if rawAmount < 0 {
		return "withdrawal"
	}
	if rawAmount > 0 {
		return "deposit"
	}
	return "withdrawal"
}

func transactionTypeFromHint(hint string) string {
	switch strings.ToUpper(strings.TrimSpace(hint)) {
	case "TFR", "TRANSFER", "XFER":
		return "transfer"
	case "WD", "WITHDRAWAL", "DEBIT", "DR", "D":
		return "withdrawal"
	case "DEP", "DEPOSIT", "CREDIT", "CR", "C":
		return "deposit"
	}
	return ""
}

func parseAmount(value string) (float64, error) {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, ",", "")
	return strconv.ParseFloat(value, 64)
}

func formatFireflyAmountValue(amount float64) string {
	return fmt.Sprintf("%.2f", math.Abs(amount))
}

func formatFireflyDate(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Replace(value, " ", "T", 1)
	if !strings.Contains(value, "T") {
		return value + "T00:00:00"
	}
	return value
}
