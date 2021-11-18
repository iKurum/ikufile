package config

import "github.com/fsnotify/fsnotify"

var (
	InstExecWhenStart   = "exec-when-start"
	InstShouldFinish    = "should-finish"
	InstIgnoreWarn      = "ignore-warn"
	InstIgnoreInfo      = "ignore-info"
	InstIgnoreStdout    = "ignore-stdout"
	InstIgnoreExecError = "ignore-exec-error"
)

var (
	ProjectFolder = "."
	YamlName      = ".ikufile.yaml"
	YamlPath      = ""
	IoeventMapStr = map[fsnotify.Op]string{
		fsnotify.Write:  "write",
		fsnotify.Rename: "rename",
		fsnotify.Remove: "remove",
		fsnotify.Create: "create",
		fsnotify.Chmod:  "chmod",
	}
	Cfg *FileIku
)

type FileIku struct {
	Monitor struct {
		Types          []string        `yaml:"types"`
		IncludeDirs    []string        `yaml:"includeDirs"`
		ExceptDirs     []string        `yaml:"exceptDirs"`
		Events         []string        `yaml:"events"`
		TypesMap       map[string]bool `yaml:"-"`
		IncludeDirsMap map[string]bool `yaml:"-"`
		ExceptDirsMap  map[string]bool `yaml:"-"`
		DirsMap        map[string]bool `yaml:"-"`
		IncludeDirsRec map[string]bool `yaml:"-"`
	}
	Command struct {
		Exec            []string `yaml:"exec"`
		DelayMillSecond int      `yaml:"delayMillSecond"`
	}
	Notifier struct {
		CallUrl string `yaml:"callUrl"`
	}
	Instruction    []string        `yaml:"instruction"`
	InstructionMap map[string]bool `yaml:"-"`
}
