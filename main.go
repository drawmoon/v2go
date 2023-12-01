package main

import (
	"main/cmd"
	"main/settings"
	"os"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var setting *settings.Setting
var err error

func init() {

	setting, err = settings.LoadSettings("config.json")
	if err != nil {
		panic(err)
	}

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
	log.SetLevel(func() log.Level {
		if setting.Verbose {
			return log.DebugLevel
		}
		return log.WarnLevel
	}())
}

func main() {
	app := &cobra.Command{
		Use:   "xrc",
		Short: "A xray client.",
		Long: `
__  ___ __ ___ 
\ \/ / '__/ __|
 >  <| | | (__ 
/_/\_\_|  \___|
A xray client.
		`,
	}
	app.AddCommand(
		&cobra.Command{
			Use:   "run",
			Short: "Start proxy",
			Run: func(c *cobra.Command, args []string) {
				cmd.Run(setting)
			},
		},
		&cobra.Command{
			Use:   "resub",
			Short: "Retrieve Subscriptions",
			Run: func(c *cobra.Command, args []string) {
				cmd.Resub(setting)
			},
		},
		&cobra.Command{
			Use:   "retest",
			Short: "Retest all available nodes",
			Run: func(c *cobra.Command, args []string) {
				cmd.Retest(setting)
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "List selected nodes",
			Run: func(c *cobra.Command, args []string) {
				cmd.List(setting)
			},
		},
		&cobra.Command{
			Use:   "list-cores",
			Short: "List all cores",
			Run: func(c *cobra.Command, args []string) {
				cmd.ListCores()
			},
		},
	)

	app.Execute()
}
