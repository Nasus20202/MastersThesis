package scenario

import "path/filepath"

type ClusterConfig struct {
	Kind KindConfig `yaml:"kind,omitempty"`
}

type KindConfig struct {
	ConfigFile string `yaml:"config,omitempty" validate:"omitempty,notblank"`
	dir        string
}

func (c *ClusterConfig) setDir(dir string) {
	c.Kind.dir = dir
}

func (c KindConfig) ConfigPath() string {
	if c.ConfigFile == "" || filepath.IsAbs(c.ConfigFile) || c.dir == "" {
		return c.ConfigFile
	}
	return filepath.Join(c.dir, c.ConfigFile)
}
