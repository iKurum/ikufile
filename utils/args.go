package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	c "github.com/iKurum/ikufile/config"
	"github.com/iKurum/ikufile/daemon"
	"github.com/iKurum/ikufile/task"
	logs "github.com/iKurum/ikufile/utils/log"
)

// rebuild args
func RebuildArgs() {
	c.YamlPath = c.ProjectFolder + "/" + c.YamlName

	loadI := -1
	for i, arg := range os.Args {
		if arg == "-Y" || arg == "-y" || arg == "--yaml" {
			loadI = i
		}
	}
	if loadI == -1 {
		parseArgs()
		return
	}

	if len(os.Args) == loadI+1 {
		logs.Exit("unknown ikufile.yaml path, use `ikufile help` show help info.")
	}

	yp := os.Args[loadI+1]
	if path.IsAbs(yp) {
		c.YamlPath = yp
	} else {
		c.YamlPath = c.ProjectFolder + "/" + yp
	}

	argsCopy := make([]string, len(os.Args))
	copy(argsCopy, os.Args)
	os.Args = []string{}
	for i, v := range argsCopy {
		if i != loadI && i != loadI+1 {
			os.Args = append(os.Args, v)
		}
	}

	parseArgs()
}

func parseArgs() {
	switch {
	case len(os.Args) == 1:
		show()
		parseConfig()
		done := make(chan bool)
		task.InitWatch()
		defer task.Watcher.Close()
		<-done
		return
	case len(os.Args) > 1:
		cs := os.Args[1]
		switch cs {
		case "daemon":
			pid, err := daemon.RunAsDaemon()
			if err != nil {
				logs.Exit(err)
			}
			logs.UInfo("PID:", pid)
			logs.UInfo("ikufile is ready. the main process will run as a daemons")
			return
		case "stop":
			err := daemon.StopDaemon()
			if err != nil {
				logs.Exit(err)
			}
			logs.UInfo("ikufile daemon is stoped.")
			return
		case "init":
			_, err := ioutil.ReadFile(c.YamlPath)
			if err == nil {
				logs.Error("profile ikufile.yaml already exists.")
				logs.Exit("delete it first when you want to regenerate ikufile conf file")
			}
			err = ioutil.WriteFile(c.YamlPath, []byte(c.ExampleFileGirl), 0644)
			if err != nil {
				logs.Error("profile ikufile create failed! ", err)
				return
			}
			logs.Info("profile ikufile.yaml created ok")
			return
		case "exec":
			parseConfig()
			task.NewTaskMan(0, c.Cfg.Notifier.CallUrl).Run(new(task.ChangedFile))
			return
		case "version", "v", "-v", "--version":
			fmt.Println(c.VersionDesc)
		case "help", "--help", "--h", "-h":
			fmt.Print(c.HelpStr)
		default:
			logs.Exit("unknown parameter, use 'ikufile help' to view available commands")
		}
	default:
		logs.Exit("unknown parameters, use `ikufile help` show help info.")
	}
}
