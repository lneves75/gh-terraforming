package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var ctx = context.Background()
var log = logrus.New()
var orgName, apiToken, logLevel, outDirectory string
var verbose bool
var api *github.Client

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gh-terraforming",
	Short: "Bootstrapping Terraform from existing Github organization",
	Long: `gh-terraforming is an application that allows teams to start
using Terraform by describing and importing existing resources in Github.`,
	PersistentPreRun:  persistentPreRun,
	PersistentPostRun: persistentPostRun,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		return
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Personal access token
	rootCmd.PersistentFlags().StringVarP(&apiToken, "token", "t", "", "Github Token")

	// Organization
	rootCmd.PersistentFlags().StringVarP(&orgName, "organization", "o", "", "Scope operations to this organization")

	// Output directory
	rootCmd.PersistentFlags().StringVarP(&outDirectory, "out-dir", "d", "", "Write resource files to this directory (default to PWD)")

	// Debug logging mode
	rootCmd.PersistentFlags().StringVarP(&logLevel, "loglevel", "l", "", "Specify logging level: (trace, debug, info, warn, error, fatal, panic)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Specify verbose output (same as setting log level to debug)")

	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindEnv("token", "TOKEN")

	viper.BindPFlag("organization", rootCmd.PersistentFlags().Lookup("organization"))
	viper.BindEnv("organization", "ORGANIZATION")
}

// initConfig reads in ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvPrefix("github")

	var cfgLogLevel = logrus.InfoLevel

	// A user may also pass the verbose flag in order to support this convention
	if verbose {
		cfgLogLevel = logrus.DebugLevel
	}

	switch strings.ToLower(logLevel) {
	case "trace":
		cfgLogLevel = logrus.TraceLevel
	case "debug":
		cfgLogLevel = logrus.DebugLevel
	case "info":
		break
	case "warn":
		cfgLogLevel = logrus.WarnLevel
	case "error":
		cfgLogLevel = logrus.ErrorLevel
	case "fatal":
		cfgLogLevel = logrus.FatalLevel
	case "panic":
		cfgLogLevel = logrus.PanicLevel
	}

	log.SetLevel(cfgLogLevel)
}

// This function runs before every root command
func persistentPreRun(cmd *cobra.Command, args []string) {

	if cmd.Name() != "version" {

		if apiToken = viper.GetString("token"); apiToken == "" {
			log.Error("-t/--token option or GITHUB_TOKEN env var must be set")
			return
		}

		if orgName = viper.GetString("organization"); orgName == "" {
			log.Error("-o/--organization option or GITHUB_ORGANIZATION env var must be set")
			return
		}

		log.WithFields(logrus.Fields{
			"Token":        fmt.Sprintf("*************%s", apiToken[:4]),
			"Organization": orgName,
		}).Debug("Initializing go-github")

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: apiToken},
		)
		tc := oauth2.NewClient(ctx, ts)

		api = github.NewClient(tc)

		if outDirectory == "" {
			outDirectory, _ = os.Getwd()
		}
	}
}

// This function runs following every root command
func persistentPostRun(cmd *cobra.Command, args []string) {
	return
}
