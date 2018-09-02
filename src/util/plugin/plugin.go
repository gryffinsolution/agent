package plugin

import (
	"bytes"
	"database/sql"
	"log"
	"os/exec"
	"time"
	"util/dao"
)

func Test() {
	log.Println("sex")
}

func RunPlugin(db *sql.DB, plugin string, cmdStr string, timeoutInt int) bool {
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
			dao.InsertData(db, plugin, outstr)
			return true
		}
	}
	return true
}

func RunPluginTest(db *sql.DB, plugin string, cmdStr string, timeoutInt int) bool {
	log.Println(plugin, "'s cmd=", cmdStr)
	output, _ := exec.Command(cmdStr).CombinedOutput()
	log.Println(plugin, "'s result=", output)
	//cmd.Process.Signal(os.Kill)
	return true

	//	stdout, _ := cmd.StdoutPipe()
	//	stderr, _ := cmd.StderrPipe()
	//	err := cmd.Start()
	//	if err != nil {
	//		log.Println("Non-zero exit code:", err)
	//		dao.InsertEvent(db, "PE002", "ERROR", plugin+"("+stderr+") execution error")
	//		return false
	//	} else {
	//		dao.InsertData(db, plugin, stdout)
	//		return true
	//	}
	//.Process.Kill()
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
