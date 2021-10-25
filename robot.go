package main

import (
	"errors"
	"fmt"

	sdk "gitee.com/openeuler/go-gitee/gitee"
	libconfig "github.com/opensourceways/community-robot-lib/config"
	"github.com/opensourceways/community-robot-lib/giteeclient"
	libplugin "github.com/opensourceways/community-robot-lib/giteeplugin"
	libutils "github.com/opensourceways/community-robot-lib/utils"
	"github.com/sirupsen/logrus"
)

const botName = "checkpr"

type iClient interface {
	UpdatePullRequest(org, repo string, number int32, param sdk.PullRequestUpdateParam) (sdk.PullRequest, error)
	GetGiteePullRequest(org, repo string, number int32) (sdk.PullRequest, error)
	GetPRCommits(org, repo string, number int32) ([]sdk.PullRequestCommits, error)
	AddPRLabel(org, repo string, number int32, label string) error
	RemovePRLabel(org, repo string, number int32, label string) error
}

func newRobot(cli iClient) *robot {
	return &robot{cli: cli}
}

type robot struct {
	cli iClient
}

func (bot *robot) NewPluginConfig() libconfig.PluginConfig {
	return &configuration{}
}

func (bot *robot) getConfig(cfg libconfig.PluginConfig) (*configuration, error) {
	if c, ok := cfg.(*configuration); ok {
		return c, nil
	}
	return nil, errors.New("can't convert to configuration")
}

func (bot *robot) RegisterEventHandler(p libplugin.HandlerRegitster) {
	p.RegisterPullRequestHandler(bot.handlePREvent)
}

func (bot *robot) handlePREvent(e *sdk.PullRequestEvent, cfg libconfig.PluginConfig, log *logrus.Entry) error {
	action := giteeclient.GetPullRequestAction(e)
	if action == giteeclient.PRActionClosed {
		return nil
	}

	config, err := bot.getConfig(cfg)
	if err != nil {
		return err
	}

	prInfo := giteeclient.GetPRInfoByPREvent(e)
	pc := config.configFor(prInfo.Org, prInfo.Repo)
	if pc == nil {
		return fmt.Errorf("no %s plugin config for this repo:%s/%s", botName, prInfo.Org, prInfo.Repo)
	}

	mr := libutils.NewMultiErrors()
	if err := bot.removeMinNumReviewerAndTester(prInfo, pc); err != nil {
		mr.AddError(err)
	}

	if action == giteeclient.PRActionOpened || action == giteeclient.PRActionChangedSourceBranch {
		if err := bot.handleCheckCommits(prInfo, pc); err != nil {
			mr.AddError(err)
		}

	}
	return mr.Err()
}

func (bot *robot) removeMinNumReviewerAndTester(prInfo giteeclient.PRInfo, cfg *botConfig) error {
	if !cfg.needResetReviewerAndTester() {
		return nil
	}

	org := prInfo.Org
	repo := prInfo.Repo
	number := prInfo.Number

	pr, err := bot.cli.GetGiteePullRequest(org, repo, number)
	if err != nil {
		return err
	}
	if pr.AssigneesNumber == 0 && pr.TestersNumber == 0 {
		return nil
	}

	changeNum := int32(0)
	param := sdk.PullRequestUpdateParam{AssigneesNumber: &changeNum, TestersNumber: &changeNum}
	_, err = bot.cli.UpdatePullRequest(org, repo, int32(number), param)
	return err
}

func (bot *robot) handleCheckCommits(prInfo giteeclient.PRInfo, cfg *botConfig) error {
	if !cfg.needCheckCommits() {
		return nil
	}

	commits, err := bot.cli.GetPRCommits(prInfo.Org, prInfo.Repo, prInfo.Number)
	if err != nil {
		return err
	}

	exceeded := len(commits) > cfg.CommitsThreshold
	hasSquashLabel := prInfo.HasLabel(cfg.SquashCommitLabel)

	if exceeded && !hasSquashLabel {
		return bot.cli.AddPRLabel(prInfo.Org, prInfo.Repo, prInfo.Number, cfg.SquashCommitLabel)
	}

	if !exceeded && hasSquashLabel {
		return bot.cli.RemovePRLabel(prInfo.Org, prInfo.Repo, prInfo.Number, cfg.SquashCommitLabel)
	}

	return nil
}
