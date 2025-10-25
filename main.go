package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const appName = "gh-auth-switch"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "hook" {
		fmt.Print(hookScript)
		return
	}
	if len(os.Args) == 1 {
		doMain()
		return
	}
	log.Fatalf("arguments must be zero or one")
}

func doMain() {
	config, err := readConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to read config. caused by %w", err))
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get working directory. caused by %w", err))
	}
	for _, configPair := range *config {
		pathPrefix := configPair.PathPrefix
		user := configPair.User
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(fmt.Errorf("failed to get user home directory. caused by %w", err))
		}
		pathPrefix = strings.Replace(pathPrefix, "~", homeDir, 1)
		matched := strings.HasPrefix(workingDirectory, pathPrefix)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to match pattern %s. caused by %w", pathPrefix, err))
		}
		if matched {
			cmd := exec.Command("gh", "auth", "switch", "-u", user)
			cmd.Run()
			fmt.Println(user)
			return
		}
	}
}

type ConfigPair struct {
	PathPrefix string
	User       string
}

type Config = []ConfigPair

func readConfig() (*Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config dir. caused by %w", err)
	}
	configPath := filepath.Join(configDir, appName, "config.yaml")
	configFile, err := openFileAndCreateIfNecessaryRecursive(configPath, os.O_RDONLY, 0777)
	if err != nil {
		return nil, fmt.Errorf("failed to open a config: %s. caused by %w", configPath, err)
	}
	defer configFile.Close()
	configBytes, err := io.ReadAll(configFile)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read a config: %s. caused by %w", configPath, err)
	}
	config, err := UnmarshalConfig(configBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal a config: %s. caused by %w", configPath, err)
	}
	return config, nil
}

func UnmarshalConfig(bytes []byte) (*Config, error) {
	if !utf8.Valid(bytes) {
		return nil, fmt.Errorf("invalid UTF-8")
	}
	str := string(bytes)
	var config Config
	lines := strings.Split(str, "\n")
	config = make(Config, 0)
	for _, line := range lines {
		if line == "" {
			continue
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		words := strings.Split(line, ":")
		if len(words) != 2 {
			return nil, fmt.Errorf("line must contain exactly one colon: %s", line)
		}
		pathPrefix := strings.TrimSpace(words[0])
		if pathPrefix == "" {
			return nil, fmt.Errorf("pathPrefix must not be empty: %s", line)
		}
		if pathPrefix[0] == '"' && pathPrefix[len(pathPrefix)-1] == '"' {
			pathPrefix = pathPrefix[1 : len(pathPrefix)-1]
		}
		user := strings.TrimSpace(words[1])
		if user == "" {
			return nil, fmt.Errorf("user must not be empty: %s", line)
		}
		config = append(config, ConfigPair{PathPrefix: pathPrefix, User: user})
	}
	return &config, nil
}

func openFileAndCreateIfNecessaryRecursive(path string, flag int, mode os.FileMode) (*os.File, error) {
	file, err := os.OpenFile(path, flag, mode)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to open the file: %s. caused by %w", path, err)
		} else {
			file, err = os.Create(path)
			if err != nil {
				if !os.IsNotExist(err) {
					return nil, fmt.Errorf("failed to create the file: %s. caused by %w", path, err)
				}
				dir := filepath.Dir(path)
				err = os.MkdirAll(dir, 0755)
				if err != nil {
					return nil, fmt.Errorf("failed to create the directory: %s. caused by %w", dir, err)
				}
				file, err = os.Create(path)
				if err != nil {
					return nil, fmt.Errorf("failed to create the file after creating the directory: %s. caused by %w", path, err)
				}
			}
		}
	}
	return file, nil
}

//go:embed hook.bash
var hookScript string
