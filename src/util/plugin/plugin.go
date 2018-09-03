package plugin

import (
	"bytes"
	"database/sql"
	"log"
	"os/exec"
	"strings"
	"time"
	"util/dao"
)

func RunPlugin(db *sql.DB, plugin string, cmdStr string, timeoutInt int, isAppend bool) bool {
	log.Println(plugin, "'s cmd=", cmdStr)
	cmd := exec.Command(cmdStr)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Start()
	// Use a channel to signal completion so we can use a select statement
	done := make(chan error)
	go func() { done <- cmd.Wait() }()
	// Start a timer
	timeout := time.After(time.Duration(timeoutInt) * time.Second)
	// The select statement allows us to execute based on which channel
	// we get a message from first.
	select {
	case <-timeout:
		// Timeout happened first, kill the process and print a message.
		cmd.Process.Kill()
		log.Println("Command", cmdStr, "timed out")
		dao.InsertEvent(db, "PE001", "ERROR", "cmd("+cmdStr+") timeout")
		return false
	case err := <-done:
		// Command completed before timeout. Print output and error if it exists.
		outstr := buf.String()
		log.Println("Output:", outstr)
		if err != nil {
			log.Println("Non-zero exit code:", err)
			//TODO: fail event generation
			//Excution error : like no such file
			dao.InsertEvent(db, "PE002", "ERROR", plugin+"("+cmdStr+") execution error")
			return false
		} else {
			if isAppend == true {
				dao.InsertData(db, plugin, outstr)
			} else {
				dao.UpdateData(db, plugin, outstr)
			}
			return true
		}
	}
	return true
}

func RunCmd(db *sql.DB, cmdStr string, timeoutInt int) (string, string, int) {
	epStart := time.Now().Unix()
	log.Println("RunCmd=", cmdStr)

	parts := strings.Fields(cmdStr)
	head := parts[0]
	parts = parts[1:len(parts)]
	cmd := exec.Command(head, parts...)

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer

	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	cmd.Start()
	// Use a channel to signal completion so we can use a select statement
	done := make(chan error)
	go func() { done <- cmd.Wait() }()
	// Start a timer
	timeout := time.After(time.Duration(timeoutInt) * time.Second)
	// The select statement allows us to execute based on which channel
	// we get a message from first.
	select {
	case <-timeout:
		// Timeout happened first, kill the process and print a message.
		cmd.Process.Kill()
		log.Println("Command", cmdStr, "timed out")
		dao.InsertEvent(db, "RC001", "ERROR", "cmd("+cmdStr+") timeout")
		return "timeout", "ERR_TIME", timeoutInt
	case err := <-done:
		// Command completed before timeout. Print output and error if it exists.
		outstr := outBuf.String()
		errstr := errBuf.String()

		log.Println("STDOUT:", outstr)
		log.Println("STDERR:", errstr)
		epEnd := time.Now().Unix()
		elapsedTimeSec := epEnd - epStart
		if err != nil {
			log.Println("Non-zero exit code:", err)

			//return outstr, "ERR_EXEC", int(elapsedTimeSec)
			return errstr, "ERR_EXEC", int(elapsedTimeSec)
		} else {
			return outstr, "DONE", int(elapsedTimeSec)
		}
	}
	return "Abnormal.Check you agent log", "ERR_EXEC", 0
}

//func main() {
//	//https://medium.com/@vCabbage/go-timeout-commands-with-os-exec-commandcontext-ba0c861ed738
//	// We'll use ping as it will provide output and we can control how long it runs.
//	cmd := exec.Command("ping", "-c 2", "-i 1", "8.8.8.8")

//	// Use a bytes.Buffer to get the output
//	var buf bytes.Buffer
//	cmd.Stdout = &buf

//	cmd.Start()

//	// Use a channel to signal completion so we can use a select statement
//	done := make(chan error)
//	go func() { done <- cmd.Wait() }()

//	// Start a timer
//	timeout := time.After(1 * time.Second)

//	// The select statement allows us to execute based on which channel
//	// we get a message from first.
//	select {
//	case <-timeout:
//		// Timeout happened first, kill the process and print a message.
//		cmd.Process.Kill()
//		fmt.Println("Command timed out")
//	case err := <-done:
//		// Command completed before timeout. Print output and error if it exists.
//		fmt.Println("Output:", buf.String())
//		if err != nil {
//			fmt.Println("Non-zero exit code:", err)
//		}
//	}
//}
