package cmd

import (
	"os"
	"strconv"

	"github.com/airbrake/gobrake"
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
	airbrake, err := setUpAirbrake()
	if err != nil {
		logrus.Warn(errors.Wrap(err, "unable to set up Airbrake notifier"))
	} else {
		defer func(airbrake *gobrake.Notifier) {
			if err := airbrake.Close(); err != nil {
				logrus.Error(errors.Wrap(err, "unable to close Airbrake notifier"))
			}
		}(airbrake)
		defer airbrake.NotifyOnPanic()
	}

	if err := rootCmd.Execute(); err != nil {
		if airbrake != nil {
			airbrake.Notify(err, nil)
			airbrake.Flush()
		}
		logrus.Fatal(err)
	}
}

// reference: https://github.com/chanzuckerberg/mergebot/blob/fa0c67c2363f5f9e5a432a116ee82a294a9e5eea/cmd/run.go
func setUpAirbrake() (*gobrake.Notifier, error) {
	env := os.Getenv("ENV")
	if env == "" {
		return nil, errors.New("please set the ENV variable to the name of the current execution environment (dev, stage, prod, etc)")
	}
	airEnv := os.Getenv("AIRBRAKE_PROJECT_ID")
	airbrakeProjectID, err := strconv.ParseInt(airEnv, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse AIRBRAKE_PROJECT_ID variable to int64")
	}

	var airbrake *gobrake.Notifier
	if airbrakeProjectID != 0 {
		airbrakeProjectKey := os.Getenv("AIRBRAKE_PROJECT_KEY")
		if airbrakeProjectKey == "" {
			return nil, errors.New("If you set AIRBRAKE_PROJECT_ID, you need to also set AIRBRAKE_PROJECT_KEY")
		}
		airbrake = gobrake.NewNotifierWithOptions(&gobrake.NotifierOptions{
			ProjectId:   airbrakeProjectID,
			ProjectKey:  airbrakeProjectKey,
			Environment: env,
		})
		airbrake.AddFilter(func(notice *gobrake.Notice) *gobrake.Notice {
			notice.Context["environment"] = env
			return notice
		})
	}
	return airbrake, nil
}
