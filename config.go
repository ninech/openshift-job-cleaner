package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

// config represents the structure of the main config file for the tool, and will be read by yaml
type config struct {
	Blacklist  []string                              `yaml:"blacklist"`
	Default    namespaceCleanupDescriptor            `yaml:"default"`
	Namespaces map[string]namespaceCleanupDescriptor `yaml:"namespaces"`
}

// namespaceCleanupDescriptor is a simple namespace descriptor allowing you to specify values based on the job exist status
type namespaceCleanupDescriptor struct {
	Success cleanupDescriptor `yaml:"success"`
	Failure cleanupDescriptor `yaml:"failure"`
}

// cleanupDescriptor is the basic unit which describes how a resource should be cleaned
type cleanupDescriptor struct {
	MaxAge    int `yaml:"max_age"`
}

// currentJobstate is the result of a call to oc getting a set of jobs
type currentJobstate struct {
	APIVersion string    `yaml:"apiVersion"`
	Items      []jobItem `yaml:"items"`
}

// ocTime is a wrapper for time so we can marshall it differently
type ocTime time.Time

// UnmarshalYAML will treat an ocTime element as time.Time so it can be parsed
func (t *ocTime) UnmarshalYAML(unmarshal func(interface{}) error) error {
	tm := new(time.Time)
	err := unmarshal(tm)
	if err == nil {
		*t = ocTime(*tm)
		return nil
	}
	return err
}

// condition simplifies the condition element of the job data array
type condition struct {
	LastTransitionTime ocTime `yaml:"lastTransitionTime,omitempty"`
}

// jobItem is an abstract representation ofj a job that only parses the files that we care about
type jobItem struct {
	Status struct {
		Succeeded      int         `yaml:"succeeded"`
		Failed         int         `yaml:"failed"`
		CompletionTime ocTime      `yaml:"completionTime,omitempty"`
		Conditions     []condition `yaml:"conditions"`
	} `yaml:"status"`
	Metadata struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
}

// loadConfig reads a file from path and returns a pointer to a configuration object, and error if one occurred
func loadConfig(path string) (*config, error) {
	config := new(config)
	cfgFile, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(cfgFile, config)
	return config, err
}

// namespaceBlacklisted returns true if the project string passed in exists in c.Blacklist
func (c *config) namespaceBlacklisted(project string) bool {
	for _, v := range c.Blacklist {
		if v == project {
			return true
		}
	}
	return false
}

// getDescriptorForJob returns a descriptor to be used for the currently active job, if no custom namespace is supplied it will use the defaults
// this should be called after you have checked if the namespace is in a blacklist
func (c *config) getJobDescriptor(namespace string) *namespaceCleanupDescriptor {
	descriptor := &c.Default
	if val, ok := c.Namespaces[namespace]; ok {
		descriptor = &val
	}
	return descriptor
}
