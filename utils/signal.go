package utils

import (
	"os"
	"os/signal"
	"syscall"

	t "github.com/iKurum/ikufile/task"
	logs "github.com/iKurum/ikufile/utils/log"
)

func SignalHandler() {
	cs := make(chan os.Signal)
	signal.Notify(cs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cs
		if t.TaskMan != nil && t.TaskMan.Cmd != nil && t.TaskMan.Cmd.Process != nil {
			if err := t.TaskMan.Cmd.Process.Kill(); err != nil {
				logs.Warning("stopping the process failed: PID:", t.TaskMan.Cmd.ProcessState.Pid(), ":", err)
			}
		}
		os.Exit(0)
	}()
}
