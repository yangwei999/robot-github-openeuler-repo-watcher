package main

import (
	"encoding/base64"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"path"
	"strings"
	"sync"
	"time"
)

type yamlStruct struct {
	Packages []PackageInfo `json:"packages,omitempty"`
}

type PackageInfo struct {
	Name     string `json:"name,omitempty"`
	Obs_From string `json:"obs_from,omitempty"`
	Obs_To   string `json:"obs_to,omitempty"`
	Date     string `json:"date,omitempty"`
}

var m sync.Mutex

func (bot *robot) patchFactoryYaml(repo string, log *logrus.Entry) {

	if !bot.cfg.EnableCreatingOBSMetaProject {
		return
	}

	m.Lock()
	defer m.Unlock()
	var y yamlStruct

	project := &bot.cfg.OBSMetaProject
	readingPath := path.Join(project.ProjectDir, project.ProjectFileName)
	b := &project.Branch

	f, err := bot.cli.GetPathContent(b.Org, b.Repo, readingPath, b.Branch)
	if err != nil {
		log.Errorf("get file %s failed.", readingPath)
		return
	}

	c, err := base64.StdEncoding.DecodeString(*f.Content)
	if err != nil {
		return
	}

	if err = yaml.Unmarshal(c, &y); err != nil {
		return
	}

	var p PackageInfo
	p.Name = repo
	p.Obs_To = "openEuler:Factory"
	year, month, day := time.Now().Format("2006"), time.Now().Format("01"), time.Now().Format("02")
	p.Date = fmt.Sprintf("%s-%s-%s", year, month, day)
	y.Packages = append(y.Packages, p)

	by, err := yaml.Marshal(&y)
	if err != nil {
		return
	}

	message := fmt.Sprintf("a new series of repositories has been created")
	err = bot.cli.CreateFile(b.Org, b.Repo, readingPath, b.Branch, message, *f.SHA, by)
	if err != nil {
		log.Errorf("update file failed %v", err)
		return
	}
}

func (bot *robot) getRefresh(log *logrus.Entry) []PackageInfo {
	project := &bot.cfg.OBSMetaProject
	b := &project.Branch

	// get dir name
	tree, err := bot.cli.GetDirectoryTree(b.Org, b.Repo, "master", true)
	if err != nil {
		log.Errorf("list dirs in release-management/master failed, %v", err)
		return nil
	}

	var allPackages []PackageInfo

	for _, t := range tree {
		for p, s := range PckgShaMap {
			if *t.Path == p && s != t.GetSHA() {
				PckgShaMap[p] = t.GetSHA()

				var y yamlStruct
				f, err := bot.cli.GetPathContent(b.Org, b.Repo, *t.Path, b.Branch)
				if err != nil {
					log.Errorf("get file %s failed. error is: %v", t.Path, err)
					continue
				}

				c, err := base64.StdEncoding.DecodeString(*f.Content)
				if err != nil {
					continue
				}

				if err = yaml.Unmarshal(c, &y); err != nil {
					continue
				}

				AllPackagesInPckg[p] = y.Packages
			}
		}

	}

	for _, p := range AllPackagesInPckg {
		allPackages = append(allPackages, p...)
	}

	return allPackages
}

var AllPackagesInPckg map[string][]PackageInfo
var PckgShaMap map[string]string

func (bot *robot) loadAllPckgMgmtFile() error {
	AllPackagesInPckg = map[string][]PackageInfo{}
	PckgShaMap = map[string]string{}
	project := &bot.cfg.OBSMetaProject
	b := &project.Branch

	// get tree
	tree, err := bot.gecli.GetDirectoryTree(b.Org, b.Repo, "master", 1)
	if err != nil {
		return err
	}

	for _, t := range tree.Tree {
		if strings.Contains(t.Path, "openEuler") && strings.HasPrefix(t.Path, "master/") && strings.HasSuffix(t.Path, project.ProjectFileName) {
			PckgShaMap[t.Path] = t.Sha

			var y yamlStruct
			f, err := bot.gecli.GetPathContent(b.Org, b.Repo, t.Path, b.Branch)
			if err != nil {
				continue
			}

			c, err := base64.StdEncoding.DecodeString(f.Content)
			if err != nil {
				continue
			}

			if err = yaml.Unmarshal(c, &y); err != nil {
				continue
			}

			var allPackages []PackageInfo
			for _, r := range y.Packages {
				allPackages = append(allPackages, r)
			}

			AllPackagesInPckg[t.Path] = allPackages
		}
	}

	return nil
}
