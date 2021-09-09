package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
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
		globalPrefix := strings.Trim(args[0], "/")

		oldMap, err := kv.GetKeys(args[0])
		handleError(err)

		if globalPrefix != "" { // "cd" to appropriate folder
			if len(oldMap) == 0 {
				fmt.Fprintln(os.Stderr, "There are no keys in Consul KV, try run with prefix '/' to add some")
				return
			}
			for _, prefixKey := range strings.Split(globalPrefix, "/") {
				i, ok := oldMap[prefixKey]
				if !ok {
					fmt.Fprintf(os.Stderr, "No keys with prefix: %s\n", globalPrefix)
					return
				}
				switch v := i.(type) {
				case string:
					fmt.Fprintf(os.Stderr, "Could not open key %s, you can open only folders\n", color.YellowString(prefixKey))
					os.Exit(0)
				case map[string]interface{}:
					oldMap = v
				}
			}
		}

		data := bytes.Buffer{}
		encoder := yaml.NewEncoder(&data)
		encoder.SetIndent(2)
		err = encoder.Encode(oldMap)
		handleError(err)

		newData, err := editor.Edit(data.Bytes())
		handleError(err)

		newMap := make(map[string]interface{})
		err = yaml.Unmarshal(newData, newMap)
		handleError(err)

		diff := kv.CalculateDiff(oldMap, newMap)
		if len(diff.Added) == 0 && len(diff.Removed) == 0 {
			fmt.Println("No changes made, aboring...")
			return
		}

		fmt.Println("Your changes:")
		diff.Print(globalPrefix)

		var answer string
		for answer != "y" && answer != "n" {
			fmt.Print("Confirm changes? [y/n] ")
			fmt.Scan(&answer)
		}

		if answer == "y" {
			diff.Apply(globalPrefix)
			fmt.Println("Changes applied!")
		} else {
			fmt.Println("No changes made")
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
