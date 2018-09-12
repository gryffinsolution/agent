package netty

import (
	"crypto/md5"
	"database/sql"
	"log"
	"net/http"
	//"net/url"
	"os"
	"strconv"
	//"strings"
	"encoding/hex"
	"util/dao"
	"util/plugin"
)

var DB *sql.DB
var AutoMGR string

func getStatus(res http.ResponseWriter, req *http.Request) {
	log.Println(req)
	ret := dao.GetStatus(DB)
	log.Println(ret)
	res.Write([]byte(ret))
}

func getEvent(res http.ResponseWriter, req *http.Request) {
	log.Println(req)
	ret := dao.GetEventData(DB)
	log.Println(ret)
	res.Write([]byte(ret))
}

func getHostDataAgntMgr(res http.ResponseWriter, req *http.Request) {
	log.Println(req)
	ret := dao.GetHostDataAgntMgr(DB)
	log.Println(ret)
	res.Write([]byte(ret))
}

func getHostInfos(res http.ResponseWriter, req *http.Request) {
	log.Println(req)
	ret := dao.GetHostInfo(DB)
	log.Println(ret)
	res.Write([]byte(ret))
}

func getMetrics(res http.ResponseWriter, req *http.Request) {
	log.Println(req)
	paraDate := req.URL.Query().Get("strdate")
	var ret string
	if paraDate != "" {
		ret = dao.GetMetrics(DB, paraDate)
	} else {
		ret = "error"
		log.Println("error:param1 from manager. param data:" + paraDate)
	}
	log.Println(ret)
	res.Write([]byte(ret))
}

func cmdExec(res http.ResponseWriter, req *http.Request) {
	log.Println(req)
	keyV := req.URL.Query().Get("key")
	mJobId := req.URL.Query().Get("mjobid")
	cmdStr := req.URL.Query().Get("cmd")
	timeout := req.URL.Query().Get("timeout")

	var ret string
	if keyV != "" && mJobId != "" && cmdStr != "" && timeout != "" {
		host, _ := os.Hostname()
		log.Println("host", host)
		keyword := "flog3" + host + mJobId + cmdStr + timeout
		log.Println(keyword)
		hasher := md5.New()
		hasher.Write([]byte(keyword))
		strMd5 := hex.EncodeToString(hasher.Sum(nil))
		log.Printf("strKey:", strMd5)
		timeoutInt, err := strconv.Atoi(timeout)
		if err != nil {
			log.Fatal("timeout(" + timeout + ") value is not number")
			ret = "value error"
		} else {
			if keyV == strMd5 {
				if dao.InsertAuto(DB, mJobId, cmdStr, timeoutInt) {
					outStr, status, elapsedTime := plugin.RunCmd(DB, cmdStr, timeoutInt)
					dao.UpdateAuto(DB, mJobId, outStr, status, elapsedTime)
					runTs, endTs := dao.GetJobTimes(DB, mJobId)
					log.Println(runTs, endTs)
					ret = mJobId + ",:," + strconv.FormatInt(runTs, 10) + ",:," + strconv.FormatInt(endTs, 10) + ",:," + status + ",:," + strconv.Itoa(elapsedTime) + "-FLOG-SA-OUT-" + outStr
				} else {
					ret = "error_code1"
				}
			} else {
				ret = "error_code2"
				log.Println("Authentification error")
			}
		}
	} else {
		ret = "error_code3"
		log.Println("parameter error", keyV, mJobId, cmdStr, timeout)
	}
	log.Println(ret)
	res.Write([]byte(ret))
}

func getCustomLog(res http.ResponseWriter, req *http.Request) {
	log.Println(req)
	keyV := req.URL.Query().Get("key")
	tbl := req.URL.Query().Get("tbl")

	var ret string
	if keyV != "" && tbl != "" {
		host, _ := os.Hostname()
		log.Println("host", host)
		keyword := "flog3" + host + tbl
		log.Println(keyword)
		hasher := md5.New()
		hasher.Write([]byte(keyword))
		strMd5 := hex.EncodeToString(hasher.Sum(nil))
		log.Printf("strKey:", strMd5)

		if keyV == strMd5 {
			ret = dao.GetCustomLog(DB, tbl)
		} else {
			ret = "error_code4"
			log.Println("Authentification error")
		}

	} else {
		ret = "error_code5"
		log.Println("parameter error", keyV, tbl)
	}
	log.Println(ret)
	res.Write([]byte(ret))
}

func StartSvr(db *sql.DB, port int, autoMgr string) {
	DB = db
	AutoMGR = autoMgr

	http.HandleFunc("/getstatus", getStatus)
	http.HandleFunc("/getevent", getEvent)
	http.HandleFunc("/gethostdataagntmgr", getHostDataAgntMgr)
	http.HandleFunc("/gethostinfos", getHostInfos)
	http.HandleFunc("/getmetrics", getMetrics)
	http.HandleFunc("/getcustomlog", getCustomLog)
	http.HandleFunc("/cmdexec", cmdExec)

	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
