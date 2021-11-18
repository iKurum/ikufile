package check

import c "github.com/iKurum/ikufile/config"

func KeyInMonitorTypesMap(k string, cfg *c.FileIku) bool {
	_, ok := c.Cfg.Monitor.TypesMap[k]
	return ok
}

func KeyInInstruction(k string) bool {
	_, ok := c.Cfg.InstructionMap[k]
	return ok
}
