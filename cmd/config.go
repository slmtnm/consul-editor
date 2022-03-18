package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"text/tabwriter"

	"github.com/slmtnm/consul-editor/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	configCmd.AddCommand(getContextsCmd)
	configCmd.AddCommand(setContextCmd)
	configCmd.AddCommand(viewCmd)
	RootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure consul-editor",
}

var getContextsCmd = &cobra.Command{
	Use:   "get-contexts",
	Short: "List available contexts",
	Run: func(cmd *cobra.Command, args []string) {
		c := config.New()
		w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)
		fmt.Fprintln(w, "CURRENT\tNAME\tADDRESS\tTOKEN")
		for _, context := range c.Contexts {
			current := context.Name == c.CurrentContextName

			if current {
				fmt.Fprintln(w, fmt.Sprintf("*\t%s\t%s\t%s", context.Name, context.Address, context.Token))
			} else {
				fmt.Fprintln(w, fmt.Sprintf("\t%s\t%s\t%s", context.Name, context.Address, context.Token))
			}
		}
		w.Flush()
	},
}

var setContextCmd = &cobra.Command{
	Use:   "set-context",
	Short: "Set current context",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		newContext := args[0]

		c := config.New()
		found := false
		for _, context := range c.Contexts {
			if context.Name == newContext {
				found = true
				break
			}
		}

		if !found {
			fmt.Fprintf(os.Stderr, "context %s is undefined\n", newContext)
		}

		if config.ConfigFilename == "" {
			return
		}

		c.CurrentContextName = newContext
		bytes, err := yaml.Marshal(c)
		if err != nil {
			panic("could not marshal updated config")
		}

		if err := os.WriteFile(config.ConfigFilename, bytes, fs.FileMode(os.O_WRONLY)); err != nil {
			fmt.Fprintf(os.Stderr, "could not write config: %s\n", err.Error())
		}
	},
}

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Show configuration",
	Run: func(cmd *cobra.Command, args []string) {
		c := config.New()

		bytes, err := yaml.Marshal(c)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}

		fmt.Println(string(bytes))
	},
}
