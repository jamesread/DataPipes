package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jamesread/data-cleaner/internal/config"
)

func TestFormatFireflyAmountValue(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{50, "50.00"},
		{-50.5, "50.50"},
		{1234.56, "1234.56"},
	}
	for _, tc := range tests {
		got := formatFireflyAmountValue(tc.in)
		if got != tc.want {
			t.Fatalf("formatFireflyAmountValue(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestResolveTransactionType(t *testing.T) {
	tests := []struct {
		hint   string
		amount float64
		want   string
	}{
		{"", -10, "withdrawal"},
		{"", 10, "deposit"},
		{"TFR", -10, "transfer"},
		{"tfr", 10, "transfer"},
		{"CR", -10, "deposit"},
		{"DR", 10, "withdrawal"},
	}
	for _, tc := range tests {
		got := resolveTransactionType(tc.hint, tc.amount)
		if got != tc.want {
			t.Fatalf("resolveTransactionType(%q, %v) = %q, want %q", tc.hint, tc.amount, got, tc.want)
		}
	}
}

func TestSanitizeFireflyAccountNameStripsAppendHash(t *testing.T) {
	got := sanitizeFireflyAccountName("Coffee shop [a2007a83]")
	if got != "Coffee shop" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildFireflyTransactionMaskedWithdrawalUsesDefaultExpense(t *testing.T) {
	conn := &config.Connection{SourceAccount: "Checking"}
	row := DataRow{contents: map[string]string{
		"date":        "2024-01-15",
		"amount":      "-10",
		"description": "*********************************** [a2007a83]",
	}}

	tx, err := buildFireflyTransaction(conn, row, conn.ResolvedFireflyColumnMap())
	if err != nil {
		t.Fatal(err)
	}
	if tx.DestinationName != "Unidentified" {
		t.Fatalf("destination = %q", tx.DestinationName)
	}
}

func TestBuildFireflyTransactionTransferMaskedUsesUnidentified(t *testing.T) {
	conn := &config.Connection{
		SourceAccount:   "Checking",
		TransactionType: "type",
	}
	row := DataRow{contents: map[string]string{
		"date":        "2024-01-15",
		"amount":      "-100",
		"description": "*********************************** [a2007a83]",
		"type":        "TFR",
	}}

	tx, err := buildFireflyTransaction(conn, row, conn.ResolvedFireflyColumnMap())
	if err != nil {
		t.Fatal(err)
	}
	if tx.DestinationName != "Unidentified" {
		t.Fatalf("destination = %q", tx.DestinationName)
	}
}

func TestBuildFireflyTransactionTransferMaskedUsesDefaultTransferAccount(t *testing.T) {
	conn := &config.Connection{
		SourceAccount:          "Checking",
		TransactionType:        "type",
		DefaultTransferAccount: "Savings",
	}
	row := DataRow{contents: map[string]string{
		"date":        "2024-01-15",
		"amount":      "-100",
		"description": "*********************************** [a2007a83]",
		"type":        "TFR",
	}}

	tx, err := buildFireflyTransaction(conn, row, conn.ResolvedFireflyColumnMap())
	if err != nil {
		t.Fatal(err)
	}
	if tx.DestinationName != "Savings" {
		t.Fatalf("destination = %q", tx.DestinationName)
	}
}

func TestBuildFireflyTransactionNegativeWithdrawal(t *testing.T) {
	conn := &config.Connection{
		SourceAccount:   "Checking",
		TransactionType: "type",
	}
	row := DataRow{contents: map[string]string{
		"date":        "2024-01-15 00:00:00",
		"amount":      "-42.50",
		"description": "Coffee shop",
		"category":    "Food",
		"type":        "POS",
	}}

	tx, err := buildFireflyTransaction(conn, row, conn.ResolvedFireflyColumnMap())
	if err != nil {
		t.Fatal(err)
	}
	if tx.Type != "withdrawal" {
		t.Fatalf("type = %q", tx.Type)
	}
	if tx.Amount != "42.50" {
		t.Fatalf("amount = %q", tx.Amount)
	}
}

func TestBuildFireflyTransactionPositiveDeposit(t *testing.T) {
	conn := &config.Connection{SourceAccount: "Checking"}
	row := DataRow{contents: map[string]string{
		"date":     "2024-01-15",
		"amount":   "100.00",
		"category": "Salary",
	}}

	tx, err := buildFireflyTransaction(conn, row, conn.ResolvedFireflyColumnMap())
	if err != nil {
		t.Fatal(err)
	}
	if tx.Type != "deposit" {
		t.Fatalf("type = %q", tx.Type)
	}
}

func TestBuildFireflyTransactionTransferHint(t *testing.T) {
	conn := &config.Connection{
		SourceAccount:   "Checking",
		TransactionType: "type",
	}
	row := DataRow{contents: map[string]string{
		"date":        "2024-01-15",
		"amount":      "-500.00",
		"description": "Savings",
		"type":        "TFR",
	}}

	tx, err := buildFireflyTransaction(conn, row, conn.ResolvedFireflyColumnMap())
	if err != nil {
		t.Fatal(err)
	}
	if tx.Type != "transfer" {
		t.Fatalf("type = %q", tx.Type)
	}
	if tx.DestinationName != "Savings" {
		t.Fatalf("destination = %q", tx.DestinationName)
	}
}

func TestBuildFireflyTransactionTransferUsesDestinationColumn(t *testing.T) {
	conn := &config.Connection{
		SourceAccount:   "Checking",
		TransactionType: "type",
		LoadColumns: config.FireflyColumnMap{
			"destination_name": "counterparty",
		},
	}
	row := DataRow{contents: map[string]string{
		"date":         "2024-01-15",
		"amount":       "-100",
		"description":  "Internal move",
		"counterparty": "Holiday pot",
		"type":         "TFR",
	}}

	tx, err := buildFireflyTransaction(conn, row, conn.ResolvedFireflyColumnMap())
	if err != nil {
		t.Fatal(err)
	}
	if tx.DestinationName != "Holiday pot" {
		t.Fatalf("destination = %q", tx.DestinationName)
	}
}

func TestBuildFireflyTransactionDepositUsesSourceAccount(t *testing.T) {
	conn := &config.Connection{SourceAccount: "Checking"}
	row := DataRow{contents: map[string]string{
		"date":        "2024-01-15",
		"amount":      "100.00",
		"description": "Salary payment",
	}}

	tx, err := buildFireflyTransaction(conn, row, conn.ResolvedFireflyColumnMap())
	if err != nil {
		t.Fatal(err)
	}
	if tx.DestinationName != "Checking" {
		t.Fatalf("destination = %q", tx.DestinationName)
	}
}

func TestFireflyIIIConnectorLoadOneRequestPerRow(t *testing.T) {
	postCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/about/user"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{}}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/transactions"):
			postCount++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"id":"1"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	conn := &config.Connection{
		URL:           server.URL,
		Token:         "test-token",
		SourceAccount: "Checking",
	}
	connector := NewFireflyIIIConnector(conn)
	rows := []DataRow{
		{contents: map[string]string{"date": "2024-01-15", "amount": "-10", "category": "Food"}},
		{contents: map[string]string{"date": "2024-01-16", "amount": "20", "category": "Salary"}},
	}
	if err := connector.Load(rows, nil, nil); err != nil {
		t.Fatal(err)
	}
	if postCount != 2 {
		t.Fatalf("POST count = %d, want 2", postCount)
	}
}

func TestFireflyIIIConnectorLoad(t *testing.T) {
	var posted string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/about/user"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{}}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/transactions"):
			body := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(body)
			posted = string(body)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"id":"1"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	conn := &config.Connection{
		URL:           server.URL,
		Token:         "test-token",
		SourceAccount: "Checking",
	}
	connector := NewFireflyIIIConnector(conn)
	if err := connector.Connect(); err != nil {
		t.Fatal(err)
	}

	rows := []DataRow{{contents: map[string]string{
		"date":        "2024-01-15",
		"amount":      "10",
		"description": "Test",
		"category":    "Misc",
	}}}
	if err := connector.Load(rows, nil, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(posted, `"type":"deposit"`) {
		t.Fatalf("unexpected payload: %s", posted)
	}
	if !strings.Contains(posted, `"destination_name":"Checking"`) {
		t.Fatalf("unexpected payload: %s", posted)
	}
	if !strings.Contains(posted, `"source_name":"Misc"`) {
		t.Fatalf("unexpected payload: %s", posted)
	}
}
