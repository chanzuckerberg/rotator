package cmd

import (
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "rotator",
	SilenceUsage: true,
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
