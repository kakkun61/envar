package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

const appName = "envar"

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("arguments must be one or more: %d", len(os.Args)-1)
	}
	switch os.Args[1] {
	case "hook":
		switch len(os.Args) {
		case 2:
			fmt.Print(hookScript)
		case 4:
			if os.Args[2] != "logout" {
				log.Fatalf("unknown hook type: %s", os.Args[2])
			}
			shellPid, err := strconv.ParseUint(os.Args[3], 10, 32)
			if err != nil {
				log.Fatalf("invalid shell PID: %s", os.Args[3])
			}
			cachePath := makeCachedScriptPath(uint(shellPid))
			err = os.Remove(cachePath)
			if err != nil && !os.IsNotExist(err) {
				log.Fatal(fmt.Errorf("failed to remove cache file: %s, because %w", cachePath, err))
			}
		default:
			log.Fatalf("invalid number of arguments for hook: %d", len(os.Args)-1)
		}
	default:
		if len(os.Args) != 2 {
			log.Fatalf("give the shell PID")
		}
		shellPid, err := strconv.ParseUint(os.Args[1], 10, 32)
		if err != nil {
			log.Fatalf("invalid shell PID: %s", os.Args[1])
		}
		doMain(uint(shellPid))
	}
}

func doMain(shellPid uint) {
	config, err := readConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to read config, because %w", err))
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get working directory, because %w", err))
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get user home directory, because %w", err))
	}
	script := make([]string, 0)
	for varName, pathValues := range *config {
		for _, pathValue := range pathValues {
			pathPrefix := pathValue.PathPrefix
			pathPrefix = strings.Replace(pathPrefix, "~", homeDir, 1)
			matched := strings.HasPrefix(workingDirectory, pathPrefix)
			if matched {
				if pathValue.Exec != nil {
					v, err := runExecCommand(*pathValue.Exec)
					if err != nil {
						log.Fatal(fmt.Errorf("failed to run exec for %s, because %w", varName, err))
					}
					script = append(script, fmt.Sprintf("export %s=%s", varName, v))
				} else if pathValue.Value == nil {
					script = append(script, fmt.Sprintf("unset %s", varName))
				} else {
					script = append(script, fmt.Sprintf("export %s=%s", varName, *pathValue.Value))
				}
				goto nextVar
			}
		}
		// No match found for this variable, unset it
		script = append(script, fmt.Sprintf("unset %s", varName))
	nextVar:
	}
	previousScript := readCachedScript(shellPid)
	for _, line := range script {
		if !slices.Contains(previousScript, line) {
			fmt.Println(line)
		}
	}
	writeCachedScript(shellPid, script)
}

type PathValue struct {
	PathPrefix string
	Value      *string // nil means unset
	Exec       *string // optional command to execute for value
}

type Config = map[string][]PathValue

func readConfig() (*Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config dir, because %w", err)
	}
	configPath := filepath.Join(configDir, appName, "config.yaml")
	configFile, err := openFileAndCreateIfNecessaryRecursive(configPath, os.O_RDONLY, 0777)
	if err != nil {
		return nil, fmt.Errorf("failed to open a config: %s, because %w", configPath, err)
	}
	defer configFile.Close()
	configBytes, err := io.ReadAll(configFile)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read a config: %s, because %w", configPath, err)
	}
	config, err := UnmarshalConfig(configBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal a config: %s, because %w", configPath, err)
	}
	return config, nil
}

func UnmarshalConfig(bytes []byte) (*Config, error) {
	if !utf8.Valid(bytes) {
		return nil, fmt.Errorf("config is invalid UTF-8")
	}
	str := string(bytes)
	config := make(Config)
	lines := strings.Split(str, "\n")
	var currentVarName string
	var currentPathIndent int = -1
	var currentPathValue *PathValue
	for i := range lines {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}
		indent := countIndent(line)
		if indent == 0 {
			// 変数名の行が期待される
			if !strings.HasSuffix(trimmedLine, ":") {
				return nil, fmt.Errorf("variable name line must end with colon: %s", trimmedLine)
			}
			name := strings.TrimSpace(trimmedLine[:len(trimmedLine)-1])
			if name == "" {
				return nil, fmt.Errorf("variable name must not be empty: %s", trimmedLine)
			}
			name = trimQuotes(name)
			currentVarName = name
			if _, exists := config[currentVarName]; !exists {
				config[currentVarName] = make([]PathValue, 0)
			}
			currentPathIndent = -1
			currentPathValue = nil
			continue
		}
		// indent > 0
		if currentVarName == "" {
			return nil, fmt.Errorf("path-and-value line must be under a variable name: %s", trimmedLine)
		}
		if currentPathValue != nil && indent > currentPathIndent {
			// exec の行が期待される
			parts := strings.SplitN(trimmedLine, ":", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("nested key must be in 'key: value' format: %s", trimmedLine)
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = trimQuotes(val)
			switch key {
			case "exec":
				if val == "" {
					return nil, fmt.Errorf("exec value must not be empty: %s", trimmedLine)
				}
				currentPathValue.Value = nil
				currentPathValue.Exec = &val
			default:
				return nil, fmt.Errorf("unknown nested key: %s", key)
			}
			continue
		}
		// パスの行が期待される
		parts := strings.SplitN(trimmedLine, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("path and value must be separated by a colon: %s", trimmedLine)
		}
		pathPrefix := strings.TrimSpace(parts[0])
		if pathPrefix == "" {
			return nil, fmt.Errorf("pathPrefix must not be empty: %s", trimmedLine)
		}
		pathPrefix = trimQuotes(pathPrefix)
		valueStr := strings.TrimSpace(parts[1])
		var value *string
		if valueStr == "" || valueStr == "null" {
			value = nil
		} else {
			valueStr = trimQuotes(valueStr)
			value = &valueStr
		}
		pathValue := PathValue{PathPrefix: pathPrefix, Value: value}
		config[currentVarName] = append(config[currentVarName], pathValue)
		currentPathIndent = indent
		currentPathValue = &config[currentVarName][len(config[currentVarName])-1]
	}
	return &config, nil
}

func countIndent(s string) int {
	n := 0
	for _, r := range s {
		if r == ' ' || r == '\t' {
			n++
			continue
		}
		break
	}
	return n
}

func trimQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func makeCachedScriptPath(shellPid uint) string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get user cache dir, because %w", err))
	}
	return filepath.Join(cacheDir, appName, fmt.Sprintf("script.%d.bash", shellPid))
}

func readCachedScript(shellPid uint) []string {
	cachePath := makeCachedScriptPath(shellPid)
	cacheFile, err := openFileAndCreateIfNecessaryRecursive(cachePath, os.O_RDONLY, 0777)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to open a cache file: %s, because %w", cachePath, err))
	}
	defer cacheFile.Close()
	cacheBytes, err := io.ReadAll(cacheFile)
	if err != nil && err != io.EOF {
		log.Fatal(fmt.Errorf("failed to read a cache file: %s, because %w", cachePath, err))
	}
	if !utf8.Valid(cacheBytes) {
		log.Fatal(fmt.Errorf("cache file is not valid UTF-8: %s", cachePath))
	}
	cacheStr := string(cacheBytes)
	return strings.Split(cacheStr, "\n")
}

func writeCachedScript(shellPid uint, script []string) {
	cachePath := makeCachedScriptPath(shellPid)
	cacheFile, err := openFileAndCreateIfNecessaryRecursive(cachePath, os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to open a cache file for writing: %s, because %w", cachePath, err))
	}
	defer cacheFile.Close()
	scriptStr := strings.Join(script, "\n")
	_, err = cacheFile.WriteString(scriptStr)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to write to cache file: %s, because %w", cachePath, err))
	}
}

func openFileAndCreateIfNecessaryRecursive(path string, flag int, mode os.FileMode) (*os.File, error) {
	file, err := os.OpenFile(path, flag, mode)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to open the file: %s, because %w", path, err)
		} else {
			file, err = os.Create(path)
			if err != nil {
				if !os.IsNotExist(err) {
					return nil, fmt.Errorf("failed to create the file: %s, because %w", path, err)
				}
				dir := filepath.Dir(path)
				err = os.MkdirAll(dir, 0755)
				if err != nil {
					return nil, fmt.Errorf("failed to create the directory: %s, because %w", dir, err)
				}
				file, err = os.Create(path)
				if err != nil {
					return nil, fmt.Errorf("failed to create the file after creating the directory: %s, because %w", path, err)
				}
			}
		}
	}
	return file, nil
}

// runExecCommand executes a shell command and returns its stdout trimmed of trailing newlines.
func runExecCommand(cmd string) (string, error) {
	c := exec.Command("sh", "-c", cmd)
	out, err := c.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}

//go:embed hook.bash
var hookScript string
