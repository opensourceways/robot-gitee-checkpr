package main

import (
	"errors"
	"fmt"

	sdk "gitee.com/openeuler/go-gitee/gitee"
	"github.com/opensourceways/robot-gitee-plugin-lib/giteeclient"
	libplugin "github.com/opensourceways/robot-gitee-plugin-lib/plugin"
	libutils "github.com/opensourceways/robot-gitee-plugin-lib/utils"
	"github.com/sirupsen/logrus"
)

type cpClient interface {
	UpdatePullRequest(org, repo string, number int32, param sdk.PullRequestUpdateParam) (sdk.PullRequest, error)
	GetGiteePullRequest(org, repo string, number int) (sdk.PullRequest, error)
	GetPRCommits(org, repo string, number int) ([]sdk.PullRequestCommits, error)
	AddPRLabel(org, repo string, number int, label string) error
	RemovePRLabel(org, repo string, number int, label string) error
}

type checkPr struct {
	ghc cpClient
}

func newCheckPr(gec cpClient) libplugin.Plugin {
	return &checkPr{gec}
}

func (cp *checkPr) Exit() {
}

func (cp *checkPr) PluginName() string {
	return "checkpr"
}

func (cp *checkPr) NewPluginConfig() libplugin.PluginConfig {
	return &configuration{}
}

func (cp *checkPr) RegisterEventHandler(p libplugin.HandlerRegitster) {
	p.RegisterPullRequestHandler(cp.handlePREvent)
}

func (cp *checkPr) handlePREvent(e *sdk.PullRequestEvent, cfg libplugin.PluginConfig, log *logrus.Entry) error {
	action := giteeclient.GetPullRequestAction(e)
	if action == giteeclient.PRActionClosed {
		return nil
	}

	config, err := cp.pluginConfig(cfg)
	if err != nil {
		return err
	}

	prInfo := giteeclient.GetPRInfoByPREvent(e)
	pc := config.CheckPRFor(prInfo.Org, prInfo.Repo)
	if pc == nil {
		return fmt.Errorf("no %s plugin config for this repo:%s/%s", cp.PluginName(), prInfo.Org, prInfo.Repo)
	}

	mr := libutils.NewMultiErrors()
	if err := cp.removeMinNumReviewerAndTester(prInfo, pc); err != nil {
		mr.AddError(err)
	}

	if action == giteeclient.PRActionOpened || action == giteeclient.PRActionChangedSourceBranch {
		if err := cp.handleCheckCommits(prInfo, pc); err != nil {
			mr.AddError(err)
		}

	}
	return mr.Err()
}

func (cp *checkPr) removeMinNumReviewerAndTester(prInfo giteeclient.PRInfo, cfg *pluginConfig) error {
	if !cfg.needResetReviewerAndTester() {
		return nil
	}

	org := prInfo.Org
	repo := prInfo.Repo
	number := prInfo.Number

	pr, err := cp.ghc.GetGiteePullRequest(org, repo, number)
	if err != nil {
		return err
	}
	if pr.AssigneesNumber == 0 && pr.TestersNumber == 0 {
		return nil
	}

	changeNum := int32(0)
	param := sdk.PullRequestUpdateParam{AssigneesNumber: &changeNum, TestersNumber: &changeNum}
	_, err = cp.ghc.UpdatePullRequest(org, repo, int32(number), param)
	return err
}

func (cp *checkPr) pluginConfig(cfg libplugin.PluginConfig) (*configuration, error) {
	c, ok := cfg.(*configuration)
	if !ok {
		return nil, errors.New("can't convert to configuration")
	}
	return c, nil
}

func (cp *checkPr) handleCheckCommits(prInfo giteeclient.PRInfo, cfg *pluginConfig) error {
	if !cfg.needCheckCommits() {
		return nil
	}

	commits, err := cp.ghc.GetPRCommits(prInfo.Org, prInfo.Repo, prInfo.Number)
	if err != nil {
		return err
	}

	exceeded := len(commits) > cfg.CommitsThreshold
	hasSquashLabel := prInfo.HasLabel(cfg.SquashCommitLabel)

	if exceeded && !hasSquashLabel {
		return cp.ghc.AddPRLabel(prInfo.Org, prInfo.Repo, prInfo.Number, cfg.SquashCommitLabel)
	}

	if !exceeded && hasSquashLabel {
		return cp.ghc.RemovePRLabel(prInfo.Org, prInfo.Repo, prInfo.Number, cfg.SquashCommitLabel)
	}

	return nil
}
