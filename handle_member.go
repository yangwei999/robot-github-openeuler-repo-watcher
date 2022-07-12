package main

import (
	"fmt"
	gc "github.com/opensourceways/community-robot-lib/githubclient"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (bot *robot) handleMember(expectRepo expectRepoInfo, localMembers []string, repoOwner *string, log *logrus.Entry) []string {
	org := expectRepo.org
	repo := expectRepo.getNewRepoName()

	if len(localMembers) == 0 {
		ms, err := bot.cli.ListCollaborator(gc.PRInfo{Org: org, Repo: repo})
		members := make([]string, 0)
		if err != nil || len(ms) == 0 {
			log.Errorf("handle repo members and get repo:%s, err:%s", repo, err.Error())
			return nil
		}

		for _, n := range ms {
			members = append(members, *n.Login)
		}
		localMembers = toLowerOfMembers(members)
	}

	expect := sets.NewString(expectRepo.expectOwners...)
	lm := sets.NewString(localMembers...)
	r := expect.Intersection(lm).UnsortedList()

	// add new
	if v := expect.Difference(lm); v.Len() > 0 {
		for k := range v {
			l := log.WithField("add member", fmt.Sprintf("%s:%s", repo, k))
			l.Info("start")

			// how about adding a member but he/she exits? see the comment of 'addRepoMember'
			if err := bot.addRepoMember(org, repo, k); err != nil {
				l.Error(err)
			} else {
				r = append(r, k)
			}
		}
	}

	// remove
	if v := lm.Difference(expect); v.Len() > 0 {
		o := *repoOwner

		for k := range v {
			if k == o {
				// Gitee does not allow to remove the repo owner.
				continue
			}

			l := log.WithField("remove member", fmt.Sprintf("%s:%s", repo, k))
			l.Info("start")

			if err := bot.cli.RemoveRepoMember(gc.PRInfo{Org: org, Repo: repo}, k); err != nil {
				l.Error(err)

				r = append(r, k)
			}
		}
	}

	return r
}

// Gitee api will be successful even if adding a member repeatedly.
func (bot *robot) addRepoMember(org, repo, login string) error {
	return bot.cli.AddRepoMember(gc.PRInfo{Org: org, Repo: repo}, login, "push")
}

func (bot *robot) addRepoAdmin(org, repo, login string) error {
	return bot.cli.AddRepoMember(gc.PRInfo{Org: org, Repo: repo}, login, "maintain")
}

func toLowerOfMembers(m []string) []string {
	v := make([]string, len(m))
	for i := range m {
		v[i] = strings.ToLower(m[i])
	}
	return v
}
