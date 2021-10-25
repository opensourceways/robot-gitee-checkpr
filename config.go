package main

import libconfig "github.com/opensourceways/community-robot-lib/config"

type configuration struct {
	ConfigItems []botConfig `json:"config_items,omitempty"`
}

func (c *configuration) configFor(org, repo string) *botConfig {
	if c == nil {
		return nil
	}

	items := c.ConfigItems
	v := make([]libconfig.IPluginForRepo, len(items))
	for i := range items {
		v[i] = &items[i]
	}

	if i := libconfig.FindConfig(org, repo, v); i >= 0 {
		return &items[i]
	}
	return nil
}

func (c *configuration) Validate() error {
	if c == nil {
		return nil
	}

	items := c.ConfigItems
	for i := range items {
		if err := items[i].validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c *configuration) SetDefault() {
	if c == nil {
		return
	}

	Items := c.ConfigItems
	for i := range Items {
		Items[i].setDefault()
	}
}

type botConfig struct {
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

func (c *botConfig) setDefault() {
	if c.SquashCommitLabel == "" {
		c.SquashCommitLabel = "stat/needs-squash"
	}
}

func (c *botConfig) validate() error {
	return c.PluginForRepo.Validate()
}

func (c *botConfig) needResetReviewerAndTester() bool {
	return !c.NoNeedResetReviewerTester
}

func (c *botConfig) needCheckCommits() bool {
	return c.CommitsThreshold > 0
}
