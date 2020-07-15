package cmd

import (
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/honeycombio/beeline-go"
	"github.com/kelseyhightower/envconfig"
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

	err = configureHoneycombTelemetry()
	if err != nil {
		logrus.Warn(errors.Wrap(err, "Unable to set up Honeycomb Telemetry"))
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

type HoneycombEnvironment struct {
	SECRET_KEY   string
	DATASET_NAME string `default:"rotator"`
	SERVICE_NAME string `default:"rotator"`
}

func loadHoneycombEnv() (*HoneycombEnvironment, error) {
	env := &HoneycombEnvironment{}
	err := envconfig.Process("HONEYCOMB", env)
	if err != nil {
		return env, errors.Wrap(err, "Unable to load all the honeycomb environment variables")
	}
	return env, nil
}

func configureHoneycombTelemetry() error {
	honeycombEnv, err := loadHoneycombEnv()
	if err != nil {
		return err
	}
	// if env var not set, ignore
	if honeycombEnv.SECRET_KEY == "" {
		logrus.Debug("Honeycomb Secret Key not set. Skipping Honeycomb Configuration")
		return nil
	}
	beeline.Init(beeline.Config{
		WriteKey:    honeycombEnv.SECRET_KEY,
		Dataset:     honeycombEnv.DATASET_NAME,
		ServiceName: honeycombEnv.SERVICE_NAME,
	})

	return nil
}
