package main

import (
	"fmt"
	"strings"

	gc "github.com/opensourceways/robot-github-lib/client"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (bot *robot) handleMember(expectRepo expectRepoInfo, localMembers, localAdmins []string, repoOwner *string, log *logrus.Entry) ([]string, []string) {
	org := expectRepo.org
	repo := expectRepo.getNewRepoName()

	if len(localMembers) == 0 {
		ms, err := bot.cli.ListCollaborator(gc.PRInfo{Org: org, Repo: repo})
		members := make([]string, 0)
		if err != nil || len(ms) == 0 {
			log.Errorf("handle repo members and get repo:%s, err:%s", repo, err.Error())
			return nil, nil
		}

		for _, n := range ms {
			members = append(members, *n.Login)
		}
		localMembers = toLowerOfMembers(members)
		*repoOwner = owner
	}

	expect := sets.NewString(expectRepo.expectOwners...)
	lm := sets.NewString(localMembers...)
	r := expect.Intersection(lm).UnsortedList()

	expectAdmins := toLowerOfMembers(expectRepo.expectAdmins)
	if len(expectAdmins) > 0 {
		if len(localAdmins) == 0 {
			allCollaborators, err := bot.cli.ListCollaborator(gc.PRInfo{Org: org, Repo: repo})
			if err != nil {
				log.Errorf("list %s's all collaborators failed, err: %v", repo, err)
			}

			for _, item := range allCollaborators {
				for per := range item.Permissions {
					if per == "admin" && item.Permissions[per] == true {
						localAdmins = append(localAdmins, strings.ToLower(*item.Login))
					}
				}
			}
		}
	}
	ea := sets.NewString(expectAdmins...)
	la := sets.NewString(localAdmins...)
	a := ea.Intersection(la).UnsortedList()

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

	// add maintain
	if v := ea.Difference(la); v.Len() > 0 {
		for k := range v {
			if !expect.Has(k) {
				continue
			}

			if expect.Has(k) {

				l := log.WithField("update developer to admin", fmt.Sprintf("%s:%s", repo, k))
				l.Info("start")

				if err := bot.cli.RemoveRepoMember(gc.PRInfo{Org: org, Repo: repo}, k); err != nil {
					l.Errorf("remove developer %s from %s/%s failed, err: %v", k, org, repo, err)
				}

				if err := bot.addRepoAdmin(org, repo, k); err != nil {
					l.Errorf("add admin %s to %s/%s failed, err: %v", k, org, repo, err)
				} else {
					a = append(a, k)
				}
			}
		}
	}

	// remove maintain
	if v := la.Difference(ea); v.Len() > 0 {
		o := *repoOwner

		if o == "" {
			v, err := bot.cli.GetRepo(org, repo)
			if err != nil {
				log.Errorf("handle repo members and get repo:%s, err:%s", repo, err.Error())
				return nil, nil
			}
			*repoOwner = *v.Owner.Login
			o = *repoOwner
		}

		for k := range v {

			if k == o || k == "openeuler-ci-bot" {
				continue
			}

			if expect.Has(k) {

				l := log.WithField("update admin to developer", fmt.Sprintf("%s:%s", repo, k))
				l.Info("start")

				if err := bot.cli.RemoveRepoMember(gc.PRInfo{Org: org, Repo: repo}, k); err != nil {
					l.Errorf("remove admin %s from %s/%s failed, err: %v", k, org, repo, err)
					a = append(a, k)
				}

				if err := bot.addRepoMember(org, repo, k); err != nil {
					l.Errorf("add developer %s to %s/%s failed, err: %v", k, org, repo, err)
				}
			}
		}
	}

	return r, a
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
