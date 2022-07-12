package main

import (
	gc "github.com/opensourceways/community-robot-lib/githubclient"
	"github.com/opensourceways/robot-github-repo-watcher/models"
)

type localState struct {
	repos map[string]*models.Repo
}

func (r *localState) getOrNewRepo(repo string) *models.Repo {
	if v, ok := r.repos[repo]; ok {
		return v
	}

	v := models.NewRepo(repo, models.RepoState{})
	r.repos[repo] = v

	return v
}

func (r *localState) clear(isExpectedRepo func(string) bool) {
	for k := range r.repos {
		if !isExpectedRepo(k) {
			delete(r.repos, k)
		}
	}
}

func (bot *robot) loadALLRepos(org string) (*localState, error) {
	items, err := bot.cli.GetRepos(org)
	if err != nil {
		return nil, err
	}

	r := localState{
		repos: make(map[string]*models.Repo),
	}

	for i := range items {
		item := items[i]
		org, repo := gc.GetOrgRepo(item)
		// get repo's members
		m := make([]string, 0)
		members, err := bot.cli.ListCollaborator(gc.PRInfo{Org: org, Repo: repo})
		if err != nil {
			r.repos[*item.Name] = models.NewRepo(*item.Name, models.RepoState{
				Available: true,
				Members:   m,
				Property: models.RepoProperty{
					Private: *item.Private,
				},
			})
		}
		for _, i := range members {
			m = append(m, *i.Login)
		}
		r.repos[*item.Name] = models.NewRepo(*item.Name, models.RepoState{
			Available: true,
			Members:   toLowerOfMembers(m),
			Property: models.RepoProperty{
				Private: *item.Private,
			},
		})
	}

	return &r, nil
}
