package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func PrintStdout(msgin chan *string) {
	fmt.Println("Entering PrintStdout")
	msg := <-msgin

	for msg != nil {
		fmt.Println(*msg)
		msg = <-msgin
		if msg == nil {
			break
		}
	}

	fmt.Println("Exiting PrintStdOut")
}

func MostRecentFileLines(path string, msgout chan *string) {
	fmt.Println("Entering MostRecentFileLines")
	for i := 0; i < 30; i++ {
		msg := ("Message line " + strconv.Itoa(i))
		msgout <- &msg
	}
}

func main() {
	logpath := ""

	if os.PathSeparator == '/' {
		// This doesn't handle MacOS at all!
		logpath = "/var/opt/BESClient/__BESData/__Global/Logs"
	} else {
		logpath = "C:\\Program Files (x86)\\BigFix Enterprise\\BES Client\\__BESData\\__Global\\Logs"
	}

	fmt.Println(logpath)

	msg_ch := make(chan *string)

	go PrintStdout(msg_ch)
	go MostRecentFileLines("NothingYet", msg_ch)

	// Wait for an interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	s := <-c

	// "Close" the writer
	msg_ch <- nil

	time.Sleep(1 * time.Second)

	fmt.Println("Main exits with signal ", s)
}
