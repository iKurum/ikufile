package task

import (
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"time"

	c "github.com/iKurum/ikufile/config"
	"github.com/iKurum/ikufile/utils/check"
	logs "github.com/iKurum/ikufile/utils/log"

	"github.com/fsnotify/fsnotify"
)

func InitWatch() {
	var err error
	if Watcher != nil {
		_ = Watcher.Close()
	}
	Watcher, err = fsnotify.NewWatcher()
	if err != nil {
		logs.Exit(err)
	}
	TaskMan = NewTaskMan(c.Cfg.Command.DelayMillSecond, c.Cfg.Notifier.CallUrl)
	go func() {
		for {
			select {
			case event, ok := <-Watcher.Events:
				if !ok {
					return
				}
				// directory structure changes, dynamically add, delete and monitor according to rules
				// TODO // this method cannot be triggered when the parent folder of the change folder is not monitored
				go watchChangeHandler(event)
				fmt.Println("\033[H\033[2J")
				eventDispatcher(event)
			case err, ok := <-Watcher.Errors:
				if !ok {
					return
				}
				logs.Error(err)
			}
		}
	}()
	addWatcher()
}

func cmdParse2Array(s string, cf *ChangedFile) []string {
	a := strings.Split(s, " ")
	r := make([]string, 0)
	for i := 0; i < len(a); i++ {
		if ss := strings.Trim(a[i], " "); ss != "" {
			r = append(r, strParseRealStr(ss, cf))
		}
	}
	return r
}

func strParseRealStr(s string, cf *ChangedFile) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(s, "{{file}}", cf.Name),
				"{{ext}}", cf.Ext,
			),
			"{{changed}}", strconv.FormatInt(cf.Changed, 10),
		),
		"{{event}}", cf.Event,
	)
}

func watchChangeHandler(event fsnotify.Event) {
	// stop the fileboy daemon process when the .fileboy.pid file is changed
	if event.Name == c.GetPidFile() &&
		(event.Op == fsnotify.Remove ||
			event.Op == fsnotify.Write ||
			event.Op == fsnotify.Rename) {
		logs.UInfo("exit daemon process")
		c.StopSelf()
		return
	}
	if event.Op != fsnotify.Create && event.Op != fsnotify.Rename {
		return
	}
	_, err := ioutil.ReadDir(event.Name)
	if err != nil {
		return
	}
	do := false
	for rec := range c.Cfg.Monitor.IncludeDirsRec {
		if !strings.HasPrefix(event.Name, rec) {
			continue
		}
		// check exceptDirs
		if hitDirs(event.Name, &c.Cfg.Monitor.ExceptDirs) {
			continue
		}

		_ = Watcher.Remove(event.Name)
		err := Watcher.Add(event.Name)
		if err == nil {
			do = true
			logs.Info("watcher add -> ", event.Name)
		} else {
			logs.Warning("watcher add faild:", event.Name, err)
		}
	}

	if do {
		return
	}

	// check map
	if _, ok := c.Cfg.Monitor.DirsMap[event.Name]; ok {
		_ = Watcher.Remove(event.Name)
		err := Watcher.Add(event.Name)
		if err == nil {
			logs.Info("watcher add -> ", event.Name)
		} else {
			logs.Warning("watcher add faild:", event.Name, err)
		}
	}
}

func eventDispatcher(event fsnotify.Event) {
	if event.Name == c.GetPidFile() {
		return
	}
	ext := path.Ext(event.Name)
	if len(c.Cfg.Monitor.Types) > 0 &&
		!check.KeyInMonitorTypesMap(".*", c.Cfg) &&
		!check.KeyInMonitorTypesMap(ext, c.Cfg) {
		return
	}

	op := c.IoeventMapStr[event.Op]
	if len(c.Cfg.Monitor.Events) != 0 && !inStrArray(op, c.Cfg.Monitor.Events) {
		return
	}
	logs.UInfo("EVENT ", event.Op.String(), ":", event.Name)
	TaskMan.Put(&ChangedFile{
		Name:    relativePath(c.ProjectFolder, event.Name),
		Changed: time.Now().UnixNano(),
		Ext:     ext,
		Event:   op,
	})
}

func addWatcher() {
	logs.Info("collecting directory information...")
	dirsMap := map[string]bool{}
	for _, dir := range c.Cfg.Monitor.ExceptDirs {
		if dir == "." {
			logs.Exit("exceptDirs must is not project root path ! err path:", dir)
		}
	}
	for _, dir := range c.Cfg.Monitor.IncludeDirs {
		darr := dirParse2Array(dir)
		if len(darr) < 1 || len(darr) > 2 {
			logs.Exit("filegirl section monitor dirs is error. ", dir)
		}
		if strings.HasPrefix(darr[0], "/") {
			logs.Exit("dirs must be relative paths ! err path:", dir)
		}
		if darr[0] == "." {
			if len(darr) == 2 && darr[1] == "*" {
				// The highest priority
				dirsMap = map[string]bool{
					c.ProjectFolder: true,
				}
				listFile(c.ProjectFolder, func(d string) {
					dirsMap[d] = true
				})
				c.Cfg.Monitor.IncludeDirsRec[c.ProjectFolder] = true
				break
			} else {
				dirsMap[c.ProjectFolder] = true
			}
		} else {
			md := c.ProjectFolder + "/" + darr[0]
			dirsMap[md] = true
			if len(darr) == 2 && darr[1] == "*" {
				listFile(md, func(d string) {
					dirsMap[d] = true
				})
				c.Cfg.Monitor.IncludeDirsRec[md] = true
			}
		}

	}

	for dir := range dirsMap {
		logs.Info("watcher add -> ", dir)
		err := Watcher.Add(dir)
		if err != nil {
			logs.Exit(err)
		}
	}
	logs.Info("total monitored dirs: " + strconv.Itoa(len(dirsMap)))
	logs.Info("ikufile is ready.")
	c.Cfg.Monitor.DirsMap = dirsMap
}

func hitDirs(d string, dirs *[]string) bool {
	d += "/"
	for _, v := range *dirs {
		if strings.HasPrefix(d, c.ProjectFolder+"/"+v+"/") {
			return true
		}
	}
	return false
}

func inStrArray(s string, arr []string) bool {
	for _, v := range arr {
		if s == v {
			return true
		}
	}
	return false
}

func relativePath(folder, p string) string {
	s := strings.ReplaceAll(strings.TrimPrefix(p, folder), "\\", "/")
	if strings.HasPrefix(s, "/") && len(s) > 1 {
		s = s[1:]
	}
	return s
}

func listFile(folder string, fun func(string)) {
	files, _ := ioutil.ReadDir(folder)
	for _, file := range files {
		if file.IsDir() {
			d := folder + "/" + file.Name()
			if hitDirs(d, &c.Cfg.Monitor.ExceptDirs) {
				continue
			}
			fun(d)
			listFile(d, fun)
		}
	}
}

func dirParse2Array(s string) []string {
	a := strings.Split(s, ",")
	r := make([]string, 0)
	for i := 0; i < len(a); i++ {
		if ss := strings.Trim(a[i], " "); ss != "" {
			r = append(r, ss)
		}
	}
	return r
}
