package main

import (
	"strconv"
	"testing"

	"github.com/opensourceways/robot-github-openeuler-repo-watcher/community"
)

func TestCanProcess(t *testing.T) {
	testCase := [][]string{
		{"", "github", "true"},
		{"", "gitee", "false"},
		{"xxx", "github", "false"},
		{"xxx", "gitee", "false"},
		{"xxx", "", "false"},
		{"", "", "false"},
	}

	for k, v := range testCase {
		expect := expectRepoInfo{
			org: "src-openeuler",
			expectRepoState: &community.Repository{
				Name:     "test",
				RepoUrl:  v[0],
				Platform: v[1],
			},
		}

		if strconv.FormatBool(CanProcess(expect)) != v[2] {
			t.Errorf("case num %d failed", k)
		}
	}
}
