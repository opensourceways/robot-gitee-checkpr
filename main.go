package main

import (
	"flag"
	"os"

	"github.com/opensourceways/robot-gitee-plugin-lib/giteeclient"
	"github.com/opensourceways/robot-gitee-plugin-lib/logrusutil"
	liboptions "github.com/opensourceways/robot-gitee-plugin-lib/options"
	libplugin "github.com/opensourceways/robot-gitee-plugin-lib/plugin"
	"github.com/opensourceways/robot-gitee-plugin-lib/secret"
	"github.com/sirupsen/logrus"
)

type options struct {
	plugin liboptions.PluginOptions
	gitee  liboptions.GiteeOptions
}

func (o *options) Validate() error {
	if err := o.plugin.Validate(); err != nil {
		return err
	}

	return o.gitee.Validate()
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	o.gitee.AddFlags(fs)
	o.plugin.AddFlags(fs)

	fs.Parse(args)
	return o
}

func main() {
	logrusutil.ComponentInit(pluginName)

	o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	if err := o.Validate(); err != nil {
		logrus.WithError(err).Fatal("Invalid options")
	}

	secretAgent := new(secret.Agent)
	if err := secretAgent.Start([]string{o.gitee.TokenPath}); err != nil {
		logrus.WithError(err).Fatal("Error starting secret agent.")
	}

	c := giteeclient.NewClient(secretAgent.GetTokenGenerator(o.gitee.TokenPath))

	p := newCheckPr(c, func() { secretAgent.Stop() })

	libplugin.Run(p, o.plugin)
}
