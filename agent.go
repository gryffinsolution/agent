package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"util/dao"
	"util/jsonconf"
	"util/netty"
	"util/plugin"

	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {

	var runningInterval int64
	runningInterval = 300 //default

	confFile := flag.String("conf", "config.json", "config file for agent")
	isDebug := flag.Bool("debug", false, "debug   mode : console output\nrelease mode : logfile output")
	*isDebug = true //TEST force debug mode

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	//log.SetFlags(log.LstdFlags)

	flag.Parse()

	log.SetOutput(&lumberjack.Logger{
		Filename:   "logfile.log", //CONFIG
		MaxSize:    500,           //megabytes //CONFIG
		MaxBackups: 3,
		MaxAge:     1, //days //CONFIG
	})
	if *isDebug == true {
		log.SetOutput(os.Stdout)
		log.SetOutput(os.Stderr)
	}

	version, runningInterval, keepingDays, port, loadAvgLimit, dbPath, cmdRetry, pluginsJson, autoMgr, err := jsonconf.GetGnrlInfo(*confFile, *isDebug)
	fmt.Println(version, runningInterval, keepingDays, port, loadAvgLimit, dbPath, pluginsJson)
	fmt.Println(pluginsJson)

	if err != nil {
		log.Fatal(err)
	}
	log.Println(runningInterval)
	plgNames, plgPath, plgInterval, plgTimeout, plgIsAppend, plgVersion, plgColumnData := jsonconf.GetPlgInfo(pluginsJson)
	//	var isTextColumn map[string]bool
	//	isTextColumn = make(map[string]bool)
	//	makeColumnInfo(plgColumnData, isTextColumn)

	conn, err := dao.GetConnection(dbPath)
	var isDbFileExist bool
	if err != nil {
		isDbFileExist = false
	} else {
		isDbFileExist = true
	}

	dao.ChkTbls(conn, isDbFileExist, plgNames, plgColumnData)

	defer dao.DisConnect(conn)
	dao.InitAgentMgr(conn, version, keepingDays)
	dao.InitUpdateTbls(conn, plgNames, plgIsAppend, plgColumnData)
	dao.InsertEvent(conn, "AGENT000", "INFO", "Agent was started")

	go netty.StartSvr(conn, port, autoMgr)

	var loopCnt int64
	var norTime int64
	var epNow int64
	var norTime int64

	deleteCheckMaxCnt := 86400 / runningInterval
	log.Println("deleteCheckMaxCnt =", deleteCheckMaxCnt)
	loopCnt = 1

	for {
		//TODO check config.json update and do restart
		dao.SetLastUpdatedTime(conn)
		epNow = time.Now().Unix()
		log.Println("epNow  =", epNow)
		norTime = epNow - epNow%runningInterval

		for _, pluginName := range plgNames {
			log.Println(norTime, "%", plgInterval[pluginName], "=", norTime%int64(plgInterval[pluginName]))

			if norTime%int64(plgInterval[pluginName]) == 0 {
				plugin.RunPlugin(conn, pluginName, plgPath[pluginName], plgTimeout[pluginName], plgIsAppend[pluginName])
			} else {
				log.Println("skip=", pluginName, "epNow=", epNow, "sleepTime=", sleepTime, "interval=", plgInterval[pluginName])
			}
		}

		if loopCnt >= deleteCheckMaxCnt || loopCnt == 0 {
			dao.DeleteData(conn, plgNames, keepingDays, epNow)
			loopCnt = 1
		}
		loopCnt++

		epNow = time.Now().Unix()
		log.Println("epNow =", epNow)
		sleepTime = runningInterval - epNow%runningInterval

		log.Println("sleep seconds=", sleepTime)
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
}
