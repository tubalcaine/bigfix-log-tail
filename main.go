package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
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

func mostRecentFilename(path string) string {
	ls, err := os.ReadDir(path)

	if err != nil {
		panic(err)
	}

	mostRecentName := ""
	mostRecentTimestamp := int64(0)

	for _, file := range ls {
		info, err := file.Info()

		if err != nil {
			panic(err)
		}

		if info.ModTime().Unix() > mostRecentTimestamp {
			mostRecentName = file.Name()
			mostRecentTimestamp = info.ModTime().Unix()
		}
	}

	return path + string(os.PathSeparator) + mostRecentName
}

func MostRecentFileLines(path string, msgout chan *string) {
	fmt.Println("Entering MostRecentFileLines")

	finm := mostRecentFilename(path)

	f, err := os.Open(finm)

	if err != nil {
		panic(err)
	}

	sc := bufio.NewScanner(f)

	for {
		if sc.Scan() {
			// There is data; scan it and send it to the channel
			msg := sc.Text()
			msgout <- &msg
		} else {
			// There isn't data available; see if we have a new file
			newname := mostRecentFilename(path)
			if newname != finm {
				fmt.Println("SWITCHING FROM", finm, "TO", newname)
				finm = newname
				f, err = os.Open(finm)

				if err != nil {
					panic(err)
				}

				sc = bufio.NewScanner(f)
			}
		}
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

	args := os.Args
	if len(args) == 2 {
		logpath = args[1]
	}

	fmt.Println(logpath)

	msg_ch := make(chan *string)

	go PrintStdout(msg_ch)
	go MostRecentFileLines(logpath, msg_ch)

	// Wait for an interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	s := <-c

	// "Close" the writer
	msg_ch <- nil

	time.Sleep(1 * time.Second)

	fmt.Println("Main exits with signal ", s)
}
