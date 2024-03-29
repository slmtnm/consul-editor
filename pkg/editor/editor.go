package editor

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

const header = `# This is Consul KV converted to YAML format.
# Feel free to edit it, then save and close the file to
# see the difference you impact and eventually apply those changes.

`

var editorExecutable string

func init() {
	editorExecutable = os.Getenv("EDITOR")
	if editorExecutable == "" {
		editorExecutable = "nano"
	}
}

func Edit(data []byte) ([]byte, error) {
	file, err := ioutil.TempFile("", "consul-editor*.yaml")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	defer os.Remove(file.Name())

	_, err = file.Write([]byte(header))
	if err != nil {
		return nil, err
	}

	_, err = file.Write(data)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(editorExecutable, file.Name())
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	file.Seek(0, io.SeekStart)
	newData, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return newData, nil
}
