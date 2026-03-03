package multirepo

// Config holds configuration for multi-repository analysis
type Config struct {
	Repositories []RepositoryConfig `yaml:"repositories" json:"repositories"`
	OutputPath   string             `yaml:"output_path" json:"output_path"`
}

// RepositoryConfig defines a single repository to analyze
type RepositoryConfig struct {
	Name string `yaml:"name" json:"name"`
	Path string `yaml:"path" json:"path"`
}
