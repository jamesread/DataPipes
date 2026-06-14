package api

import (
	"fmt"
	"sort"
	"strings"

	pb "github.com/jamesread/data-cleaner/gen/data_cleaner/api/v1"
	"github.com/jamesread/data-cleaner/internal/config"
)

func newTransformation(name, description string, ordinal int32) *pb.Transformation {
	return &pb.Transformation{
		Name:        name,
		Description: description,
		Ordinal:     ordinal,
	}
}

func TransformationsFromConfig(cfg *config.Config) []*pb.Transformation {
	return transformationsFromConfig(cfg, false)
}

// JobTransformationsFromConfig returns a compact transformation list for ListJobs.
func JobTransformationsFromConfig(cfg *config.Config) []*pb.Transformation {
	return transformationsFromConfig(cfg, true)
}

func transformationsFromConfig(cfg *config.Config, summarizeAddCategory bool) []*pb.Transformation {
	if cfg == nil {
		return nil
	}

	var out []*pb.Transformation
	for i, step := range cfg.TransformSteps() {
		ordinal := int32(i + 1)
		out = append(out, stepTransformations(step, ordinal, summarizeAddCategory)...)
	}
	return out
}

func stepTransformations(step config.TransformStep, ordinal int32, summarizeAddCategory bool) []*pb.Transformation {
	switch step.Kind() {
	case "replacements":
		return replacementsTransformations(step.Replacements, ordinal)
	case "add_category":
		return addCategoryTransformations(step.AddCategory, ordinal, summarizeAddCategory)
	case "date_normalize":
		return dateNormalizeTransformations(step.DateNormalize, ordinal)
	case "date_to_incremental":
		return []*pb.Transformation{newTransformation(
			"date_to_incremental",
			fmt.Sprintf("date_to_incremental: %q (+1s on duplicate)", step.DateToIncremental.Column),
			ordinal,
		)}
	case "rolling_total":
		return []*pb.Transformation{newTransformation(
			"rolling_total",
			fmt.Sprintf("rolling_total: validate balance continuity (tolerance %v)", step.RollingTotal.ToleranceOrDefault()),
			ordinal,
		)}
	case "drop_column":
		return dropColumnTransformations(step.DropColumn, ordinal)
	case "rename_column":
		return renameColumnTransformations(step.RenameColumn, ordinal)
	case "append_hash":
		return []*pb.Transformation{newTransformation(
			"append_hash",
			fmt.Sprintf("append_hash: %q (+ short SHA256 of source row)", step.AppendHash.Column),
			ordinal,
		)}
	default:
		return nil
	}
}

func replacementsTransformations(replacements *config.ReplacementsConfig, ordinal int32) []*pb.Transformation {
	if replacements == nil {
		return nil
	}
	var out []*pb.Transformation

	exactKeys := make([]string, 0, len(replacements.Exact))
	for from := range replacements.Exact {
		exactKeys = append(exactKeys, from)
	}
	sort.Strings(exactKeys)
	for _, from := range exactKeys {
		out = append(out, newTransformation(
			"exact",
			fmt.Sprintf("Exact match %q → %q", from, replacements.Exact[from]),
			ordinal,
		))
	}

	regexKeys := make([]string, 0, len(replacements.Regex))
	for pattern := range replacements.Regex {
		regexKeys = append(regexKeys, pattern)
	}
	sort.Strings(regexKeys)
	for _, pattern := range regexKeys {
		out = append(out, newTransformation(
			"regex",
			fmt.Sprintf("Regex %q → %q", pattern, replacements.Regex[pattern]),
			ordinal,
		))
	}
	return out
}

func addCategoryTransformations(addCategory *config.AddCategoryConfig, ordinal int32, summarizeAddCategory bool) []*pb.Transformation {
	if addCategory == nil {
		return nil
	}
	resolved, err := addCategory.Resolve(config.ConfigDirectory())
	if err != nil || resolved == nil || !resolved.HasMappings() {
		return nil
	}

	if summarizeAddCategory {
		return []*pb.Transformation{newTransformation(
			"add_category",
			addCategorySummaryDescription(resolved),
			ordinal,
		)}
	}

	var out []*pb.Transformation
	if resolved.FromFile != "" {
		out = append(out, newTransformation(
			"add_category",
			fmt.Sprintf("add_category: from_file %q", resolved.FromFile),
			ordinal,
		))
	}
	for _, key := range config.SortedStringMapKeys(resolved.Values) {
		out = append(out, newTransformation(
			"add_category",
			fmt.Sprintf("add_category: %q → %q (from %q to %q)",
				key, resolved.Values[key], resolved.SourceColumn, resolved.TargetColumn),
			ordinal,
		))
	}
	for _, pattern := range config.SortedStringMapKeys(resolved.Regex) {
		out = append(out, newTransformation(
			"add_category",
			fmt.Sprintf("add_category regex: %q → %q (from %q to %q)",
				pattern, resolved.Regex[pattern], resolved.SourceColumn, resolved.TargetColumn),
			ordinal,
		))
	}
	return out
}

func dateNormalizeTransformations(dateNormalize *config.DateNormalizeConfig, ordinal int32) []*pb.Transformation {
	if dateNormalize == nil || !dateNormalize.Configured() {
		return nil
	}
	var out []*pb.Transformation
	for _, col := range config.SortedDateNormalizeColumns(dateNormalize.TargetColumns()) {
		out = append(out, newTransformation(
			"date_normalize",
			fmt.Sprintf("date_normalize: %q → ISO datetime", col),
			ordinal,
		))
	}
	return out
}

func dropColumnTransformations(cols []string, ordinal int32) []*pb.Transformation {
	out := make([]*pb.Transformation, 0, len(cols))
	for _, col := range sortedDropColumns(cols) {
		out = append(out, newTransformation(
			"drop_column",
			fmt.Sprintf("drop_column: %q", col),
			ordinal,
		))
	}
	return out
}

func renameColumnTransformations(rename map[string]string, ordinal int32) []*pb.Transformation {
	out := make([]*pb.Transformation, 0, len(rename))
	for _, src := range sortedRenameKeys(rename) {
		out = append(out, newTransformation(
			"rename_column",
			fmt.Sprintf("rename_column: %q → %q", src, rename[src]),
			ordinal,
		))
	}
	return out
}

func addCategorySummaryDescription(resolved *config.ResolvedAddCategory) string {
	var parts []string
	valueCount := len(resolved.Values)
	regexCount := len(resolved.Regex)
	if valueCount > 0 {
		part := "value replacements"
		if valueCount == 1 {
			part = "value replacement"
		}
		parts = append(parts, fmt.Sprintf("%d %s", valueCount, part))
	}
	if regexCount > 0 {
		part := "regex replacements"
		if regexCount == 1 {
			part = "regex replacement"
		}
		parts = append(parts, fmt.Sprintf("%d %s", regexCount, part))
	}
	return strings.Join(parts, ", ")
}
