package config

import (
	"fmt"
	"io"
	"io/ioutil"

	"sigs.k8s.io/yaml"
)

// Repository is the items that are requird to update a specific file in a repo.
type Repository struct {
	Name               string `json:"name"`
	SourceRepo         string `json:"sourceRepo"`
	SourceBranch       string `json:"sourceBranch"`
	FilePath           string `json:"filePath`
	UpdateKey          string `json:"updateKey"`
	BranchGenerateName string `json:"branchGenerateName"`
}

func Parse(in io.Reader) (*RepoConfiguration, error) {
	body, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML: %w", err)
	}
	rc := &RepoConfiguration{}
	err = yaml.Unmarshal(body, rc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}
	return rc, nil
}

// RepoConfiguration is a slice of Repository values.
type RepoConfiguration struct {
	Repositories []*Repository `json:"repositories"`
}

func (c RepoConfiguration) Find(name string) *Repository {
	for _, cfg := range c.Repositories {
		if cfg.Name == name {
			return cfg
		}
	}
	return nil
}
