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

	"go.yaml.in/yaml/v4"
)

const appName = "envar"

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("arguments must be one or more: %d", len(os.Args)-1)
	}
	switch os.Args[1] {
	case "help":
		fmt.Print(usageMessage)
	case "path":
		if len(os.Args) != 3 {
			log.Fatalf("invalid number of arguments for path: %d", len(os.Args)-1)
		}
		if os.Args[2] != "config" {
			log.Fatalf("unknown path type: %s", os.Args[2])
		}
		configDir, err := os.UserConfigDir()
		if err != nil {
			log.Fatalf("failed to get user config dir, because %v", err)
		}
		path := filepath.Join(configDir, appName)
		fmt.Println(path)
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
	varsConfig, execsConfig, err := readConfigs()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to read configs, because %w", err))
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
	for varName, pathItems := range *varsConfig {
		for _, pathItem := range pathItems {
			path := pathItem.Path
			path = strings.Replace(path, "~", homeDir, 1)
			matched := strings.HasPrefix(workingDirectory, path)
			if matched {
				if pathItem.Exec != nil {
					commandTemplate, ok := (*execsConfig)[pathItem.Exec.Id]
					if !ok {
						log.Fatal(fmt.Errorf("exec reference '%s' not found in execs.yaml for variable %s", pathItem.Exec.Id, varName))
					}
					v, err := runExecCommand(commandTemplate, pathItem.Exec.Args)
					if err != nil {
						log.Fatal(fmt.Errorf("failed to run exec for %s, because %w", varName, err))
					}
					script = append(script, fmt.Sprintf("export %s=%s", varName, v))
				} else if pathItem.Value == nil {
					script = append(script, fmt.Sprintf("unset %s", varName))
				} else {
					script = append(script, fmt.Sprintf("export %s=%s", varName, *pathItem.Value))
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

type VarsConfig = map[VarName][]PathItem

type ExecsConfig = map[ExecId]ExecPattern

type VarName = string

type ExecId = string

type ExecPattern = string

type PathItem struct {
	Path  string
	Value *string   // nil means unset
	Exec  *ExecItem // optional reference to exec command
}

type ExecItem struct {
	Id   ExecId
	Args []string
}

func readConfigs() (*VarsConfig, *ExecsConfig, error) {
	varsBytes, err := readConfig("vars.yaml")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read vars config, because %w", err)
	}
	execsBytes, err := readConfig("execs.yaml")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read execs config, because %w", err)
	}
	config, err := UnmarshalVarsConfig(varsBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal vars config, because %w", err)
	}
	execsConfig, err := UnmarshalExecsConfig(execsBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal execs config, because %w", err)
	}
	return config, execsConfig, nil
}

func readConfig(fileName string) ([]byte, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config dir, because %w", err)
	}
	path := filepath.Join(configDir, appName, fileName)
	file, err := openFileAndCreateIfNecessaryRecursive(path, os.O_RDONLY, 0777)
	if err != nil {
		return nil, fmt.Errorf("failed to open vars config: %s, because %w", path, err)
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read vars config: %s, because %w", path, err)
	}
	return bytes, nil
}

func UnmarshalVarsConfig(bytes []byte) (*VarsConfig, error) {
	if !utf8.Valid(bytes) {
		return nil, fmt.Errorf("config is invalid UTF-8")
	}
	cfg := make(VarsConfig)
	// 空入力は空設定として扱う
	if strings.TrimSpace(string(bytes)) == "" {
		return &cfg, nil
	}
	var root yaml.Node
	if err := yaml.Unmarshal(bytes, &root); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}
	// DocumentNode の直下を取得
	if len(root.Content) == 0 {
		return &cfg, nil
	}
	top := root.Content[0]
	if top.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("top-level yaml must be a mapping")
	}
	// トップレベル：変数名 → マップ
	for i := 0; i < len(top.Content); i += 2 {
		k := top.Content[i]
		v := top.Content[i+1]
		if k.Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("variable name must be a scalar, got kind=%v", k.Kind)
		}
		varName := strings.TrimSpace(k.Value)
		if varName == "" {
			return nil, fmt.Errorf("variable name must not be empty")
		}
		if _, ok := cfg[varName]; !ok {
			cfg[varName] = make([]PathItem, 0)
		}
		// 値が null の場合はスキップ（エントリーなし）
		if v.Kind == yaml.ScalarNode && v.Tag == "!!null" {
			continue
		}
		if v.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("variable '%s' must be a mapping", varName)
		}
		// セカンドレベル：パス → 値
		for j := 0; j+1 < len(v.Content); j += 2 {
			pk := v.Content[j]
			pv := v.Content[j+1]
			if pk.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("path prefix must be a scalar under '%s'", varName)
			}
			path := strings.TrimSpace(pk.Value)
			if path == "" {
				return nil, fmt.Errorf("path must not be empty under '%s'", varName)
			}
			var pathItem PathItem
			pathItem.Path = path
			switch pv.Kind {
			case yaml.ScalarNode:
				// 値がリテラルで書かれているか null が期待される
				if pv.Tag == "!!null" || strings.TrimSpace(pv.Value) == "" {
					pathItem.Value = nil
				} else {
					val := pv.Value
					pathItem.Value = &val
				}
			case yaml.MappingNode:
				// ExecId → 引数 が期待される
				if len(pv.Content) != 2 {
					return nil, fmt.Errorf("nested mapping under path '%s' must have exactly one key-value pair", path)
				}
				nk := pv.Content[0]
				nv := pv.Content[1]
				if nk.Kind != yaml.ScalarNode {
					return nil, fmt.Errorf("exec reference key must be a scalar under path '%s'", path)
				}
				execName := nk.Value
				pathItem.Value = nil
				pathItem.Exec = &ExecItem{Id: execName}
				pathItem.Exec.Args = make([]string, 0, len(nv.Content))
				// 引数の解析
				switch nv.Kind {
				case yaml.ScalarNode:
					// 単一引数
					pathItem.Exec.Args = append(pathItem.Exec.Args, nv.Value)
				case yaml.SequenceNode:
					// 配列引数
					for _, argNode := range nv.Content {
						if argNode.Kind != yaml.ScalarNode {
							return nil, fmt.Errorf("exec arguments must be scalars under path '%s'", path)
						}
						pathItem.Exec.Args = append(pathItem.Exec.Args, argNode.Value)
					}
				default:
					return nil, fmt.Errorf("exec argument must be a scalar or array under path '%s'", path)
				}
			default:
				return nil, fmt.Errorf("unsupported value node kind under path '%s': %v", path, pv.Kind)
			}
			cfg[varName] = append(cfg[varName], pathItem)
		}
	}

	return &cfg, nil
}

func UnmarshalExecsConfig(bytes []byte) (*ExecsConfig, error) {
	if !utf8.Valid(bytes) {
		return nil, fmt.Errorf("execs config is invalid UTF-8")
	}
	cfg := make(ExecsConfig)
	// 空入力は空設定として扱う
	if strings.TrimSpace(string(bytes)) == "" {
		return &cfg, nil
	}
	var root yaml.Node
	if err := yaml.Unmarshal(bytes, &root); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}
	// DocumentNode の直下を取得
	if len(root.Content) == 0 {
		return &cfg, nil
	}
	top := root.Content[0]
	if top.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("top-level yaml must be a mapping")
	}
	// トップレベル：名前 → コマンドテンプレート
	for i := 0; i < len(top.Content); i += 2 {
		k := top.Content[i]
		v := top.Content[i+1]
		if k.Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("exec name must be a scalar, got kind: %v", k.Kind)
		}
		if v.Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("exec command must be a scalar, got kind: %v", v.Kind)
		}
		execName := strings.TrimSpace(k.Value)
		if execName == "" {
			return nil, fmt.Errorf("exec name must not be empty")
		}
		cfg[execName] = v.Value
	}
	return &cfg, nil
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

func runExecCommand(commandTemplate string, args []string) (string, error) {
	var command string
	command = commandTemplate
	for _, arg := range args {
		command = strings.Replace(command, "%s", arg, 1)
	}
	c := exec.Command("sh", "-c", command)
	out, err := c.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}

const usageMessage = "envar\n" +
	"\n" +
	"This is a command-line tool that automatically switches values of environment variables based on the current directory path.\n" +
	"\n" +
	"envar <shell-pid>\n" +
	"  Outputs shell script to set/unset environment variables. Call `eval $(envar $$)`.\n" +
	"envar hook\n" +
	"  Outputs shell hook script. Call `eval $(envar hook)`.\n" +
	"envar hook logout <shell-pid>\n" +
	"  Cleans up cached data.\n" +
	"envar path config\n" +
	"  Displays the path to the configuration directory.\n" +
	"envar help\n" +
	"  Displays this help message.\n" +
	"\n" +
	"https://github.com/kakkun61/envar\n"

//go:embed hook.bash
var hookScript string
