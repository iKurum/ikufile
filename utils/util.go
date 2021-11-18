package utils

import (
	"fmt"
	"io/ioutil"

	c "github.com/iKurum/ikufile/config"
	logs "github.com/iKurum/ikufile/utils/log"

	"gopkg.in/yaml.v2"
)

// 展示信息
func show() {
	fmt.Printf("%s\n", c.Logo)
	fmt.Printf("%s\n\n\n", c.Statement)
}

// 读取yaml
func parseConfig() {
	c.Cfg = new(c.FileIku)
	f, err := ioutil.ReadFile(c.YamlPath)
	if err != nil {
		logs.Error("the ikufile configuration file is not exist! ", err)
		logs.Info(c.FirstRunHelp)
		logs.Exit("ikufile unable to run.")
	}
	err = yaml.Unmarshal(f, c.Cfg)
	if err != nil {
		logs.Exit("parsed fail ", c.YamlPath, ":", err)
	}

	// init map
	c.Cfg.Monitor.TypesMap = map[string]bool{}
	c.Cfg.Monitor.IncludeDirsMap = map[string]bool{}
	c.Cfg.Monitor.ExceptDirsMap = map[string]bool{}
	c.Cfg.Monitor.IncludeDirsRec = map[string]bool{}
	c.Cfg.InstructionMap = map[string]bool{}
	// convert to map
	for _, v := range c.Cfg.Monitor.Types {
		c.Cfg.Monitor.TypesMap[v] = true
	}
	for _, v := range c.Cfg.Instruction {
		c.Cfg.InstructionMap[v] = true
	}
}
