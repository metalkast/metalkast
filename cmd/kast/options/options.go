package options

import "github.com/spf13/cobra"

var Verbosity int

const maxVerbosity = 5

func Add(cmd *cobra.Command) {
	cmd.PersistentFlags().CountVar(&Verbosity, "v", "number for the log level verbosity")
	if Verbosity > maxVerbosity {
		Verbosity = maxVerbosity
	}
}
