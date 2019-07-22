package cmd

import (
	"context"
	"fmt"

	"github.com/chanzuckerberg/rotator/pkg/config"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	rotateCmd.Flags().StringP("file", "f", "", "Config file to read from")
	rootCmd.AddCommand(rotateCmd)
}

var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate secrets",
	Long: `rotate parses a config file, rotates the secret at 
			the source, and writes the new secret to each sink`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// handle flags
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			return errors.Wrap(err, "unable to parse config flag")
		}

		config, err := config.FromFile(file)
		if err != nil {
			return errors.Wrap(err, "unable to read config from file")
		}
		return RotateSecrets(config)
	},
}

// RotateSecrets takes a config, reads the secret from the source,
// and writes it to each sink.
func RotateSecrets(config *config.Config) error {
	ctx := context.Background()
	for _, secret := range config.Secrets {
		var errs *multierror.Error

		// Rotate credential at source
		src := secret.Source
		newCreds, err := src.Read()
		if err != nil {
			return errors.Wrapf(err, "%s: unable to rotate secret at %s", secret.Name, src.Kind())
		}
		if newCreds == nil {
			return nil
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
	return nil
}
