package src

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var htmlFolder string
var goFile string
var mapName string
var packageName string

var blankMap = make(map[string]interface{})

// HTMLInit : Dateipfade festlegen
// mapName = Name der zu erstellenden map
// htmlFolder: Ordner der HTML Dateien
// packageName: Name des package
func HTMLInit(name, pkg, folder, file string) {

	htmlFolder = folder
	goFile = file
	mapName = name
	packageName = pkg

}

// BuildGoFile : Erstellt das GO Dokument
func BuildGoFile() error {

	var err = checkHTMLFile(htmlFolder)

	if err != nil {
		return err
	}

	var content string
	content += `package ` + packageName + "\n\n"
	content += `var ` + mapName + ` = make(map[string]interface{})` + "\n\n"
	content += "func loadHTMLMap() {" + "\n\n"

	mapContent, err := createMapFromFiles(htmlFolder)
	if err != nil {
		return err
	}
	content += mapContent + "\n"

	content += "}" + "\n\n"
	return writeStringToFile(goFile, content)
}

// GetHTMLString : base64 -> string
func GetHTMLString(base string) string {
	content, _ := base64.StdEncoding.DecodeString(base)
	return string(content)
}

func createMapFromFiles(folder string) (string, error) {

	var path = getLocalPath(folder)

	err := filepath.Walk(path, readFilesToMap)
	if err != nil {
		return "", err
	}

	var content string

	for key := range blankMap {
		var newKey = key
		content += `  ` + mapName + `["` + newKey + `"` + `] = "` + blankMap[key].(string) + `"` + "\n"
	}

	return content, nil
}

func readFilesToMap(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("missing file information for %s", path)
	}

	if info.IsDir() == false {
		base64Str, err := fileToBase64(getLocalPath(path))
		if err != nil {
			return err
		}
		blankMap[path] = base64Str
	}

	return nil
}

func fileToBase64(file string) (string, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(content), nil
}

func getLocalPath(filename string) string {

	path, file := filepath.Split(filename)
	var newPath = filepath.Dir(path)

	var newFileName = newPath + "/" + file

	return newFileName
}

func writeStringToFile(filename, content string) error {

	err := os.WriteFile(getPlatformFile(filename), []byte(content), 0644)
	if err != nil {
		checkErr(err)
		return err
	}

	return nil
}

func checkHTMLFile(filename string) error {

	if _, err := os.Stat(getLocalPath(filename)); os.IsNotExist(err) {
		fmt.Println(filename)
		checkErr(err)
		return err
	}

	return nil
}

func checkErr(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Println("ERROR: [", err, "] in ", file, line)
	}
}
