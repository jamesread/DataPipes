package config

import "fmt"

type FireflyColumnMap map[string]string

func (c *Connection) TransactionTypeHintColumn() string {
	if c == nil {
		return ""
	}
	return c.TransactionType
}

func (c *Connection) FireflyApplyRules() bool {
	if c == nil || c.ApplyRules == nil {
		return true
	}
	return *c.ApplyRules
}

func (c *Connection) ResolvedFireflyColumnMap() FireflyColumnMap {
	defaults := FireflyColumnMap{
		"date":             "date",
		"amount":           "amount",
		"description":      "description",
		"destination_name": "category",
		"category_name":    "category",
	}
	if c == nil || len(c.LoadColumns) == 0 {
		return defaults
	}
	out := make(FireflyColumnMap, len(defaults)+len(c.LoadColumns))
	for k, v := range defaults {
		out[k] = v
	}
	for k, v := range c.LoadColumns {
		if v != "" {
			out[k] = v
		}
	}
	return out
}

func (c *Connection) ValidateFireflyIII() error {
	if c == nil {
		return nil
	}
	if c.URL == "" {
		return fmt.Errorf("url is required")
	}
	if c.Token == "" {
		return fmt.Errorf("token is required")
	}
	if c.SourceAccount == "" {
		return fmt.Errorf("source_account is required")
	}
	return nil
}

func (c *Connection) DefaultExpenseAccountOr(fallback string) string {
	if c != nil && c.DefaultExpenseAccount != "" {
		return c.DefaultExpenseAccount
	}
	return fallback
}
