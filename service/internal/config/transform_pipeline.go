package config

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// TransformPipeline is an ordered list of transformation steps. List order defines execution order and ordinals.
type TransformPipeline []TransformStep

// TransformStep holds exactly one transformation type per list entry.
type TransformStep struct {
	Replacements      *ReplacementsConfig      `yaml:"replacements,omitempty"`
	AddCategory       *AddCategoryConfig       `yaml:"add_category,omitempty"`
	DateNormalize     *DateNormalizeConfig     `yaml:"date_normalize,omitempty"`
	DateToIncremental *DateToIncrementalConfig `yaml:"date_to_incremental,omitempty"`
	RollingTotal      *RollingTotalConfig      `yaml:"rolling_total,omitempty"`
	DropColumn        []string                 `yaml:"drop_column,omitempty"`
	RenameColumn      map[string]string        `yaml:"rename_column,omitempty"`
	AppendHash        *AppendHashConfig        `yaml:"append_hash,omitempty"`
}

type legacyTransformBlock struct {
	Replacements      *ReplacementsConfig      `yaml:"replacements,omitempty"`
	AddCategory       *AddCategoryConfig       `yaml:"add_category,omitempty"`
	DateNormalize     *DateNormalizeConfig     `yaml:"date_normalize,omitempty"`
	DateToIncremental *DateToIncrementalConfig `yaml:"date_to_incremental,omitempty"`
	RollingTotal      *RollingTotalConfig      `yaml:"rolling_total,omitempty"`
	DropColumn        []string                 `yaml:"drop_column,omitempty"`
	RenameColumn      map[string]string        `yaml:"rename_column,omitempty"`
	AppendHash        *AppendHashConfig        `yaml:"append_hash,omitempty"`
}

func (p *TransformPipeline) UnmarshalYAML(unmarshal func(any) error) error {
	var raw any
	if err := unmarshal(&raw); err != nil {
		return err
	}
	if raw == nil {
		*p = nil
		return nil
	}

	switch v := raw.(type) {
	case []any:
		steps := make(TransformPipeline, len(v))
		for i, item := range v {
			data, err := yaml.Marshal(item)
			if err != nil {
				return fmt.Errorf("transform step %d: %w", i+1, err)
			}
			var step TransformStep
			if err := yaml.Unmarshal(data, &step); err != nil {
				return fmt.Errorf("transform step %d: %w", i+1, err)
			}
			if step.Kind() == "" {
				return fmt.Errorf("transform step %d: unknown or empty transformation", i+1)
			}
			steps[i] = step
		}
		*p = steps
		return nil
	case map[any]any:
		data, err := yaml.Marshal(v)
		if err != nil {
			return err
		}
		var legacy legacyTransformBlock
		if err := yaml.Unmarshal(data, &legacy); err != nil {
			return err
		}
		*p = legacy.toPipeline()
		return nil
	default:
		return fmt.Errorf("transform must be a YAML sequence or mapping")
	}
}

func (l legacyTransformBlock) toPipeline() TransformPipeline {
	var steps TransformPipeline
	if hasReplacements(l.Replacements) {
		steps = append(steps, TransformStep{Replacements: l.Replacements})
	}
	if l.AddCategory != nil {
		steps = append(steps, TransformStep{AddCategory: l.AddCategory})
	}
	if l.DateNormalize != nil && l.DateNormalize.Configured() {
		steps = append(steps, TransformStep{DateNormalize: l.DateNormalize})
	}
	if l.DateToIncremental != nil && l.DateToIncremental.Configured() {
		steps = append(steps, TransformStep{DateToIncremental: l.DateToIncremental})
	}
	if l.RollingTotal != nil {
		steps = append(steps, TransformStep{RollingTotal: l.RollingTotal})
	}
	if len(l.DropColumn) > 0 {
		steps = append(steps, TransformStep{DropColumn: l.DropColumn})
	}
	if len(l.RenameColumn) > 0 {
		steps = append(steps, TransformStep{RenameColumn: l.RenameColumn})
	}
	if l.AppendHash != nil && l.AppendHash.Configured() {
		steps = append(steps, TransformStep{AppendHash: l.AppendHash})
	}
	return steps
}

func hasReplacements(replacements *ReplacementsConfig) bool {
	if replacements == nil {
		return false
	}
	return len(replacements.Exact) > 0 || len(replacements.Regex) > 0
}

func (s TransformStep) Kind() string {
	switch {
	case hasReplacements(s.Replacements):
		return "replacements"
	case s.AddCategory != nil:
		return "add_category"
	case s.DateNormalize != nil && s.DateNormalize.Configured():
		return "date_normalize"
	case s.DateToIncremental != nil && s.DateToIncremental.Configured():
		return "date_to_incremental"
	case s.RollingTotal != nil:
		return "rolling_total"
	case len(s.DropColumn) > 0:
		return "drop_column"
	case len(s.RenameColumn) > 0:
		return "rename_column"
	case s.AppendHash != nil && s.AppendHash.Configured():
		return "append_hash"
	default:
		return ""
	}
}

func (c *Config) TransformSteps() TransformPipeline {
	if c == nil {
		return nil
	}
	return c.Transform
}

func StepsThroughOrdinal(steps TransformPipeline, throughOrdinal int32) TransformPipeline {
	if throughOrdinal <= 0 || int(throughOrdinal) >= len(steps) {
		return steps
	}
	return steps[:throughOrdinal]
}

func (p TransformPipeline) DateNormalizeSteps() []DateNormalizeConfig {
	var out []DateNormalizeConfig
	for _, step := range p {
		if step.DateNormalize != nil && step.DateNormalize.Configured() {
			out = append(out, *step.DateNormalize)
		}
	}
	return out
}
