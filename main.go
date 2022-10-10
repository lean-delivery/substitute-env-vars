package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

func getOsEnv(envVarName string, mode string) string {
	value := os.Getenv(envVarName)
	if value == "" && mode == "required" {
		log.Fatalf("ERROR: %s is not set", envVarName)
	}
	if value == "" && mode == "optional" {
		log.Printf("WARNING: %s is not set", envVarName)
	}
	return value
}

func getEnvVarsToReplace(variable string) []string {
	return strings.Split(os.Getenv(variable), ",")
}

func getMapEnvVarsToReplace(keysToReplace []string) map[string]string {
	var res = make(map[string]string)

	for _, key := range keysToReplace {
		res[key] = getOsEnv(key, "optional")
	}

	return res
}

func printBanner(mapToReplace map[string]string, mode string) {
	b := new(bytes.Buffer)
	for key, value := range mapToReplace {
		fmt.Fprintf(b, "%s = \"%s\"\n", key, value)
	}

	var announcementTemplate = `

## Target environment variables
_{TARGET_VAR_NAMES_PLACEHOLDER}_

## Sets of values
_{SETS_OF_VALUES}_

## Mode
_{MODE}_

`

	announcement := strings.ReplaceAll(announcementTemplate, "_{TARGET_VAR_NAMES_PLACEHOLDER}_", b.String())
	announcement = strings.ReplaceAll(announcement, "_{SETS_OF_VALUES}_", "N/A")
	announcement = strings.ReplaceAll(announcement, "_{MODE}_", mode)

	log.Print(announcement)

}

func processDir(_path string, mapToReplace map[string]string) {
	err := filepath.Walk(_path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				processFile(path, mapToReplace, info.Mode())
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
}

func processFile(_path string, mapToReplace map[string]string, _fileMode os.FileMode) {
	var fileContent []byte
	var err error

	if fileContent, err = os.ReadFile(_path); err != nil {
		log.Fatalf("ERROR: Failed reading file %s:\n%v", _path, err)
	}

	var fileContentStr = string(fileContent)

	for key, value := range mapToReplace {
		fileContentStr = strings.ReplaceAll(fileContentStr, fmt.Sprintf("_{%s}_", key), value)
	}

	if err = os.WriteFile(_path, []byte(fileContentStr), 0644); err != nil {
		log.Fatalf("ERROR: Failed writing file %s:\n%v", _path, err)
	}
	log.Printf("INFO: File %s processed", _path)
}

func readYAML(path string, key string) map[string]string {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("ERROR: Failed reading file %s:\n%v", path, err)
	}

	result := make(map[interface{}]interface{})
	err = yaml.Unmarshal(yamlFile, &result)
	if err != nil {
		log.Fatalf("ERROR: Failed unmarshalling file %s:\n%v", path, err)
	}
	var res = make(map[string]string)

	for k, v := range result[key].(map[string]interface{}) {
		res[k] = v.(string)
	}

	return res
}

func readJSON(path string, key string) map[string]string {
	jsonFile, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("ERROR: Failed reading file %s:\n%v", path, err)
	}

	result := make(map[string]interface{})
	err = json.Unmarshal(jsonFile, &result)
	if err != nil {
		log.Fatalf("ERROR: Failed unmarshalling file %s:\n%v", path, err)
	}
	var res = make(map[string]string)

	for k, v := range result[key].(map[string]interface{}) {
		res[k] = v.(string)
	}

	return res
}

func pathExists(path string) (os.FileInfo, bool, error) {
	if stat, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, false, err
			// file does not exist
		} else {
			return stat, true, err
		}
	} else {
		return stat, true, nil
	}
}

// detect operation mode
func _detectMode() string {

	// yaml mode
	if os.Getenv("SEV_YAML_PATH") != "" {
		return "yaml"
	}

	// env variables mode
	if os.Getenv("VAR_NAMES_STORAGE") != "" {
		return "env"
	}

	// json mode
	if os.Getenv("SEV_JSON_PATH") != "" {
		return "json"
	}
	return ""
}

func main() {

	var mapToReplace map[string]string
	var mode string

	switch _detectMode() {
	case "yaml":
		mode = "Pulling sets of values for variables from YAML file"
		mapToReplace = readYAML(getOsEnv("SEV_YAML_PATH", "required"), getOsEnv("SEV_YAML_KEY", "required"))
	case "json":
		mode = "Pulling sets of values for variables from JSON file"
		mapToReplace = readJSON(getOsEnv("SEV_JSON_PATH", "required"), getOsEnv("SEV_JSON_KEY", "required"))
	case "env":
		mode = "Pulling values for variables from environment"
		keysToReplace := getEnvVarsToReplace("VAR_NAMES_STORAGE")
		mapToReplace = getMapEnvVarsToReplace(keysToReplace)
	default:
		log.Println(_detectMode())
		log.Fatalf("ERROR: Unknown mode")
	}

	if len(os.Args) == 1 {
		log.Fatalf("ERROR: destination path missing")
	}

	var _path = os.Args[1]

	printBanner(mapToReplace, mode)

	stat, pathExists, _ := pathExists(_path)

	if pathExists {
		switch {
		case stat.IsDir():
			processDir(_path, mapToReplace)
		case !stat.IsDir():
			processFile(_path, mapToReplace, stat.Mode())
		}
	}
}
