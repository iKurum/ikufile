package main

import (
	"log"
	"os"

	c "github.com/iKurum/ikufile/config"
	u "github.com/iKurum/ikufile/utils"
	logs "github.com/iKurum/ikufile/utils/log"
)

func main() {
	log.SetPrefix("[IKUFILE]~")
	log.SetFlags(2)
	log.SetOutput(os.Stdout)

	var err error
	c.ProjectFolder, err = os.Getwd()
	if err != nil {
		logs.Exit(err)
	}
	u.SignalHandler()
	u.RebuildArgs()
}
