package cmd

import (
	"fmt"
	"os"

	"github.com/slmtnm/consul-editor/pkg/editor"
	"github.com/slmtnm/consul-editor/pkg/kv"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func handleError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "consul-editor [path-to-kv-folder]",
	Short: `Edit your consul KV via local editor`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		oldMap, err := kv.GetKeys(args[0])
		handleError(err)

		data, err := yaml.Marshal(oldMap)
		handleError(err)

		newData, err := editor.Edit(data)
		handleError(err)

		if newData == nil {
			fmt.Println("No changes, aboring...")
			return
		}

		newMap := make(map[string]interface{})
		err = yaml.Unmarshal(newData, newMap)
		handleError(err)

		diff := kv.CalculateDiff(oldMap, newMap)
		fmt.Println("Your changes:")
		diff.Print()

		fmt.Print("Confirm changes? [y/n] ")
		var answer string
		for answer != "y" && answer != "n" {
			fmt.Scan(&answer)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
