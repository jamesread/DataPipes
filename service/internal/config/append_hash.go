package config

type AppendHashConfig struct {
	Column string `yaml:"column"`
}

func (a *AppendHashConfig) Configured() bool {
	return a != nil && a.Column != ""
}
