package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/jamesread/data-cleaner/internal/config"
)

const appendHashLength = 8

func applyAppendHashStep(cfg *config.AppendHashConfig, row *DataRow, sourceRow map[string]string, columnOrder []string) {
	if cfg == nil || cfg.Column == "" {
		return
	}

	hash := shortSourceRowHash(sourceRow, columnOrder)
	current := row.Get(cfg.Column)
	row.Set(cfg.Column, fmt.Sprintf("%s [%s]", current, hash))
}

func shortSourceRowHash(row map[string]string, columnOrder []string) string {
	var b strings.Builder
	for i, col := range columnOrder {
		if i > 0 {
			b.WriteByte('\x1e')
		}
		b.WriteString(row[col])
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])[:appendHashLength]
}
