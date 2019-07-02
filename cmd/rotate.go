package cmd

import (
	"github.com/chanzuckerberg/rotator/pkg/config"
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
			return errors.Wrap(err, "couldn't parse config flag")
		}

		config, err := config.FromFile(file)
		if err != nil {
			return err
		}
		return RotateSecrets(config)
	},
}

// RotateSecrets takes a config, reads the secret from the source,
// and writes it to each sink.
func RotateSecrets(config *config.Config) error {
	for _, secret := range config.Secrets {
		// Rotate credential at source
		src := secret.Source
		newCreds, err := src.Read()
		if err != nil {
			return errors.Wrapf(err, "%s: unable to rotate secret at %s", secret.Name, src.Kind())
		}
		if newCreds == nil {
			return nil
		}

		// Write new credential to each sink
		for _, sink := range secret.Sinks {
			err = sink.Write(newCreds)
			if err != nil {
				return errors.Wrapf(err, "%s: unable to write secret to %s", secret.Name, sink.Kind())
			}
		}
	}
	return nil
}
