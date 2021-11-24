package task

import (
	"os/exec"
	"strings"
	"sync"

	c "github.com/iKurum/ikufile/config"
	"github.com/iKurum/ikufile/utils/check"
	notify "github.com/iKurum/ikufile/utils/fs"
	logs "github.com/iKurum/ikufile/utils/log"
)

var (
	TaskMan *Task
	Watcher *notify.Batcher
)

type NetNotifier struct {
	CallUrl string
	CanPost bool
}

type ChangedFile struct {
	Name    string
	Changed int64
	Ext     string
	Event   string
}

type PostParams struct {
	ProjectFolder string `json:"project_folder"`
	File          string `json:"file"`
	Changed       int64  `json:"changed"`
	Ext           string `json:"ext"`
	Event         string `json:"event"`
}

type Task struct {
	LastTaskID int64
	Delay      int
	Cmd        *exec.Cmd
	Notifier   *NetNotifier
	putLock    sync.Mutex
	runLock    sync.Mutex

	WaitChan  chan bool
	WaitQueue []*ChangedFile
}

// init taskMan
func NewTaskMan(delay int, callURL string) *Task {
	t := &Task{
		Delay:     delay,
		Notifier:  newNetNotifier(callURL),
		WaitChan:  make(chan bool, 1),
		WaitQueue: []*ChangedFile{},
	}

	if check.KeyInInstruction(c.InstShouldFinish) {
		go func() {
			for {
				<-t.WaitChan
				if len(t.WaitQueue) < 1 {
					continue
				}
				cf := t.WaitQueue[len(t.WaitQueue)-1]
				if len(t.WaitQueue) > 1 {
					logs.Info("redundant tasks dropped:", len(t.WaitQueue)-1)
				}
				t.WaitQueue = []*ChangedFile{}
				go t.PreRun(cf)
			}
		}()
	}

	return t
}

func newNetNotifier(callUrl string) *NetNotifier {
	callPost := true
	if strings.TrimSpace(callUrl) == "" {
		callPost = false
	}
	return &NetNotifier{
		CallUrl: callUrl,
		CanPost: callPost,
	}
}
