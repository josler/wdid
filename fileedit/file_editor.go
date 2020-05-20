package fileedit

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/josler/wdid/config"
)

const defaultEditor = "vim"

func NewWithNoContent() (io.Reader, error) {
	filePath := config.ConfigDir() + "/WDID_TEMP"
	return editWithContent(filePath, strings.NewReader(""))
}

func EditExisting(data string) (io.Reader, error) {
	filePath := config.ConfigDir() + "/WDID_TEMP"
	return editWithContent(filePath, strings.NewReader(data))
}

func editWithContent(filePath string, content io.Reader) (io.Reader, error) {
	err := writeTmpFile(filePath, content)
	if err != nil {
		return content, err
	}
	defer os.Remove(filePath)

	cmd := editorCmd(filePath)
	err = cmd.Run()
	if err != nil {
		return content, err
	}
	return os.Open(filePath)
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
	editorPath := editorPath()
	splitPath := strings.Split(editorPath, " ")
	splitPath = append(splitPath, filePath)
	// we always have at least 2 elements. the editor path and the filepath
	editor := exec.Command(splitPath[0], splitPath[1:]...)

	editor.Stdin = os.Stdin
	editor.Stdout = os.Stdout
	editor.Stderr = os.Stderr

	return editor
}

func editorPath() string {
	conf, _ := config.Load()
	if conf.Editor != "" {
		return conf.Editor
	}
	editorPath := os.Getenv("EDITOR")
	if editorPath != "" {
		return editorPath
	}
	return defaultEditor
}
