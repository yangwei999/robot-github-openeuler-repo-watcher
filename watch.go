package main

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/opensourceways/robot-github-openeuler-repo-watcher/community"
	"github.com/opensourceways/robot-github-openeuler-repo-watcher/models"
)

type expectRepoInfo struct {
	expectRepoState *community.Repository
	expectOwners    []string
	expectAdmins    []string
	org             string
}

func (e *expectRepoInfo) getNewRepoName() string {
	return e.expectRepoState.Name
}

func (bot *robot) run(ctx context.Context, log *logrus.Entry) error {
	w := &bot.cfg.WatchingFiles
	expect := &expectState{
		w:         w.repoBranch,
		log:       log,
		cli:       bot.cli,
		gecli:     bot.gecli,
		sigOwners: make(map[string]*expectSigOwners),
		sigInfos:  make(map[string]*expectSigInfos),
	}

	org, err := expect.init(w.RepoOrg, w.SigFilePath, w.SigDir)
	if err != nil {
		return err
	}

	local, err := bot.loadALLRepos(org)
	if err != nil {
		log.Errorf("Load repos of org(%s) failed, err:%s", org, err.Error())
		return err
	}

	err = bot.loadAllPckgMgmtFile()
	if err != nil {
		log.Errorf("load all pckg-mgmt.yaml failed, err:%s", err.Error())
	}

	bot.watch(ctx, org, local, expect)
	return nil
}

func (bot *robot) watch(ctx context.Context, org string, local *localState, expect *expectState) {
	if interval := bot.cfg.Interval; interval <= 0 {
		for {
			if isCancelled(ctx) {
				break
			}

			bot.checkOnce(ctx, org, local, expect)
		}
	} else {
		t := time.Duration(interval) * time.Minute

		for {
			if isCancelled(ctx) {
				break
			}

			s := time.Now()

			bot.checkOnce(ctx, org, local, expect)

			e := time.Now()
			if v := e.Sub(s); v < t {
				time.Sleep(t - v)
			}
		}
	}

	bot.wg.Wait()
}

func (bot *robot) checkOnce(ctx context.Context, org string, local *localState, expect *expectState) {
	f := func(repo *community.Repository, owners []string, admins []string, sigLabel string, log *logrus.Entry) {
		if repo == nil {
			return
		}
		cpo := make([]string, len(owners))
		if len(owners) > 0 {
			copy(cpo, owners)
		}

		cpa := make([]string, len(admins))
		if len(admins) > 0 {
			copy(cpa, admins)
		}

		e := expectRepoInfo{
			org:             org,
			expectOwners:    cpo,
			expectAdmins:    cpa,
			expectRepoState: repo,
		}

		if !CanProcess(e) {
			return
		}

		err := bot.execTask(
			local.getOrNewRepo(repo.Name),
			e,
			sigLabel,
			log,
		)
		if err != nil {
			log.Errorf("submit task of repo:%s, err:%s", repo.Name, err.Error())
		}
	}

	isStopped := func() bool {
		return isCancelled(ctx)
	}

	expect.log.Info("new check")

	expect.check(org, isStopped, local.clear, f)
}

// check if the repo should be handle by github robot
func CanProcess(e expectRepoInfo) bool {
	// repository_url meas the repo was hosted on other platform, ignore it
	if e.expectRepoState.RepoUrl != "" {
		logrus.Infof("repo %s host on other platform will not be proceed", e.expectRepoState.RepoUrl)
		return false
	}

	if e.expectRepoState.Platform == "github" {
		logrus.Infof("github platform means hosted on github, will be proceed")
		return true
	}

	logrus.Infof("platform %s should't be proceed on github", e.expectRepoState.Platform)
	return false
}

func (bot *robot) execTask(localRepo *models.Repo, expectRepo expectRepoInfo, sigLabel string, log *logrus.Entry) error {
	f := func(before models.RepoState) models.RepoState {
		if !before.Available {
			return bot.createRepo(expectRepo, log, bot.patchFactoryYaml)
		}

		ms, as := bot.handleMember(expectRepo, before.Members, before.Admins, &before.Owner, log)

		return models.RepoState{
			Available: true,
			Branches:  bot.handleBranch(expectRepo, before.Branches, log),
			Members:   ms,
			Admins:    as,
			Property:  bot.updateRepo(expectRepo, before.Property, log),
			Owner:     before.Owner,
		}
	}

	bot.wg.Add(1)
	err := bot.pool.Submit(func() {
		defer bot.wg.Done()

		localRepo.Update(f)
	})
	if err != nil {
		bot.wg.Done()
	}
	return err
}

func isCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
