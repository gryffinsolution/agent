// agent project main.go
package dao

import (
	"database/sql"
	"log"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("db nil")
	}
	return db
}

func GetConnection(dbpath string) (*sql.DB, error) {
	_, err := os.Stat(dbpath)
	if err != nil {
		log.Println("db file", dbpath, "is missing. creating...")
	}
	db := InitDB(dbpath)
	return db, err
}

func DisConnect(db *sql.DB) {
	db.Close()
}

func getTables(db *sql.DB) []string {
	sqlGetTables := `SELECT name FROM sqlite_master WHERE type='table'`
	rows, err := db.Query(sqlGetTables)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var tblNames []string
	var name string
	for rows.Next() {

		errScan := rows.Scan(&name)
		if errScan != nil {
			log.Fatal(errScan)
			panic(errScan)
		}
		log.Println("tblName=", name)
		tblNames = append(tblNames, name)
	}
	return tblNames
}

func tblExists(db *sql.DB, tblName string) bool {
	sqlGetTable := "SELECT NAME FROM SQLITE_MASTER WHERE TYPE ='table' AND NAME='" + tblName + "'"
	rows, err := db.Query(sqlGetTable)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var name string
	for rows.Next() {
		errScan := rows.Scan(&name)
		if errScan != nil {
			log.Fatal(errScan)
			panic(errScan)
		}
		log.Println("tblName=", name)
		if name == tblName {
			return true
		}
	}
	return false
}

func chkTbl(db *sql.DB, plgName string, plgColumnData map[string]string) bool {
	log.Println("plugin table check", plgName)

	if tblExists(db, plgName) {
		return true
	} else {
		log.Println("create table")
		sqlCreateTbl := "DROP TABLE IF EXISTS " + plgName + " ;CREATE TABLE " + plgName + " (TIME TIMESTAMP DEFAULT (STRFTIME('%s','now') ) "

		strColumnLine := plgColumnData[plgName]
		kvitems := strings.Split(strColumnLine, ",")
		log.Println("kvitems=", kvitems)
		for i := range kvitems {
			items := strings.Split(kvitems[i], ":")
			sqlCreateTbl = sqlCreateTbl + ", " + items[0] + " " + items[1]
		}
		sqlCreateTbl = sqlCreateTbl + ")"
		log.Println("sql=", sqlCreateTbl)
		_, err := db.Exec(sqlCreateTbl)
		if err != nil {
			log.Println("table creation failed")
			return false
		} else {
			return true
		}
	}
}

func ChkTbls(db *sql.DB, isDbFileExist bool, plgNames []string, plgColumnData map[string]string) {

	log.Println("isDbFileExist flag:", isDbFileExist)

	if !isDbFileExist { //creation tables
		log.Println("AGENT_MGR table creating...")
		sqlCreateTbl := "DROP TABLE IF EXISTS AGENT_MGR ;CREATE TABLE AGENT_MGR (ID INTEGER, AGENT_START_TS TIMESTAMP DEFAULT (STRFTIME('%s','now')), AGENT_LAST_TS TIMESTAMP DEFAULT (STRFTIME('%s','now')), EVT_LAST_TS TIMESTAMP DEFAULT (STRFTIME('%s','now')), AUTO_LAST_TS TIMESTAMP DEFAULT (STRFTIME('%s','now')), LAST_SENT_EVENT_ID INTEGER, AGENT_VER TEXT, DATA_KEEPING_DAYS INTEGER, CMD_RETRY_LIMIT INTEGER, RESTART_AUTO_FLAG INT)"
		_, err := db.Exec(sqlCreateTbl)
		if err != nil {
			log.Println("AGENT_MGR table creation failed" + sqlCreateTbl)
			os.Exit(2)
		}

		log.Println("AGENT_EVENT table creating...")
		sqlCreateTbl = "DROP TABLE IF EXISTS AGENT_EVENT ;CREATE TABLE AGENT_EVENT (ID INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, EVENT_CODE TEXT, SEVERITY TEXT, MESSAGE TEXT, TIME TIMESTAMP DEFAULT (STRFTIME('%s','now')))"
		log.Println("sql=", sqlCreateTbl)
		_, err1 := db.Exec(sqlCreateTbl)
		if err1 != nil {
			log.Println("AGENT_EVENT table creation failed")
			os.Exit(2)
		}

		log.Println("AUTO_JOBS creating...")
		sqlCreateTbl = "DROP TABLE IF EXISTS AUTO_JOBS ;CREATE TABLE AUTO_JOBS (MJOBID TEXT, RUN_TS TIMESTAMP DEFAULT (STRFTIME('%s','now')), END_TS TIMESTAMP, STATUS TEXT, CMD TEXT, TIMEOUT INT, IS_SENT INT, ELAPSED_TIME_SEC INT, STDOUT TEXT)"
		log.Println("sql=", sqlCreateTbl)
		_, err2 := db.Exec(sqlCreateTbl)
		if err2 != nil {
			log.Println("AUTO_JOBS table creation failed")
			os.Exit(2)
		}
	}

	for i, plgName := range plgNames {
		if chkTbl(db, plgName, plgColumnData) {
			log.Println(plgName, "plugin table created")
		} else {
			log.Fatal("table create failed for plugin (", plgName, "). abnormal exit")
			os.Exit(3)
		}
		log.Println(i, plgName)
	}
}

func DeleteData(db *sql.DB, tbls []string, keepdays int, epNow int64) bool {
	tbls = append(tbls, "AGENT_EVENT")
	baseTime := epNow - 60*60*24*int64(keepdays)
	log.Println(baseTime)
	for _, tbl := range tbls {
		log.Println(tbl)
		sqlDeleteOld := "DELETE FROM " + tbl + " WHERE STRFTIME('%s',TIME) <" + strconv.FormatInt(baseTime, 10)
		log.Println(sqlDeleteOld)
		_, err := db.Exec(sqlDeleteOld)
		if err != nil {
			log.Println("old data cleaning failed")
		}
	}
	return true
}

func InsertEvent(db *sql.DB, eventCode string, severity string, message string) bool {
	sql := "INSERT INTO AGENT_EVENT(EVENT_CODE,SEVERITY,MESSAGE) values('" + eventCode + "','" + severity + "','" + message + "')"
	log.Println(sql)
	_, err := db.Exec(sql)
	if err != nil {
		log.Println("EVENT INSERT failed")
		return false
	}
	return true
}

func isAgentMgrNew(db *sql.DB) bool {
	sql := "SELECT ID FROM AGENT_MGR WHERE ID=0"
	rows, err := db.Query(sql)
	if err != nil {
		log.Print("sql=" + sql)
	}
	defer rows.Close()
	var id int
	for rows.Next() {
		errScan := rows.Scan(&id)
		if errScan != nil {
			log.Fatal(errScan)
			panic(errScan)
		}
		log.Println("id=", id)
		return true
	}
	return false
}

func InitAgentMgr(db *sql.DB, agentVer string, dataKeepingDays int) bool {
	if !isAgentMgrNew(db) {
		sql := "INSERT INTO AGENT_MGR (ID, AGENT_LAST_TS, AGENT_VER, DATA_KEEPING_DAYS, CMD_RETRY_LIMIT, LAST_SENT_EVENT_ID) VALUES (0, STRFTIME('%s','now'),'" + agentVer + "'," + strconv.Itoa(dataKeepingDays) + ",1,0)" //1 is running error limit
		log.Println(sql)
		_, err := db.Exec(sql)
		if err != nil {
			log.Println("AGENT_MGR init failed")
			return false
		}
	}
	return true
}

func isUpdateTableNew(db *sql.DB, tbl string) bool {
	sql := "SELECT ID FROM " + tbl + " WHERE ID=0"
	rows, err := db.Query(sql)
	if err != nil {
		log.Print("sql=" + sql)
	}
	defer rows.Close()
	var id int
	for rows.Next() {
		errScan := rows.Scan(&id)
		if errScan != nil {
			log.Fatal(errScan)
			panic(errScan)
		}
		log.Println("id=", id)
		return true
	}
	return false
}

func InitUpdateTbls(db *sql.DB, tbls []string, tblIsAppend map[string]bool, plgColumnData map[string]string) bool {
	for _, tbl := range tbls {
		if tblIsAppend[tbl] == false {
			log.Println(tbl + " checking")
			if !isUpdateTableNew(db, tbl) {
				strColumnLine := plgColumnData[tbl]
				kvitems := strings.Split(strColumnLine, ",")
				log.Println("kvitems=", kvitems)
				sqlMid := "ID"
				sqlPost := "0"
				for i := range kvitems {
					if i == 0 {
						continue
					}
					log.Println(kvitems[i])
					items := strings.Split(kvitems[i], ":")
					sqlMid = sqlMid + ", " + items[0]
					sqlPost = sqlPost + ", null"
				}
				sqlInitUpdate := "INSERT INTO " + tbl + " (" + sqlMid + ") VALUES (" + sqlPost + ")"
				log.Println(sqlInitUpdate)
				_, err := db.Exec(sqlInitUpdate)
				if err != nil {
					log.Println("sqlInitUpdate cleaning failed")
				}
			}
		}
	}
	return true
}

func SetLastUpdatedTime(db *sql.DB) bool {
	sqlUpdate := "UPDATE AGENT_MGR SET AGENT_LAST_TS = STRFTIME('%s','now') WHERE ID=0"
	log.Println(sqlUpdate)
	_, err := db.Exec(sqlUpdate)
	if err != nil {
		log.Println("LAST_UPDATED_TIMESTAMP update failed")
		return false
	}
	return true
}

func InsertData(db *sql.DB, plugin string, outstr string) bool {
	lines := strings.Split(outstr, "\n")
	for _, line := range lines {
		log.Println("line=", line)
		if len(line) < 1 {
			continue
		}
		preSql := "INSERT INTO "
		if strings.HasPrefix(line, "EVENT_CODE") {
			preSql = preSql + "AGENT_EVENT"
		} else {
			preSql = preSql + plugin
		}
		midSql := "("
		postSql := " VALUES ("
		kvitems := strings.Split(line, ",,")
		log.Println("kvitems=", kvitems)
		for i := range kvitems {
			log.Println(kvitems[i])
			items := strings.Split(kvitems[i], "::")
			if len(items[0]) > 0 && len(items[1]) > 0 {
				if i == 0 {
					midSql = midSql + items[0]
					postSql = postSql + "'" + items[1] + "'"
				} else {
					midSql = midSql + "," + items[0]
					postSql = postSql + ",'" + items[1] + "'"
				}
			}
		}
		midSql = midSql + ") "
		postSql = postSql + ")"
		sql := preSql + midSql + postSql
		log.Println(sql)
		_, err := db.Exec(sql)
		if err != nil {
			log.Println(plugin, " ", outstr, "data input failed")
			return false
		}
	}
	return true
}

func UpdateData(db *sql.DB, plugin string, outstr string) bool {
	sql := "UPDATE " + plugin + " SET TIME=STRFTIME('%s','now'),LOG='" + outstr + "' WHERE ID=0"
	log.Println(sql)
	_, err := db.Exec(sql)
	if err != nil {
		log.Println(plugin, outstr, "data input failed")
		return false
	}
	return true
}

func GetStatus(db *sql.DB) string {
	sql := "SELECT STRFTIME('%s',DATETIME(AGENT_LAST_TS,'unixepoch')) TIME FROM AGENT_MGR WHERE ID=0"
	rows, err := db.Query(sql)
	if err != nil {
		log.Fatal("sql=" + sql)
	}
	defer rows.Close()
	var time string
	retString := ""
	for rows.Next() {
		errScan := rows.Scan(&time)
		if errScan != nil {
			log.Fatal(errScan)
		}
		retString += time
	}
	log.Println("msg=", retString)
	return retString
}

func GetEventData(db *sql.DB) string {
	sql := "SELECT ID,EVENT_CODE,SEVERITY,MESSAGE,STRFTIME('%s',DATETIME(TIME,'unixepoch')) TIME FROM AGENT_EVENT WHERE ID>(SELECT LAST_SENT_EVENT_ID FROM AGENT_MGR WHERE ID=0) ORDER BY ID"
	rows, err := db.Query(sql)
	if err != nil {
		log.Fatal("sql=" + sql)
	}
	log.Println("sql=" + sql)
	defer rows.Close()

	var id int
	var eventCode string
	var severity string
	var message string
	var time int64

	retString := ""
	for rows.Next() {
		errScan := rows.Scan(&id, &eventCode, &severity, &message, &time)
		if errScan != nil {
			log.Fatal(errScan)
		}
		retString += strconv.Itoa(id)
		retString += ","
		retString += eventCode
		retString += ","
		retString += severity
		retString += ","
		retString += message
		retString += ","
		retString += strconv.FormatInt(time, 10)
		retString += "\n"
	}
	log.Print("msg=", retString)
	sqlUpdate := "UPDATE AGENT_MGR SET LAST_SENT_EVENT_ID = " + strconv.Itoa(id) + " WHERE ID=0"
	log.Println(sqlUpdate)
	_, err1 := db.Exec(sqlUpdate)
	if err1 != nil {
		log.Fatal("LAST_UPDATED_TIMESTAMP update failed")
	}
	return retString
}

func GetHostDataAgntMgr(db *sql.DB) string {
	sql := "SELECT AGENT_VER, LAST_SENT_EVENT_ID, DATA_KEEPING_DAYS  FROM AGENT_MGR"

	rows, err := db.Query(sql)
	if err != nil {
		log.Fatal("sql=" + sql)
	}
	log.Println("sql=" + sql)
	defer rows.Close()
	var agentVer string
	var lastSentID int
	var keepingDays int

	retString := ""
	for rows.Next() {
		errScan := rows.Scan(&agentVer, &lastSentID, &keepingDays)
		if errScan != nil {
			log.Fatal(errScan)
		}
		retString += agentVer
		retString += ","
		retString += strconv.Itoa(lastSentID)
		retString += ","
		retString += strconv.Itoa(keepingDays)
	}
	log.Print("msg=", retString)
	return retString
}

func GetHostInfo(db *sql.DB) string {
	sql := "SELECT * FROM HOST_INFO ORDER BY TIME DESC"

	rows, err := db.Query(sql)
	if err != nil {
		log.Fatal("sql=" + sql)
	}
	log.Println("sql=" + sql)
	defer rows.Close()

	retString := ""
	for rows.Next() {
		columns, _ := rows.Columns()
		values := make([][]byte, len(columns))
		scanArgs := make([]interface{}, len(values))

		for i := range values {
			scanArgs[i] = &values[i]
		}
		rows.Scan(scanArgs...)

		var value string
		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
				retString = retString + columns[i] + "::" + value
				log.Println(columns[i], value)
			}
			retString = retString + ",,"
		}
		break
	}
	log.Println("msg=", retString)
	return retString
}

func GetMetrics(db *sql.DB, epTime string) string {

	retString := ""

	var time int64
	var cpuIrq float64
	var cpuNice float64
	var cpuSoftirq float64
	var cpuSystem float64
	var cpuIowait float64
	var cpuUser float64
	sqlCpu := "SELECT STRFTIME('%s',DATETIME(TIME,'unixepoch')) TIME,CPU_IRQ,CPU_NICE,CPU_SOFTIRQ,CPU_SYSTEM,CPU_IOWAIT,CPU_USER FROM CPU WHERE TIME>=" + epTime
	rowsCpu, errCpu := db.Query(sqlCpu)
	if errCpu != nil {
		log.Fatal("sqlCpuError=" + sqlCpu)
	}
	log.Println("sqlCpuNormal=" + sqlCpu)
	defer rowsCpu.Close()
	for rowsCpu.Next() {
		errScan := rowsCpu.Scan(&time, &cpuIrq, &cpuNice, &cpuSoftirq, &cpuSystem, &cpuIowait, &cpuUser)
		if errScan != nil {
			log.Fatal(errScan)
		}
		retString += strconv.FormatInt(time, 10)
		retString += ",,"
		retString += strconv.FormatFloat(cpuIrq, 'f', 2, 64)
		retString += ",,"
		retString += strconv.FormatFloat(cpuNice, 'f', 2, 64)
		retString += ",,"
		retString += strconv.FormatFloat(cpuSoftirq, 'f', 2, 64)
		retString += ",,"
		retString += strconv.FormatFloat(cpuSystem, 'f', 2, 64)
		retString += ",,"
		retString += strconv.FormatFloat(cpuIowait, 'f', 2, 64)
		retString += ",,"
		retString += strconv.FormatFloat(cpuUser, 'f', 2, 64)
		retString += "\n"
	}
	log.Print("cpuMsg=", retString)

	retString += "-FLOG-CPU-"

	var load1 float64
	var load5 float64
	var load15 float64
	sqlCpuLoad := "SELECT STRFTIME('%s',DATETIME(TIME,'unixepoch')) TIME,LOAD1,LOAD5,LOAD15 FROM CPU_LOAD WHERE TIME>=" + epTime
	rowsCpuLoad, errCpuLoad := db.Query(sqlCpuLoad)
	if errCpuLoad != nil {
		log.Fatal("sqlCpuLoadError=" + sqlCpuLoad)
	}
	log.Println("sqlCpuLoadNormal=" + sqlCpuLoad)
	defer rowsCpuLoad.Close()
	for rowsCpuLoad.Next() {
		errScan := rowsCpuLoad.Scan(&time, &load1, &load5, &load15)
		if errScan != nil {
			log.Fatal(errScan)
		}
		retString += strconv.FormatInt(time, 10)
		retString += ",,"
		retString += strconv.FormatFloat(load1, 'f', 2, 64)
		retString += ",,"
		retString += strconv.FormatFloat(load5, 'f', 2, 64)
		retString += ",,"
		retString += strconv.FormatFloat(load15, 'f', 2, 64)
		retString += "\n"
	}
	log.Print("cpuLoadMsg=", retString)

	retString += "-FLOG-CPULOAD-"

	var memTotal float64
	var swapTotal float64
	var memFree float64
	var buffers float64
	var swapFree float64
	var cached float64
	sqlMem := "SELECT STRFTIME('%s',DATETIME(TIME,'unixepoch')) TIME,MEMTOTAL,SWAPTOTAL,MEMFREE,BUFFERS,SWAPFREE,CACHED FROM MEM WHERE TIME>=" + epTime
	rowsMem, errMem := db.Query(sqlMem)
	if errMem != nil {
		log.Fatal("sqlMemError=" + sqlMem)
	}
	log.Println("sqlMemNormal" + sqlMem)
	defer rowsMem.Close()
	for rowsMem.Next() {
		errScan := rowsMem.Scan(&time, &memTotal, &swapTotal, &memFree, &buffers, &swapFree, &cached)
		if errScan != nil {
			log.Fatal(errScan)
		}
		retString += strconv.FormatInt(time, 10)
		retString += ",,"
		retString += strconv.FormatFloat(memTotal, 'f', 0, 64)
		retString += ",,"
		retString += strconv.FormatFloat(swapTotal, 'f', 0, 64)
		retString += ",,"
		retString += strconv.FormatFloat(memFree, 'f', 0, 64)
		retString += ",,"
		retString += strconv.FormatFloat(buffers, 'f', 0, 64)
		retString += ",,"
		retString += strconv.FormatFloat(swapFree, 'f', 0, 64)
		retString += ",,"
		retString += strconv.FormatFloat(cached, 'f', 0, 64)
		retString += "\n"
	}
	log.Print("memMsg=", retString)

	retString += "-FLOG-MEM-"

	var devName string
	var capacity float64
	var totalKbytes float64
	var usedKbytes float64
	sqlDisk := "SELECT STRFTIME('%s',DATETIME(TIME,'unixepoch')) TIME,DEV_NAME,CAPACITY,TOTAL_KBYTES,USED_KBYTES FROM DISK WHERE TIME>=" + epTime
	rowsDisk, errDisk := db.Query(sqlDisk)
	if errDisk != nil {
		log.Fatal("sqlDiskError=" + sqlDisk)
	}
	log.Println("sqlDiskNormal=" + sqlDisk)
	defer rowsDisk.Close()
	for rowsDisk.Next() {
		errScan := rowsDisk.Scan(&time, &devName, &capacity, &totalKbytes, &usedKbytes)
		if errScan != nil {
			log.Fatal(errScan)
		}
		retString += strconv.FormatInt(time, 10)
		retString += ",,"
		retString += devName
		retString += ",,"
		retString += strconv.FormatFloat(capacity, 'f', 0, 64)
		retString += ",,"
		retString += strconv.FormatFloat(totalKbytes, 'f', 0, 64)
		retString += ",,"
		retString += strconv.FormatFloat(usedKbytes, 'f', 0, 64)
		retString += "\n"
	}
	log.Print("diskMsg=", retString)

	retString += "-FLOG-DISK-"

	var kbRead float64
	var kbReadPSec float64
	var kbWrtn float64
	var kbWrtnPSec float64
	var tps float64
	sqlDiskIo := "SELECT STRFTIME('%s',DATETIME(TIME,'unixepoch')) TIME,DEV_NAME,KB_READ,KB_READPSEC,KB_WRTN,KB_WRTNPSEC,TPS FROM DISK_IO WHERE TIME>=" + epTime
	rowsDiskIo, errDiskIo := db.Query(sqlDiskIo)
	if errDiskIo != nil {
		log.Fatal("sqlDiskIoError=" + sqlDiskIo)
	}
	log.Println("sqlDiskIoNormal=" + sqlDiskIo)
	defer rowsDiskIo.Close()
	for rowsDiskIo.Next() {
		errScan := rowsDiskIo.Scan(&time, &devName, &kbRead, &kbReadPSec, &kbWrtn, &kbWrtnPSec, &tps)
		if errScan != nil {
			log.Fatal(errScan)
		}
		retString += strconv.FormatInt(time, 10)
		retString += ",,"
		retString += devName
		retString += ",,"
		retString += strconv.FormatFloat(kbRead, 'f', 0, 64)
		retString += ",,"
		retString += strconv.FormatFloat(kbReadPSec, 'f', 0, 64)
		retString += ",,"
		retString += strconv.FormatFloat(kbWrtn, 'f', 0, 64)
		retString += ",,"
		retString += strconv.FormatFloat(kbWrtnPSec, 'f', 0, 64)
		retString += ",,"
		retString += strconv.FormatFloat(tps, 'f', 0, 64)
		retString += "\n"
	}
	log.Print("diskIoMsg=", retString)

	retString += "-FLOG-DISK_IO-"

	var rxErrs float64
	var tx float64
	var frame float64
	var txErrs float64
	var colls float64
	var txPackets float64
	var rxPackets float64
	var txDrop float64
	var rx float64
	var rxDrop float64
	sqlNetworkIo := "SELECT STRFTIME('%s',DATETIME(TIME,'unixepoch')) TIME,DEV_NAME,RX_ERRS,TX,FRAME,TX_ERRS,COLLS,TX_PACKETS,RX_PACKETS,TX_DROP,RX,RX_DROP FROM NETWORK_IO WHERE TIME>=" + epTime
	rowsNetworkIo, errNetworkIo := db.Query(sqlNetworkIo)
	if errNetworkIo != nil {
		log.Fatal("sqlNetworkIoError=" + sqlNetworkIo)
	}
	log.Println("sqlNetworkIoNormal=" + sqlNetworkIo)
	defer rowsNetworkIo.Close()
	for rowsNetworkIo.Next() {
		errScan := rowsNetworkIo.Scan(&time, &devName, &rxErrs, &tx, &frame, &txErrs, &colls, &txPackets, &rxPackets, &txDrop, &rx, &rxDrop)
		if errScan != nil {
			log.Fatal(errScan)
		}
		retString += strconv.FormatInt(time, 10)
		retString += ",,"
		retString += devName
		retString += ",,"
		retString += strconv.FormatFloat(rxErrs, 'f', 1, 64)
		retString += ",,"
		retString += strconv.FormatFloat(tx, 'f', 0, 64)
		retString += ",,"
		retString += strconv.FormatFloat(frame, 'f', 1, 64)
		retString += ",,"
		retString += strconv.FormatFloat(txErrs, 'f', 1, 64)
		retString += ",,"
		retString += strconv.FormatFloat(colls, 'f', 1, 64)
		retString += ",,"
		retString += strconv.FormatFloat(txPackets, 'f', 0, 64)
		retString += ",,"
		retString += strconv.FormatFloat(rxPackets, 'f', 0, 64)
		retString += ",,"
		retString += strconv.FormatFloat(txDrop, 'f', 1, 64)
		retString += ",,"
		retString += strconv.FormatFloat(rx, 'f', 0, 64)
		retString += ",,"
		retString += strconv.FormatFloat(rxDrop, 'f', 1, 64)
		retString += "\n"
	}
	log.Print("networkIoMsg=", retString)

	retString += "-FLOG-NETWORK-"

	var pid string
	var cmd string
	var pcpu float64
	var rss float64
	var userName string
	var vsz float64
	sqlProcessCpu := "SELECT STRFTIME('%s',DATETIME(TIME,'unixepoch')) TIME,PID,CMD,PCPU,RSS,USER_NAME,VSZ FROM PROCESS_CPUINFO WHERE TIME>=" + epTime
	rowsProcessCpu, errProcessCpu := db.Query(sqlProcessCpu)
	if errProcessCpu != nil {
		log.Fatal("sqlProcessCpuError=" + sqlProcessCpu)
	}
	log.Println("sqlProcessCpuNormal=" + sqlProcessCpu)
	defer rowsProcessCpu.Close()
	for rowsProcessCpu.Next() {
		errScan := rowsProcessCpu.Scan(&time, &pid, &cmd, &pcpu, &rss, &userName, &vsz)
		if errScan != nil {
			log.Fatal(errScan)
		}
		retString += strconv.FormatInt(time, 10)
		retString += ",,"
		retString += pid
		retString += ",,"
		retString += cmd
		retString += ",,"
		retString += strconv.FormatFloat(pcpu, 'f', 1, 64)
		retString += ",,"
		retString += strconv.FormatFloat(rss, 'f', 0, 64)
		retString += ",,"
		retString += userName
		retString += ",,"
		retString += strconv.FormatFloat(vsz, 'f', 0, 64)
		retString += "\n"
	}
	log.Print("processCpuMsg=", retString)

	retString += "-FLOG-PROCESS_CPU-"

	var filerName string
	var value float64
	var dataType string
	var userID string
	sqlNfs := "SELECT STRFTIME('%s',DATETIME(TIME,'unixepoch')) TIME,FILER_NAME,VALUE,DATA_TYPE,USER_ID FROM NFS WHERE TIME>=" + epTime
	rowsNfs, errNfs := db.Query(sqlNfs)
	if errNfs != nil {
		log.Fatal("sqlNfsError=" + sqlNfs)
	}
	log.Println("sqlNfsNormal=" + sqlNfs)
	defer rowsNfs.Close()
	for rowsNfs.Next() {
		errScan := rowsNfs.Scan(&time, &filerName, &value, &dataType, &userID)
		if errScan != nil {
			log.Fatal(errScan)
		}
		retString += strconv.FormatInt(time, 10)
		retString += ",,"
		retString += filerName
		retString += ",,"
		retString += strconv.FormatFloat(value, 'f', 1, 64)
		retString += ",,"
		retString += dataType
		retString += ",,"
		retString += userID
		retString += "\n"
	}
	log.Print("nfsMsg=", retString)
	return retString
}

func InsertAuto(db *sql.DB, mJobId string, cmdStr string, timeoutInt int) bool {

	sql := "INSERT INTO AUTO_JOBS (MJOBID, CMD, TIMEOUT, IS_SENT) VALUES ('" + mJobId + "','" + cmdStr + "'," + strconv.Itoa(timeoutInt) + ",0)"
	log.Println(sql)
	_, err := db.Exec(sql)
	if err != nil {
		log.Println("JOB INSERT failed")
		InsertEvent(db, "RC000", "ERROR", "mjobid "+mJobId+" insertion error")
		return false
	}
	return true
}

func UpdateAuto(db *sql.DB, mJobId string, stdOut string, status string, elapsed_time_sec int) bool {
	stdOut = strings.Replace(stdOut, "'", "''", -1)
	sql := "UPDATE AUTO_JOBS SET STDOUT='" + stdOut + "',STATUS='" + status + "',ELAPSED_TIME_SEC=" + strconv.Itoa(elapsed_time_sec) + ",END_TS=STRFTIME('%s','now') WHERE MJOBID='" + mJobId + "'"
	log.Println(sql)
	_, err := db.Exec(sql)
	if err != nil {
		log.Println("JOB UPDATE failed")
		InsertEvent(db, "RC002", "ERROR", "mjobid "+mJobId+" update error")
		return false
	}
	return true
}

func GetJobTimes(db *sql.DB, mJobId string) (int64, int64) {
	sql := "SELECT STRFTIME ('%s',DATETIME(RUN_TS,'unixepoch')) RUN_TS, STRFTIME('%s',DATETIME(END_TS,'unixepoch')) END_TS FROM AUTO_JOBS WHERE MJOBID='" + mJobId + "'"
	rows, err := db.Query(sql)
	if err != nil {
		log.Fatal("sql=" + sql)
	}
	log.Println("sql=" + sql)

	defer rows.Close()

	var runTs int64
	var endTs int64

	for rows.Next() {
		errScan := rows.Scan(&runTs, &endTs)
		if errScan != nil {
			log.Fatal(errScan)
		}
	}
	sqlUpdate := "UPDATE AUTO_JOBS SET IS_SENT = 1 WHERE MJOBID='" + mJobId + "'"
	log.Println(sqlUpdate)
	_, err1 := db.Exec(sqlUpdate)
	if err1 != nil {
		log.Fatal("IS_SENT update failed")
	}
	return runTs, endTs
}

//DROP TABLE IF EXISTS AGENT_EKimkhVENT ;CREATE TABLE AGENT_EVENT (ID INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,  EVENT_CODE TEXT, SEVERITY TEXT, MESSAGE TEXT, TIME INTEGER64 DEFAULT (cast(strftime('%s','now') as int64)))
//insert into agent_event (id) values(1)
//select time from agent_event
//select  strftime('%s',time) tt from agent_event

//DROP TABLE IF EXISTS AGENT_EVENT ;CREATE TABLE AGENT_EVENT (ID INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,  EVENT_CODE TEXT, SEVERITY TEXT, MESSAGE TEXT, time timestamp(6) default current_timestamp)
//insert into agent_event (id) values(1)
//select datetime(time,'localtime') localtime from agent_event
//select strftime('%s',time) tt from agent_event where strftime('%s',time) > 1529675696

//select   strftime('%s',time) test from agent_event

//func main() {
//	const dbpath = "foo.db"

//	db := InitDB(dbpath)
//	defer db.Close()
//	CreateTable(db)

//	items := []TestItem{
//		TestItem{"1", "A", "213"},
//		TestItem{"2", "B", "214"},
//	}
//	StoreItem(db, items)

//	readItems := ReadItem(db)
//	fmt.Println(readItems)

//	items2 := []TestItem{
//		TestItem{"1", "C", "215"},
//		TestItem{"3", "D", "216"},
//	}
//	StoreItem(db, items2)

//	readItems2 := ReadItem(db)
//	fmt.Println(readItems2)
//}

//select distinct(server_name) from
//(
//	select server_name from t_server_ud where time_stamp > to_date ('2018-04-08 00:00:00','YYYY-MM-DD HH24:MI:SS')
//	minus
//	select server_name from t_server_ud where time_stamp < to_date ('2018-04-08 00:00:00','YYYY-MM-DD HH24:MI:SS')
//)

//type TestItem struct {
//	Id    string
//	Name  string
//	Phone string
//}
//func StoreItem(db *sql.DB, items []TestItem) {
//	sql_additem := `
//    INSERT OR REPLACE INTO items(
//        Id,
//        Name,
//        Phone,
//        InsertedDatetime
//    ) values(?, ?, ?, CURRENT_TIMESTAMP)
//    `

//	stmt, err := db.Prepare(sql_additem)
//	if err != nil {
//		panic(err)
//	}
//	defer stmt.Close()

//	for _, item := range items {
//		_, err2 := stmt.Exec(item.Id, item.Name, item.Phone)
//		if err2 != nil {
//			panic(err2)
//		}
//	}
//}

//func ReadItem(db *sql.DB) []TestItem {
//	sql_readall := `
//    SELECT Id, Name, Phone FROM items
//    ORDER BY datetime(InsertedDatetime) DESC
//    `

//	rows, err := db.Query(sql_readall)
//	if err != nil {
//		panic(err)
//	}
//	defer rows.Close()

//	var result []TestItem
//	for rows.Next() {
//		item := TestItem{}
//		err2 := rows.Scan(&item.Id, &item.Name, &item.Phone)
//		if err2 != nil {
//			panic(err2)
//		}
//		result = append(result, item)
//	}
//	return result
//}
