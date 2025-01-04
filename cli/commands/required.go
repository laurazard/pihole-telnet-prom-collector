package commands

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// ExactArgs returns an error if there is not the exact number of args
func ExactArgs(number int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == number {
			return nil
		}
		return errors.Errorf(
			"%[1]s: '%[2]s' requires %[3]d %[4]s\n\nUsage:  %[5]s\n\nRun '%[2]s --help' for more information",
			binName(cmd),
			cmd.CommandPath(),
			number,
			pluralize("argument", number),
			cmd.UseLine(),
		)
	}
}

// binName returns the name of the binary / root command (usually 'docker').
func binName(cmd *cobra.Command) string {
	return cmd.Root().Name()
}

//nolint:unparam
func pluralize(word string, number int) string {
	if number == 1 {
		return word
	}
	return word + "s"
}
