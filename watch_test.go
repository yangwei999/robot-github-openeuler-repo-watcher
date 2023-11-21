package main

import (
	"strconv"
	"testing"

	"github.com/opensourceways/robot-github-openeuler-repo-watcher/community"
)

var bot = robot{
	cfg: &botConfig{
		OMApi: OMApi{
			Endpoint:  "https://omapi.osinfra.cn",
			AppId:     "xxxx",
			AppSecret: "xxxx",
		},
	},
}

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

func TestGetToken(t *testing.T) {
	_, err := bot.getToken()
	if err != nil {
		t.Error(err)
	}
}

func TestGetUserInfo(t *testing.T) {
	testCase := []string{"georgecao", "I-am-a-robot", "xxxxxxx"}
	for k, v := range testCase {
		_, err := bot.getUserInfo(v)
		if err != nil {
			t.Errorf("case num %d failed:%s", k, err.Error())
		}
	}
}

func TestTransformGiteeId(t *testing.T) {
	testCase := []string{"georgecao", "I-am-a-robot", "xxxxxxx"}
	githubId := bot.transformGiteeId(testCase)

	if len(githubId) != 1 {
		t.Fail()
	}
}
