package main

import (
	"testing"

	"github.com/opensourceways/robot-github-openeuler-repo-watcher/community"
)

func TestCanProcessWithEmptyUrl(t *testing.T) {
	expect := expectRepoInfo{
		org: "src-openeuler",
		expectRepoState: &community.Repository{
			Name:     "test",
			Platform: "github",
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
			Name:     "test",
			Platform: "github",
			RepoUrl:  "https://gitee.com/src-openeuler/test",
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
			Name:     "test",
			Platform: "github",
			RepoUrl:  "https://github.com/src-openeuler/test",
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
			Name:     "test",
			Platform: "github",
			RepoUrl:  "gsdgsdggdfgdfg",
		},
	}

	if CanProcess(expect) {
		t.Fail()
	}
}

func TestCanProcessWithOtherPlatform(t *testing.T) {
	expect := expectRepoInfo{
		org: "src-openeuler",
		expectRepoState: &community.Repository{
			Name:     "test",
			Platform: "gitee",
			RepoUrl:  "https://github.com/src-openeuler/test",
		},
	}

	if CanProcess(expect) {
		t.Fail()
	}
}
