package api

import pb "github.com/jamesread/data-cleaner/gen/data_cleaner/api/v1"

func (api *EtlApi) ListTransformationTypes() *pb.ListTransformationTypesResponse {
	return &pb.ListTransformationTypesResponse{
		Types: []*pb.TransformationType{
			{
				Id:          "replacements",
				Name:        "Replacements",
				Description: "Maps description text to a category using exact string matches or regular expressions.",
				YamlExample: `transform:
  - replacements:
      exact:
        "TESCO STORES": Groceries
      regex:
        "^AMAZON.*": Shopping`,
			},
			{
				Id:          "add_category",
				Name:        "Add category",
				Description: "Sets a target column from exact values and regex rules keyed on a source column. Mappings can be inline or loaded from a file.",
				YamlExample: `transform:
  - add_category:
      source_column: description
      target_column: category
      from_file: categories.yaml
      values:
        "INLINE MERCHANT": Groceries`,
			},
			{
				Id:          "date_normalize",
				Name:        "Date normalize",
				Description: "Detects the date format used in a column and normalizes all values to ISO datetimes (YYYY-MM-DD 00:00:00 when no time is present).",
				YamlExample: `transform:
  - date_normalize:
      column: date`,
			},
			{
				Id:          "date_to_incremental",
				Name:        "Date to incremental",
				Description: "Walks rows in source order and adds one second when a row's datetime equals the previous row's, making duplicate dates sortable while preserving row order.",
				YamlExample: `transform:
  - date_to_incremental:
      column: date`,
			},
			{
				Id:          "rolling_total",
				Name:        "Rolling total",
				Description: "Validates that previous balance plus value equals the current balance for each row. Column names default to value and balance.",
				YamlExample: `transform:
  - rolling_total:
      value_column: value
      balance_column: balance
      tolerance: 0.001`,
			},
			{
				Id:          "drop_column",
				Name:        "Drop column",
				Description: "Removes columns from the pipeline before load.",
				YamlExample: `transform:
  - drop_column:
      - internal_id
      - notes`,
			},
			{
				Id:          "rename_column",
				Name:        "Rename column",
				Description: "Renames pipeline columns to INSERT column names for load. Columns not listed keep their source names.",
				YamlExample: `transform:
  - rename_column:
      date: transaction_date
      description: memo
      value: amount`,
			},
			{
				Id:          "append_hash",
				Name:        "Append hash",
				Description: "Appends a short SHA256 fingerprint of the original source row, in brackets, to the end of a column value.",
				YamlExample: `transform:
  - append_hash:
      column: description`,
			},
		},
	}
}
