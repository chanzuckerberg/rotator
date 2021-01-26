package cmd

import (
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	flagVerbose = "verbose"
)

var rootCmd = &cobra.Command{
	Use:          "rotator",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// parse flags
		verbose, err := cmd.Flags().GetBool(flagVerbose)
		if err != nil {
			return errors.Wrap(err, "Missing verbose flag")
		}
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
			logrus.SetReportCaller(true) // add the calling method as a field
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolP(flagVerbose, "v", false, "Use this to enable verbose mode")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	sentryEnabled, err := setUpSentry()
	if err != nil {
		logrus.Warn(errors.Wrap(err, "unable to set up Sentry notifier"))
	} else if sentryEnabled {
		defer sentry.Flush(time.Second * 5)
		defer sentry.Recover()
	}

	if err := rootCmd.Execute(); err != nil {
		if sentryEnabled {
			sentry.CaptureException(err)
			sentry.Flush(time.Second * 5)
		}
		logrus.Fatal(err)
	}
}

func setUpSentry() (bool, error) {
	env := os.Getenv("ENV")
	if env == "" {
		return false, errors.New("please set the ENV variable to the name of the current execution environment (dev, stage, prod, etc)")
	}
	sentryDsn := os.Getenv("SENTRY_DSN")

	if sentryDsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:         sentryDsn,
			Environment: env,
		})
		if err != nil {
			return false, errors.Wrap(err, "sentry initialization failed")
		}
	}
	return true, nil
}
