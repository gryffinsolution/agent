package netty

import (
	"database/sql"
	"log"
	"net/http"
	"util/dao"
)

var DB *sql.DB

func getEvent(res http.ResponseWriter, req *http.Request) {
	log.Println(req)
	ret := dao.GetEventData(DB)
	log.Println(ret)
	res.Write([]byte(ret))
}

func setLastEventID(res http.ResponseWriter, req *http.Request) {

	log.Println("GET params were:", req.URL.Query())

	param1 := req.URL.Query().Get("eventID")
	if param1 != "" {
		dao.SetEventEndTimestamp(DB, param1)
	}
	log.Println(req)
	res.Write([]byte("fuck1"))
}

func getHostDataAgntMgr(res http.ResponseWriter, req *http.Request) {
	log.Println(req)
	ret := dao.GetHostDataAgntMgr(DB)
	log.Println(ret)
	res.Write([]byte(ret))
}

//func healthCheck(res http.ResponseWriter, req *http.Request) {
//	log.Println(req)
//	res.Write([]byte("fuck1"))
//}

func StartSvr(db *sql.DB) {
	DB = db
	//http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/getEvent", getEvent)
	http.HandleFunc("/setLastEventId", setLastEventID)
	http.HandleFunc("/getHostDataAgntMgr", getHostDataAgntMgr)
	http.HandleFunc("/getHostInfos", getHostDataAgntMgr)
	http.ListenAndServe(":42320", nil)
}
