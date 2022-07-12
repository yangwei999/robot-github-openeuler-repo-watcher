package main

import (
	"encoding/base64"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"path"
	"sync"
	"time"
)

type yamlStruct struct {
	Packages []PackageInfo `json:"packages,omitempty"`
}

type PackageInfo struct {
	Name    string `json:"name,omitempty"`
	ObsFrom string `json:"obs_from,omitempty"`
	ObsTo   string `json:"obs_to,omitempty"`
	Date    string `json:"date,omitempty"`
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
	p.ObsFrom = " "
	p.ObsTo = "openEuler:Factory"
	year, month, day := time.Now().Format("2006"), time.Now().Format("01"), time.Now().Format("02")
	p.Date = fmt.Sprintf("%s-%s-%s", year, month, day)
	y.Packages = append(y.Packages, p)

	by, err := yaml.Marshal(&y)
	if err != nil {
		return
	}

	message := fmt.Sprintf("a new repository %s has been created", repo)
	err = bot.cli.CreateFile(b.Org, b.Repo, readingPath, b.Branch, message, *f.SHA, by)
	if err != nil {
		log.Errorf("update file failed %v", err)
		return
	}
}
