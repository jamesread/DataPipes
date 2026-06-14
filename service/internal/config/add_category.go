package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v2"
)

type addCategoryFile struct {
	Values map[string]string `yaml:"values"`
	Regex  map[string]string `yaml:"regex"`
}

// ResolvedAddCategory holds merged add_category mappings from inline config and optional file.
type ResolvedAddCategory struct {
	SourceColumn string
	TargetColumn string
	Values       map[string]string
	Regex        map[string]string
	FromFile     string
}

func (ac *AddCategoryConfig) fromFilePath() string {
	if ac == nil {
		return ""
	}
	if ac.FromFile != "" {
		return ac.FromFile
	}
	return ac.FromFileCamel
}

func (ac *AddCategoryConfig) Resolve(configDir string) (*ResolvedAddCategory, error) {
	if ac == nil {
		return nil, nil
	}

	values := copyStringMap(ac.Values)
	regex := copyStringMap(ac.Regex)

	fromFile := ac.fromFilePath()
	if fromFile != "" {
		fileValues, fileRegex, err := loadAddCategoryFile(fromFile, configDir)
		if err != nil {
			return nil, err
		}
		values = mergeStringMaps(fileValues, values)
		regex = mergeStringMaps(fileRegex, regex)
	}

	target := ac.TargetColumn
	if target == "" {
		target = "category"
	}

	return &ResolvedAddCategory{
		SourceColumn: ac.SourceColumn,
		TargetColumn: target,
		Values:       values,
		Regex:        regex,
		FromFile:     fromFile,
	}, nil
}

func (r *ResolvedAddCategory) HasMappings() bool {
	if r == nil {
		return false
	}
	return len(r.Values) > 0 || len(r.Regex) > 0
}

func ConfigDirectory() string {
	if configPath == "" {
		return ""
	}
	return filepath.Dir(configPath)
}

func loadAddCategoryFile(filePath, configDir string) (values map[string]string, regex map[string]string, err error) {
	if !filepath.IsAbs(filePath) && configDir != "" {
		filePath = filepath.Join(configDir, filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("read add_category file %s: %w", filePath, err)
	}

	var file addCategoryFile
	if err := yaml.UnmarshalStrict(data, &file); err != nil {
		return nil, nil, fmt.Errorf("parse add_category file %s: %w", filePath, err)
	}

	return copyStringMap(file.Values), copyStringMap(file.Regex), nil
}

func mergeStringMaps(base, override map[string]string) map[string]string {
	out := copyStringMap(base)
	for k, v := range override {
		out[k] = v
	}
	return out
}

func copyStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func SortedStringMapKeys(m map[string]string) []string {
	return sortedStringMapKeys(m)
}

func sortedStringMapKeys(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
