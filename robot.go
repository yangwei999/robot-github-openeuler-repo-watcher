package main

import (
	sdk "github.com/google/go-github/v36/github"
	gc "github.com/opensourceways/community-robot-lib/githubclient"
	"sync"

	"github.com/panjf2000/ants/v2"
)

const botName = "repo-watcher"

type iClient interface {
	SetProtectionBranch(org, repo, branch string, pre *sdk.ProtectionRequest) error
	GetDirectoryTree(org, repo, branch string, recursive bool) ([]*sdk.TreeEntry, error)
	GetPathContent(org, repo, path, branch string) (*sdk.RepositoryContent, error)
	CreateFile(org, repo, path, branch, commitMSG, sha string, content []byte) error
	GetRepos(org string) ([]*sdk.Repository, error)
	GetRepo(org, repo string) (*sdk.Repository, error)
	CreateRepo(org string, r *sdk.Repository) error
	UpdateRepo(org, repo string, r *sdk.Repository) error
	ListCollaborator(pr gc.PRInfo) ([]*sdk.User, error)
	RemoveProtectionBranch(org, repo, branch string) error
	GetRef(org, repo, ref string) (*sdk.Reference, error)
	CreateBranch(org, repo string, reference *sdk.Reference) error
	ListBranches(org, repo string) ([]*sdk.Branch, error)
	RemoveRepoMember(pr gc.PRInfo, login string) error
	AddRepoMember(pr gc.PRInfo, login, permission string) error
}

func newRobot(cli iClient, pool *ants.Pool, cfg *botConfig) *robot {
	return &robot{cli: cli, pool: pool, cfg: cfg}
}

type robot struct {
	pool *ants.Pool
	cfg  *botConfig
	cli  iClient
	wg   sync.WaitGroup
}
