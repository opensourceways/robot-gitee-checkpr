package main

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/opensourceways/robot-gitee-plugin-lib/giteeclient"
	"github.com/opensourceways/robot-gitee-plugin-lib/logrusutil"
	liboptions "github.com/opensourceways/robot-gitee-plugin-lib/options"
	libplugin "github.com/opensourceways/robot-gitee-plugin-lib/plugin"
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
	logrusutil.ComponentInit()

	o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	if err := o.Validate(); err != nil {
		logrus.WithError(err).Fatal("Invalid options")
	}

	token, err := ioutil.ReadFile(o.gitee.TokenPath)
	if err != nil {
		logrus.WithError(err).Fatal("Invalid token path")
	}
	c := giteeclient.NewClient(func() []byte {
		return token
	})

	libplugin.Run(newCheckPr(c), o.plugin)
}
