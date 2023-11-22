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

func TestTransformGiteeId(t *testing.T) {
	bot := robot{
		om: new(omServiceTest),
	}

	testCase := []string{"tom-gitee", "I-am-a-robot", "xxxxxxx"}
	githubId := bot.transformGiteeId(testCase)

	if len(githubId) != 1 || githubId[0] != "tom-github" {
		t.Fail()
	}
}

type omServiceTest struct {
}

func (o *omServiceTest) GetToken() (string, error) {
	return "xxxxxxxxxx", nil
}

func (o *omServiceTest) GetUserInfo(giteeId string) ([]Identities, error) {
	_, err := o.GetToken()
	if err != nil {
		return nil, err
	}

	switch giteeId {
	case "tom-gitee":
		return []Identities{
			{
				LoginName: "tom-github",
				Identity:  "github",
			},
			{
				LoginName: "tom-gitee",
				Identity:  "gitee",
			},
		}, nil
	case "I-am-a-robot":
		return []Identities{
			{
				LoginName: "I-am-a-robot",
				Identity:  "gitee",
			},
		}, nil

	default:
		return nil, nil
	}
}
