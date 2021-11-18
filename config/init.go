package config

import (
	"os"
	"os/exec"
	"strconv"
)

func StopSelf() {
	pid := os.Getpid()
	os.Remove(GetPidFile())
	_ = exec.Command("kill", strconv.Itoa(pid)).Run()
}

func GetPidFile() string {
	return ProjectFolder + "/.fileboy.pid"
}
