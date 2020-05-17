package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRepoConfigurationFind(t *testing.T) {
	findTests := []struct {
		name string
		want *Repository
	}{
		{"testing", &Repository{Name: "testing"}},
		{"unknown", nil},
	}

	cfgs := RepoConfiguration{
		Repositories: []*Repository{
			{Name: "testing"},
			{Name: "another"},
		},
	}

	for _, tt := range findTests {
		if diff := cmp.Diff(tt.want, cfgs.Find(tt.name)); diff != "" {
			t.Errorf("Find(%s) failed:\n %s", tt.name, diff)
		}
	}
}

func TestParse(t *testing.T) {
	parseTests := []struct {
		filename string
		want     *RepoConfiguration
	}{
		{
			"testdata/config.yaml", &RepoConfiguration{
				Repositories: []*Repository{
					{
						Name:               "testing/repo-image",
						SourceRepo:         "example/example-source",
						SourceBranch:       "master",
						FilePath:           "test/file.yaml",
						UpdateKey:          "person.name",
						BranchGenerateName: "repo-imager-",
					},
				},
			},
		},
	}

	for _, tt := range parseTests {
		t.Run(fmt.Sprintf("parsing %s", tt.filename), func(rt *testing.T) {
			f, err := os.Open(tt.filename)
			if err != nil {
				rt.Errorf("failed to open %v: %s", tt.filename, err)
			}
			defer f.Close()

			got, err := Parse(f)
			if err != nil {
				rt.Errorf("failed to parse %v: %s", tt.filename, err)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				rt.Errorf("Parse(%s) failed diff\n%s", tt.filename, diff)
			}
		})
	}
}
