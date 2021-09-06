package cmd

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/hashicorp/consul/api"
	"github.com/slmtnm/consul-editor/pkg"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var rootCmd = &cobra.Command{
	Use:   "consul-editor [path-to-kv-folder]",
	Short: `Edit your consul KV via local editor`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			log.Fatal(err)
		}
		kv := client.KV()
		path := args[0]

		keys, _, err := kv.List(path, nil)
		if err != nil {
			log.Fatal(err)
		}
		if len(keys) == 0 {
			fmt.Printf("Key '%s' not found\n", path)
			os.Exit(1)
		}

		data := pkg.KeysToMap(keys)
		d, err := yaml.Marshal(data)

		if err != nil {
			panic(err)
		}

		file, err := ioutil.TempFile("", "consul-editor*.yaml")
		if err != nil {
			panic(err)
		}

		_, err = file.Write(d)
		if err != nil {
			panic(err)
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano"
		}

		file.Seek(0, io.SeekStart)
		d, err = ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		oldHash := sha256.Sum256(d)
		file.Close()

		editorCmd := exec.Command(editor, file.Name())
		editorCmd.Stderr = os.Stderr
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		if err := editorCmd.Run(); err != nil {
			panic(err)
		}

		file, err = os.Open(file.Name())
		if err != nil {
			panic(err)
		}
		defer file.Close()
		defer os.Remove(file.Name())

		d, err = ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		newHash := sha256.Sum256(d)

		if newHash == oldHash {
			fmt.Println("No changes, aborting...")
			return
		}

		yamlData := make(map[string]interface{})
		err = yaml.Unmarshal(d, yamlData)
		if err != nil {
			panic(err)
		}

		_, err = pkg.MapToKeys(yamlData)
		if err != nil {
			panic(err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
