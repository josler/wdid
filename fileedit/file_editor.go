package fileedit

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const defaultEditor = "vim"

func NewWithNoContent(filePath string) (string, error) {
	return EditWithExistingContent(filePath, strings.NewReader(""))
}

func EditWithExistingContent(filePath string, content io.Reader) (string, error) {
	err := writeTmpFile(filePath, content)
	if err != nil {
		return "", err
	}
	defer os.Remove(filePath)

	cmd := editorCmd(filePath)
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadFile(filePath)
	return string(data), err
}

func writeTmpFile(fpath string, content io.Reader) error {
	f, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, content)
	return err
}

func editorCmd(filePath string) *exec.Cmd {
	editorPath := os.Getenv("EDITOR")
	if editorPath == "" {
		editorPath = defaultEditor
	}
	editor := exec.Command(editorPath, filePath)

	editor.Stdin = os.Stdin
	editor.Stdout = os.Stdout
	editor.Stderr = os.Stderr

	return editor
}
