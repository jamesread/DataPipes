package config

type DateToIncrementalConfig struct {
	Column string `yaml:"column"`
}

func (d *DateToIncrementalConfig) Configured() bool {
	return d != nil && d.Column != ""
}
