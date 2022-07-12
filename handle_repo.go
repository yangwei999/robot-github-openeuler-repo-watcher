package main

import (
	"fmt"
	gc "github.com/opensourceways/community-robot-lib/githubclient"

	sdk "github.com/google/go-github/v36/github"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/robot-github-repo-watcher/community"
	"github.com/opensourceways/robot-github-repo-watcher/models"
)

func (bot *robot) createRepo(
	expectRepo expectRepoInfo,
	log *logrus.Entry,
	hook func(string, *logrus.Entry),
) models.RepoState {
	org := expectRepo.org
	repo := expectRepo.expectRepoState
	repoName := expectRepo.getNewRepoName()

	if n := repo.RenameFrom; n != "" && n != repoName {
		return bot.renameRepo(expectRepo, log, hook)
	}

	log = log.WithField("create repo", repoName)
	log.Info("start")

	property, err := bot.newRepo(org, repo)
	if err != nil {
		log.Warning("repo exists already")

		if s, b := bot.getRepoState(org, repoName, log); b {
			s.Branches = bot.handleBranch(expectRepo, s.Branches, log)
			s.Members = bot.handleMember(expectRepo, s.Members, &s.Owner, log)
			return s
		}

		log.Errorf("create repo, err:%s", err.Error())

		return models.RepoState{}
	}

	defer func() {
		hook(repoName, log)
	}()

	branches, members := bot.initNewlyCreatedRepo(
		org, repoName, repo.Branches, expectRepo.expectOwners, log,
	)

	return models.RepoState{
		Available: true,
		Branches:  branches,
		Members:   members,
		Property:  property,
	}
}

func (bot *robot) newRepo(org string, repo *community.Repository) (models.RepoProperty, error) {
	has := true
	private := repo.IsPrivate()
	err := bot.cli.CreateRepo(org, &sdk.Repository{
		Name:        &repo.Name,
		Description: &repo.Description,
		HasIssues:   &has,
		HasWiki:     &has,
		AutoInit:    &has, // set `auto_init` as true to initialize `master` branch with README after repo creation
		Private:     &private,
	})
	if err != nil {
		return models.RepoProperty{}, err
	}

	//err = bot.cli.AddProjectLabels(org, repo.Name, []string{sigLabel})
	//if err != nil {
	//	log.Infof("Add project label %s failed, err: %v", sigLabel, err)
	//}

	return models.RepoProperty{
		CanComment: repo.Commentable,
		Private:    repo.IsPrivate(),
	}, nil
}

func (bot *robot) initNewlyCreatedRepo(
	org, repoName string,
	repoBranches []community.RepoBranch,
	repoOwners []string,
	log *logrus.Entry,
) ([]community.RepoBranch, []string) {
	//if err := bot.initRepoReviewer(org, repoName); err != nil {
	//	log.Errorf("initialize the reviewers, err:%s", err.Error())
	//}

	branches := []community.RepoBranch{
		{Name: community.BranchMaster},
	}
	for _, item := range repoBranches {
		if item.Name == community.BranchMaster {
			if item.Type != community.BranchProtected {
				continue
			}

			if err := bot.updateBranch(org, repoName, item.Name, true); err == nil {
				branches[0].Type = community.BranchProtected
			} else {
				log.WithFields(logrus.Fields{
					"update branch": fmt.Sprintf("%s/%s", repoName, item.Name),
					"type":          item.Type,
				}).Error(err)
			}
		} else {
			if b, ok := bot.createBranch(org, repoName, item, log); ok {
				branches = append(branches, b)
			}
		}
	}

	members := []string{}
	for _, item := range repoOwners {
		if err := bot.addRepoMember(org, repoName, item); err != nil {
			log.Errorf("add member:%s, err:%s", item, err)
		} else {
			members = append(members, item)
		}
	}

	return branches, members
}

func (bot *robot) renameRepo(
	expectRepo expectRepoInfo,
	log *logrus.Entry,
	hook func(string, *logrus.Entry),
) models.RepoState {
	org := expectRepo.org
	oldRepo := expectRepo.expectRepoState.RenameFrom
	newRepo := expectRepo.getNewRepoName()

	log = log.WithField("rename repo", fmt.Sprintf("from %s to %s", oldRepo, newRepo))
	log.Info("start")

	err := bot.cli.UpdateRepo(
		org,
		oldRepo,
		&sdk.Repository{
			Name:        &newRepo,
			Description: &expectRepo.expectRepoState.Description,
		},
	)

	//err = bot.cli.UpdateProjectLabels(org, newRepo, []string{sigLabel})
	//if err != nil {
	//	log.Infof("update label failed: %v", err)
	//}

	defer func(b bool) {
		if b {
			hook(newRepo, log)
		}
	}(err == nil)

	// if the err == nil, invoke 'getRepoState' obviously.
	// if the err != nil, it is better to call 'getRepoState' to
	// avoid the case that the repo already exists.
	if s, b := bot.getRepoState(org, newRepo, log); b {
		s.Branches = bot.handleBranch(expectRepo, s.Branches, log)
		s.Members = bot.handleMember(expectRepo, s.Members, &s.Owner, log)
		return s
	}

	if err != nil {
		log.Error(err)

		return models.RepoState{}
	}

	return models.RepoState{Available: true}
}

func (bot *robot) getRepoState(org, repo string, log *logrus.Entry) (models.RepoState, bool) {
	newRepo, err := bot.cli.GetRepo(org, repo)
	if err != nil {
		log.Errorf("get repo, err:%s", err.Error())

		return models.RepoState{}, false
	}
	//list repo's members
	ms, err := bot.cli.ListCollaborator(gc.PRInfo{Org: org, Repo: repo})
	members := make([]string, 0)
	if err != nil || len(ms) == 0 {
		r := models.RepoState{
			Available: true,
			Members:   members,
			Property: models.RepoProperty{
				Private: *newRepo.Private,
			},
		}

		branches, err := bot.listAllBranchOfRepo(org, repo)
		if err != nil {
			log.Errorf("list branch, err:%s", err.Error())
		} else {
			r.Branches = branches
		}

		return r, true
	}

	for _, m := range ms {
		members = append(members, *m.Login)
	}

	r := models.RepoState{
		Available: true,
		Members:   toLowerOfMembers(members),
		Property: models.RepoProperty{
			Private: *newRepo.Private,
		},
	}

	branches, err := bot.listAllBranchOfRepo(org, repo)
	if err != nil {
		log.Errorf("list branch, err:%s", err.Error())
	} else {
		r.Branches = branches
	}

	return r, true
}

//func (bot *robot) initRepoReviewer(org, repo string) error {
//	return bot.cli.SetRepoReviewer(
//		org,
//		repo,
//		sdk.SetRepoReviewer{
//			Assignees:       " ", // This parameter is a required one according to the Gitee API
//			Testers:         " ", // Ditto
//			AssigneesNumber: 0,
//			TestersNumber:   0,
//		},
//	)
//}

func (bot *robot) updateRepo(expectRepo expectRepoInfo, lp models.RepoProperty, log *logrus.Entry) models.RepoProperty {
	org := expectRepo.org
	repo := expectRepo.expectRepoState
	repoName := expectRepo.getNewRepoName()

	ep := repo.IsPrivate()

	if ep != lp.Private {
		log = log.WithField("update repo", repoName)
		log.Info("start")

		err := bot.cli.UpdateRepo(
			org,
			repoName,
			&sdk.Repository{
				Name:    &repoName,
				Private: &ep,
			},
		)
		if err == nil {
			return models.RepoProperty{
				Private: ep,
			}
		}

		log.WithFields(logrus.Fields{
			"Private": ep,
		}).Error(err)
	}
	return lp
}
