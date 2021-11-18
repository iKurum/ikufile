package daemon

import (
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	c "github.com/iKurum/ikufile/config"
	logs "github.com/iKurum/ikufile/utils/log"
)

func getPidFile() string {
	return c.ProjectFolder + "/.ikufile.pid"
}

func RunAsDaemon() (int, error) {
	if runtime.GOOS == "windows" {
		logs.Exit("daemons mode cannot run on windows.")
	}
	err := StopDaemon()
	if err != nil {
		logs.Exit(err)
	}
	_, err = exec.LookPath("ikufile")
	if err != nil {
		logs.Exit("cannot found `ikufile` command in the PATH")
	}
	daemon := exec.Command("ikufile")
	daemon.Dir = c.ProjectFolder
	daemon.Env = os.Environ()
	daemon.Stdout = os.Stdout
	err = daemon.Start()
	if err != nil {
		logs.Exit(err)
	}
	pid := daemon.Process.Pid
	if pid != 0 {
		ioutil.WriteFile(getPidFile(), []byte(strconv.Itoa(pid)), 0644)
	}
	return pid, nil
}

func StopDaemon() error {
	bs, err := ioutil.ReadFile(getPidFile())
	if err != nil {
		return nil
	}
	_ = exec.Command("kill", string(bs)).Run()
	os.Remove(getPidFile())
	return nil
}
