package main

import (
	"testing"

	"github.com/opensourceways/robot-github-openeuler-repo-watcher/community"
)

func TestCanProcessWithEmptyUrl(t *testing.T) {
	expect := expectRepoInfo{
		org: "src-openeuler",
		expectRepoState: &community.Repository{
			Name: "test",
		},
	}

	if CanProcess(expect) {
		t.Fail()
	}
}

func TestCanProcessWithGiteeUrl(t *testing.T) {
	expect := expectRepoInfo{
		org: "src-openeuler",
		expectRepoState: &community.Repository{
			Name:    "test",
			RepoUrl: "https://gitee.com/src-openeuler/test",
		},
	}

	if CanProcess(expect) {
		t.Fail()
	}
}

func TestCanProcessWithGithubUrl(t *testing.T) {
	expect := expectRepoInfo{
		org: "src-openeuler",
		expectRepoState: &community.Repository{
			Name:    "test",
			RepoUrl: "https://github.com/src-openeuler/test",
		},
	}

	if !CanProcess(expect) {
		t.Fail()
	}
}

func TestCanProcessWithInvalidString(t *testing.T) {
	expect := expectRepoInfo{
		org: "src-openeuler",
		expectRepoState: &community.Repository{
			Name:    "test",
			RepoUrl: "gsdgsdggdfgdfg",
		},
	}

	if CanProcess(expect) {
		t.Fail()
	}
}
