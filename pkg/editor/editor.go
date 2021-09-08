package editor

import (
	"crypto/sha256"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

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

	_, err = file.Write(data)
	if err != nil {
		return nil, err
	}

	oldHash := sha256.Sum256(data)

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

	if oldHash == sha256.Sum256(newData) {
		return nil, errors.New("no changes")
	}

	return newData, nil
}
