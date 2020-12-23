package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// Reads all .txt files in the current folder
// and encodes them as strings literals in textfiles.go
func main() {
	fs, _ := ioutil.ReadFile("generator/pre_install.yaml")

	newFolderPath := "autogenerate"
	_ = os.Mkdir(newFolderPath, 0777)
	out, err := os.Create(path.Join(newFolderPath, "autogenerate.go"))
	if err != nil {
		fmt.Println(err)
	}

	out.Write([]byte("// THIS FILE IS AUTO GENERATED ON MAKE COMMAND - DO NOT EDIT\n \n"))
	out.Write([]byte("package autogenerate \n"))
	out.Write([]byte("var PreInstallYaml = `"))
	out.Write(fs)
	out.Write([]byte("`\n"))
}
