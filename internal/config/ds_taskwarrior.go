package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

type TaskwarriorConfig struct {
	TaskRcFile string `yaml:"taskrc_file" validate:"omitempty,file"`
	Limit      int    `yaml:"limit"`

	TemplateFile string `yaml:"template_file" validate:"omitempty,filepath"`

	Cached      bool          `yaml:"cached"`
	CacheExpiry time.Duration `yaml:"cache_expiry"`

	SummaryDays        int  `yaml:"summary_days" validate:"omitempty,gte=1,lte=14"`
	ExcludeFromSummary bool `yaml:"exclude_from_summary"`
}

func (ds *TaskwarriorConfig) UnmarshalYAML(node *yaml.Node) error {
	type tmp TaskwarriorConfig

	conf := &tmp{
		Cached:      false,
		CacheExpiry: 5 * time.Minute,
	}
	if err := node.Decode(&conf); err != nil {
		return err
	}

	*ds = TaskwarriorConfig(*conf)
	return nil
}

func (ds *TaskwarriorConfig) Type() string {
	return Taskwarrior
}

func (ds *TaskwarriorConfig) IsCached() bool {
	return ds.Cached
}

func (ds *TaskwarriorConfig) GetCacheExpiry() time.Duration {
	return ds.CacheExpiry
}
