package main

import libconfig "github.com/opensourceways/community-robot-lib/config"

type configuration struct {
	Checkpr []pluginConfig `json:"checkpr,omitempty"`
}

func (c *configuration) Validate() error {
	if c != nil {
		cs := c.Checkpr
		for i := range cs {
			if err := cs[i].Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *configuration) CheckPRFor(org, repo string) *pluginConfig {
	if c == nil {
		return nil
	}
	cs := c.Checkpr
	v := make([]libconfig.IPluginForRepo, 0, len(cs))
	for i := range cs {
		v = append(v, &cs[i])
	}
	if i := libconfig.FindConfig(org, repo, v); i >= 0 {
		return &cs[i]
	}
	return nil
}

func (c *configuration) SetDefault() {
	if c != nil {
		cs := c.Checkpr
		for i := range cs {
			cs[i].setDefault()
		}
	}
}

type pluginConfig struct {
	libconfig.PluginForRepo

	// NoNeedResetReviewerTester whether to reset the number of reviewers and  testers when the PR is turned on
	NoNeedResetReviewerTester bool `json:"no_need_reset_reviewer_tester,omitempty"`

	// CommitsThreshold Check the threshold of the number of PR commits,
	// and add the label specified by SquashCommitLabel to the PR if this value is exceeded.
	// zero means no check.
	CommitsThreshold int `json:"commits_threshold,omitempty"`

	// SquashCommitLabel Specify the label whose PR exceeds the threshold. default: stat/needs-squash
	SquashCommitLabel string `json:"squash_commit_label,omitempty"`
}

func (c *pluginConfig) setDefault() {
	if c.SquashCommitLabel == "" {
		c.SquashCommitLabel = "stat/needs-squash"
	}
}

func (c pluginConfig) needResetReviewerAndTester() bool {
	return !c.NoNeedResetReviewerTester
}

func (c pluginConfig) needCheckCommits() bool {
	return c.CommitsThreshold > 0
}
