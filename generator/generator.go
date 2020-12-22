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

	newFolderPath := "generatedfiles"
	_ = os.Mkdir(newFolderPath, 0777)
	out, err := os.Create(path.Join(newFolderPath, "prerunyaml.go"))
	if err != nil {
		fmt.Println(err)
	}

	out.Write([]byte("package yamlsfile \n"))
	out.Write([]byte("var PreInstallYaml = `"))
	out.Write(fs)
	out.Write([]byte("`\n"))

	// for _, f := range fs {
	// 	fmt.Println("file name: %v", f.Name())
	// 	if !strings.HasSuffix(f.Name(), ".yaml") {
	// 		continue
	// 	}
	// 	out.Write([]byte(fmt.S"var \n"))
	// 	f, _ := os.Open(f.Name())
	// 	io.Copy(out, f)
	// 	out.Write([]byte("`\n"))
	// }
}
