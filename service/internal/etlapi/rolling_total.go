package api

import (
	"fmt"
	"strconv"

	pb "github.com/jamesread/data-cleaner/gen/data_cleaner/api/v1"
	"github.com/jamesread/data-cleaner/internal/config"
)

func applyRollingTotalCheck(rtcfg *config.RollingTotalConfig, rows []Row, columnOrder []string, res *pb.PreviewResponse) {
	if len(rows) == 0 {
		return
	}
	valueCol := rtcfg.ValueColumnOrDefault()
	balanceCol := rtcfg.BalanceColumnOrDefault()
	tolerance := rtcfg.ToleranceOrDefault()
	lastBalance := 0.0
	lastFile := "?"
	lastLineNumber := int64(0)

	for _, row := range rows {
		if row.Index == 0 {
			lastBalance = parseMoney(row.Columns[balanceCol])
			lastFile = row.Filename
			lastLineNumber = row.LineNumber
			continue
		}

		value := parseMoney(row.Columns[valueCol])
		balance := parseMoney(row.Columns[balanceCol])
		newBalance := lastBalance + value
		diff := abs(newBalance - balance)
		if diff > tolerance {
			issue := &pb.Issue{
				Description: "Rolling total mismatch, possible missing data",
			}
			issue.Expected = append(issue.Expected, &pb.RowAttribute{
				Key: balanceCol,
				Val: fmt.Sprintf("%v", newBalance),
			})
			issue.Intermediate = append(issue.Intermediate, &pb.RowAttribute{
				Key: valueCol,
				Val: fmt.Sprintf("%v", value),
			})
			issue.Intermediate = append(issue.Intermediate, &pb.RowAttribute{
				Key: "Diff",
				Val: fmt.Sprintf("%v", diff),
			})
			issue.Actual = append(issue.Actual, &pb.RowAttribute{
				Key: balanceCol,
				Val: fmt.Sprintf("%v", balance),
			})
			issue.CurrentLocationLineNumber = row.LineNumber
			issue.CurrentLocationFilename = row.Filename
			issue.LastLocationFilename = lastFile
			issue.LastLocationLineNumber = lastLineNumber
			res.Issues = append(res.Issues, issue)
		}

		lastFile = row.Filename
		lastBalance = balance
		lastLineNumber = row.LineNumber
	}
}

func parseMoney(value string) float64 {
	if value == "" {
		return 0
	}
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return v
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
