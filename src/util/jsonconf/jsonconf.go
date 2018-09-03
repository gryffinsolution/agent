package jsonconf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

type Agent struct {
	Mode                  string `json:"mode"`
	Version               string `json:"version"`
	RunningIntervalSecond int64  `json:"runningIntervalSecond"`
	KeepingDays           int    `json:"keepingDays"`
	Port                  int    `json:"port"`
	LoadAvgLimit          int    `json:"loadAvgLimit"`
	DbPath                string `json:"dbPath"`
	AutoManager           string `json:"autoManager"`
	PluginsJson           string `json:"pluginsJson"`
}

type Agents struct {
	Agents []Agent `json:"agent"`
}

func GetGnrlInfo4Agent(jsonConfStr string, isDebug bool) (string, int64, int, int, int, string, string, string, error) {
	log.Println("confJson=", jsonConfStr)
	jsonFile, err := os.Open(jsonConfStr)
	if err != nil {
		log.Println("general configuration file open error")
		os.Exit(1)
	}
	fmt.Println("Successfully Opened " + jsonConfStr)
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var agents Agents
	defer jsonFile.Close()
	json.Unmarshal(byteValue, &agents)
	var retVersion string
	var retRunningInterval int64
	var retKeepingDays int
	var retNetPort int
	var retLoadAvgLimit int
	var retDbPath string
	var retAutoManager string
	var retPluginsFile string
	isMatched := false
	for i := 0; i < len(agents.Agents); i++ {
		modeFrJson := agents.Agents[i].Mode
		log.Println("value:", isDebug, modeFrJson)
		retVersion = agents.Agents[i].Version
		retRunningInterval = agents.Agents[i].RunningIntervalSecond
		retKeepingDays = agents.Agents[i].KeepingDays
		retNetPort = agents.Agents[i].Port
		retLoadAvgLimit = agents.Agents[i].LoadAvgLimit
		retDbPath = agents.Agents[i].DbPath
		retAutoManager = agents.Agents[i].AutoManager
		retPluginsFile = agents.Agents[i].PluginsJson

		if isDebug == true && modeFrJson == "debug" {
			log.Println("debug mode:", isDebug, ", start with ", modeFrJson, "mode")
			isMatched = true
			break
		} else if isDebug == false && modeFrJson == "release" {
			log.Println("debug mode:", isDebug, ", start with ", modeFrJson, "mode")
			isMatched = true
			break
		} else {
			//log.Fatal("please check your configuration file about running (debug|release)			 mode")
			//continue
		}

	}
	if isMatched == false {
		log.Fatal("please check your configuration file about running (debug|release)			 mode")
		os.Exit(1)
	}
	return retVersion, retRunningInterval, retKeepingDays, retNetPort, retLoadAvgLimit, retDbPath, retPluginsFile, retAutoManager, err

}

type Plugin struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	IntervalSec  int    `json:"intervalSec"`
	TimeoutSec   int    `json:"timeoutSec"`
	Append       bool   `json:"append"`
	Version      string `json:"version"`
	ColumnMapStr string `json:"columns"`
}

type Plugins struct {
	Plugins []Plugin `json:"plugins"`
}

func GetPlgInfo(jsonPluginsConfStr string) ([]string, map[string]string, map[string]int, map[string]int, map[string]bool, map[string]string, map[string]string) {
	log.Println("confPluginsJson=", jsonPluginsConfStr)
	jsonFile, err := os.Open(jsonPluginsConfStr)
	if err != nil {
		log.Println("plugin configuration file open error:")
		os.Exit(1)
	}
	fmt.Println("Successfully Opened " + jsonPluginsConfStr)
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var plugins Plugins
	defer jsonFile.Close()

	json.Unmarshal(byteValue, &plugins)

	var plgNames []string
	var plgPath map[string]string
	plgPath = make(map[string]string)
	var plgInterval map[string]int
	plgInterval = make(map[string]int)
	var plgTimeout map[string]int
	plgTimeout = make(map[string]int)
	var plgAppend map[string]bool
	plgAppend = make(map[string]bool)
	var plgVersion map[string]string
	plgVersion = make(map[string]string)
	var plgColumnData map[string]string
	plgColumnData = make(map[string]string)

	for i := 0; i < len(plugins.Plugins); i++ {

		log.Println("name:" + plugins.Plugins[i].Name)
		plgNames = append(plgNames, plugins.Plugins[i].Name)
		log.Println("path:", plugins.Plugins[i].Path)
		plgPath[plugins.Plugins[i].Name] = plugins.Plugins[i].Path
		log.Println("intervelSec:", strconv.Itoa(plugins.Plugins[i].IntervalSec))
		plgInterval[plugins.Plugins[i].Name] = plugins.Plugins[i].IntervalSec
		log.Println("timeoutSec:", strconv.Itoa(plugins.Plugins[i].TimeoutSec))
		plgTimeout[plugins.Plugins[i].Name] = plugins.Plugins[i].TimeoutSec
		log.Println("append:", plugins.Plugins[i].Append)
		plgAppend[plugins.Plugins[i].Name] = plugins.Plugins[i].Append
		log.Println("version:", plugins.Plugins[i].Version)
		plgVersion[plugins.Plugins[i].Name] = plugins.Plugins[i].Version
		log.Println("columnData:", plugins.Plugins[i].ColumnMapStr)
		plgColumnData[plugins.Plugins[i].Name] = plugins.Plugins[i].ColumnMapStr
	}
	return plgNames, plgPath, plgInterval, plgTimeout, plgAppend, plgVersion, plgColumnData

}
