package task

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	c "github.com/iKurum/ikufile/config"
	"github.com/iKurum/ikufile/utils/check"
	logs "github.com/iKurum/ikufile/utils/log"
)

// 直接结束旧进程
func (t *Task) PreRun(cf *ChangedFile) {
	if t.Cmd != nil && t.Cmd.Process != nil {
		logs.Info("[old process] ", t.Cmd.Process.Pid)
		if err := t.Cmd.Process.Kill(); err != nil {
			logs.Warning("stop old process err, reason:", err)
		}
	}
	go t.Run(cf)
	go t.Notifier.Put(cf)
}

// 执行 exec
func (t *Task) Run(cf *ChangedFile) {
	t.runLock.Lock()
	defer t.runLock.Unlock()
	for i := 0; i < len(c.Cfg.Command.Exec); i++ {
		carr := cmdParse2Array(c.Cfg.Command.Exec[i], cf)
		logs.Info("EXEC", carr)
		t.Cmd = exec.Command(carr[0], carr[1:]...)
		t.Cmd.Stdin = os.Stdin
		t.Cmd.Stdout = os.Stdout
		if check.KeyInInstruction(c.InstIgnoreStdout) {
			t.Cmd.Stdout = nil
		}
		t.Cmd.Stderr = os.Stderr
		t.Cmd.Dir = c.ProjectFolder
		t.Cmd.Env = os.Environ()

		err := t.Cmd.Start()
		if err != nil {
			logs.Error("run command", carr, "error. ", err)
			if check.KeyInInstruction(c.InstIgnoreExecError) {
				continue
			}
			break
		}
		err = t.Cmd.Wait()
		if err != nil {
			std, e := t.Cmd.StdinPipe()
			if e != nil {
				logs.Error("StdinPipe failed:", carr, err)
				continue
			}
			err = std.Close()
			logs.Error("command exec failed:", carr, err)
			if check.KeyInInstruction(c.InstIgnoreExecError) {
				continue
			}
			break
		}

		status := t.Cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitStatus := status.ExitStatus()
		signaled := status.Signaled()
		signal := status.Signal()
		if signaled {
			logs.Info("Signal:", signal)
		} else {
			logs.Info("Status:", exitStatus)
		}

		if t.Cmd.Process != nil {
			var b bytes.Buffer
			t.Cmd.Stdout = &b
			err = t.Cmd.Process.Kill()
			logs.Info(t.Cmd.ProcessState)
			if t.Cmd.ProcessState != nil && !t.Cmd.ProcessState.Exited() {
				logs.Error("command cannot stop!", carr, err)
			}
		}
	}

	if check.KeyInInstruction(c.InstShouldFinish) {
		t.Cmd = nil
		t.WaitChan <- true
	}
	logs.Info("EXEC end")
}

// 文件修改 转入
func (t *Task) Put(cf *ChangedFile) {
	if t.Delay < 1 {
		t.dispatcher(cf)
		return
	}
	t.putLock.Lock()
	defer t.putLock.Unlock()
	t.LastTaskID = cf.Changed
	go func() {
		<-time.After(time.Millisecond * time.Duration(t.Delay))
		if t.LastTaskID > cf.Changed {
			return
		}
		t.dispatcher(cf)
	}()
}

// 判断是否需要等待上一个进程执行完
func (t *Task) dispatcher(cf *ChangedFile) {
	if check.KeyInInstruction(c.InstShouldFinish) {
		t.WaitQueue = append(t.WaitQueue, cf)
		if t.Cmd == nil {
			t.WaitChan <- true
			return
		}
		logs.Info("waitting for the last task to finish")
		logs.Info("waiting tasks:", len(t.WaitQueue))
	} else {
		t.PreRun(cf)
	}
}

func (n *NetNotifier) Put(cf *ChangedFile) {
	if !n.CanPost {
		logs.Warning("notifier call url ignore. ", n.CallUrl)
		return
	}
	n.dispatch(&PostParams{
		ProjectFolder: c.ProjectFolder,
		File:          cf.Name,
		Changed:       cf.Changed,
		Ext:           cf.Ext,
		Event:         cf.Event,
	})
}

func (n *NetNotifier) dispatch(params *PostParams) {
	b, err := json.Marshal(params)
	if err != nil {
		logs.Error("json.Marshal n.params. ", err)
		return
	}
	client := http.DefaultClient
	client.Timeout = time.Second * 15
	req, err := http.NewRequest("POST", n.CallUrl, bytes.NewBuffer(b))
	if err != nil {
		logs.Error("http.NewRequest. ", err)
		return
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("User-Agent", "FileBoy Net Notifier v1.16")
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("notifier call failed. err:", err)
		return
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	// if resp.StatusCode >= 300 {
	// todo retry???
	// }
	logs.Error("notifier done .")
}
