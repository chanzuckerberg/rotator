package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/chanzuckerberg/rotator/pkg/config"
	"github.com/fatih/color"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
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
		skip, err := cmd.Flags().GetBool("yes")
		if err != nil {
			return errors.Wrap(err, "unable to parse yes flag")
		}
		if !skip {
			yes := getPrompt()
			if !yes {
				return nil
			}
		}

		// rotate secrets
		fmt.Println("Performing the actions described above.")
		return RotateSecrets(config)
	},
}

func printPlan(config *config.Config) {
	for _, secret := range config.Secrets {
		fmt.Println("Rotating secret", secret.Name)

		src := secret.Source
		fmt.Println("*", "reading new credentials from", src.Kind(), "source")

		for _, sink := range secret.Sinks {
			fmt.Println("*", "writing new credentials to", sink.Kind(), "sink")
		}
		fmt.Println()
	}
}

func getPrompt() bool {
	// print prompt
	b := color.New(color.Bold)
	b.Print("Do you want to perform these actions?\n")
	fmt.Println("  rotator will perform the actions described above.")
	fmt.Println("  Only 'yes' will be accepted to approve.")
	fmt.Println()

	// get user input
	b.Print("  Enter a value: ")
	reader := bufio.NewReader(os.Stdin)
	ans, _ := reader.ReadString('\n')
	fmt.Println()
	if ans != "yes\n" {
		fmt.Println()
		color.New(color.FgRed, color.Bold).Print("Error: ")
		b.Print("Rotation cancelled.\n")
		fmt.Println()
	}
	return ans == "yes\n"
}

// RotateSecrets takes a config, reads the secret from the source,
// and writes it to each sink.
func RotateSecrets(config *config.Config) error {
	var errs *multierror.Error
	ctx := context.Background()
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
