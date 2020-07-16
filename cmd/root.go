package cmd

import (
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/honeycombio/beeline-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	flagVerbose = "verbose"
)

func init() {
	rootCmd.PersistentFlags().BoolP(flagVerbose, "v", false, "Use this to enable verbose mode")
}

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
			log.SetLevel(log.DebugLevel)
			log.SetReportCaller(true)
		}

		err = configureHoneycombTelemetry()
		if err != nil {
			return errors.Wrap(err, "Unable to set up Honeycomb Telemetry")
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		beeline.Flush(cmd.Context())
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	sentryEnabled, err := setUpSentry()
	if err != nil {
		log.Warn(errors.Wrap(err, "unable to set up Sentry notifier"))
	} else if sentryEnabled {
		defer sentry.Flush(time.Second * 5)
		defer sentry.Recover()
	}

	if err := rootCmd.Execute(); err != nil {
		if sentryEnabled {
			sentry.CaptureException(err)
			sentry.Flush(time.Second * 5)
		}
		log.Fatal(err)
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
		log.Debug("Honeycomb Secret Key not set. Skipping Honeycomb Configuration")
		return nil
	}
	beeline.Init(beeline.Config{
		WriteKey:    honeycombEnv.SECRET_KEY,
		Dataset:     honeycombEnv.DATASET_NAME,
		ServiceName: honeycombEnv.SERVICE_NAME,
	})
	return nil
}
