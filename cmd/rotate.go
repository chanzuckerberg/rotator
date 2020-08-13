package cmd

import (
	"context"
	"fmt"

	"github.com/chanzuckerberg/rotator/pkg/config"
	"github.com/fatih/color"
	"github.com/hashicorp/go-multierror"
	"github.com/honeycombio/beeline-go"
	"github.com/pkg/errors"
	"github.com/segmentio/go-prompt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rotateCmd.Flags().StringP("file", "f", "", "Config file to read from")
	rotateCmd.Flags().BoolP("yes", "y", false, "Assume \"yes\" to all prompts and run non-interactively.")
	rootCmd.AddCommand(rotateCmd)
}

var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate secrets",
	Long: `rotate parses a config file, rotates the secret at
			the source, and writes the new secret to each sink`,
	SilenceErrors: true, // If we don't silence here, cobra will print them. But we want to do that in cmd/root.go
	RunE: func(cmd *cobra.Command, args []string) error {
		// parse config and print plan
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			return errors.Wrap(err, "unable to parse config flag")
		}
		config, err := config.FromFile(file)
		if err != nil {
			return errors.Wrap(err, "unable to read config from file")
		}
		printPlan(config)

		// prompt user to continue if necessary
		skipPrompt, err := cmd.Flags().GetBool("yes")
		if err != nil {
			return errors.Wrap(err, "unable to parse yes flag")
		}
		if !(skipPrompt || getPrompt()) {
			return nil
		}

		// rotate secrets
		logrus.Println("Performing the actions described above.")
		return RotateSecrets(config)
	},
}

func printPlan(config *config.Config) {
	for _, secret := range config.Secrets {
		logrus.Println("Rotating secret", secret.Name)

		src := secret.Source
		logrus.Println("*", "reading new credentials from", src.Kind(), "source")

		for _, sink := range secret.Sinks {
			logrus.Println("*", "writing new credentials to", sink.Kind(), "sink")
		}
		logrus.Println()
	}
}

func getPrompt() bool {
	// print prompt
	b := color.New(color.Bold).SprintFunc()
	logrus.Println(b("Do you want to perform these actions?"))
	logrus.Println("  rotator will perform the actions described above.")
	logrus.Println("  Yes', 'yes', 'y', 'Y' will be accepted to approve.")
	logrus.Println("  No', 'no', 'n', 'N' will cancel the rotation.")
	logrus.Println()

	// get user input
	yes := prompt.Confirm(b("  Enter a value"))
	logrus.Println()
	if !yes {
		redB := color.New(color.FgRed, color.Bold).SprintFunc()
		logrus.Errorln()
		logrus.Errorln(redB("Error: "), "Rotation cancelled.")
		logrus.Errorln()
	}
	return yes
}

// RotateSecrets takes a config, reads the secret from the source,
// and writes it to each sink.
func RotateSecrets(config *config.Config) error {
	var errs *multierror.Error
	ctx := context.Background()
	ctx, span := beeline.StartSpan(ctx, "rotateSecrets")
	defer span.Send()

	for _, secret := range config.Secrets {
		// Rotate credential at source
		src := secret.Source
		newCreds, err := src.Read()
		if err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "%s: unable to rotate secret at %s", secret.Name, src.Kind()))
			continue
		}
		if newCreds == nil { // not time to rotate yet
			continue
		}

		// Write new credentials to each sink
		for _, sink := range secret.Sinks {
			keyToName := sink.GetKeyToName()
			if keyToName == nil {
				errs = multierror.Append(errs, errors.New(fmt.Sprintf("%s: missing value in KeyToName field for %s sink", secret.Name, sink.Kind())))
				continue
			}
			for k, v := range newCreds {
				name, ok := keyToName[k]
				if !ok {
					errs = multierror.Append(errs, errors.New(fmt.Sprintf("%s: no name specified for credential with key %s for %s sink", secret.Name, k, sink.Kind())))
					continue
				}
				err = sink.Write(ctx, name, v)
				if err != nil {
					errs = multierror.Append(errors.Wrapf(err, "%s: unable to write secret to %s sink", secret.Name, sink.Kind()))
					continue
				}
			}
		}
	}
	return errs.ErrorOrNil()
}
