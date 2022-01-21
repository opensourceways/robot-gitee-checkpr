package main

import (
	"errors"
	"fmt"

	"github.com/opensourceways/community-robot-lib/config"
	"github.com/opensourceways/community-robot-lib/robot-gitee-framework"
	sdk "github.com/opensourceways/go-gitee/gitee"
	"github.com/sirupsen/logrus"
)

const botName = "checkpr"

type iClient interface {
	UpdatePullRequest(org, repo string, number int32, param sdk.PullRequestUpdateParam) (sdk.PullRequest, error)
}

func newRobot(cli iClient) *robot {
	return &robot{cli: cli}
}

type robot struct {
	cli iClient
}

func (bot *robot) NewConfig() config.Config {
	return &configuration{}
}

func (bot *robot) getConfig(cfg config.Config) (*configuration, error) {
	if c, ok := cfg.(*configuration); ok {
		return c, nil
	}
	return nil, errors.New("can't convert to configuration")
}

func (bot *robot) RegisterEventHandler(f framework.HandlerRegitster) {
	f.RegisterPullRequestHandler(bot.handlePREvent)
}

func (bot *robot) handlePREvent(e *sdk.PullRequestEvent, c config.Config, log *logrus.Entry) error {
	if sdk.GetPullRequestAction(e) == sdk.ActionClose {
		return nil
	}

	config, err := bot.getConfig(c)
	if err != nil {
		return err
	}

	org, repo := e.GetOrgRepo()

	if config.configFor(org, repo) == nil {
		return fmt.Errorf("no config for this repo:%s/%s", org, repo)
	}

	pr := e.GetPullRequest()

	if !pr.GetNeedTest() && !pr.GetNeedReview() {
		return nil
	}

	n := int32(0)
	_, err = bot.cli.UpdatePullRequest(
		org, repo, pr.GetNumber(),
		sdk.PullRequestUpdateParam{
			AssigneesNumber: &n,
			TestersNumber:   &n,
		},
	)
	return err
}
