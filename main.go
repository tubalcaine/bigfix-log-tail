package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/hpcloud/tail"
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

	retval := path + string(os.PathSeparator) + mostRecentName
	return retval
}

func MostRecentFileLines(path string, msgout chan *string) {
	fmt.Println("Entering MostRecentFileLines")

	finm := mostRecentFilename(path)

	tailfile, err := tail.TailFile(finm, tail.Config{Follow: true})

	if err != nil {
		panic(err)
	}

	tf := tailfile.Lines

	for {
		select {
		case msg := <-tf:
			ltext := msg.Text
			msgout <- &ltext

		default:
			time.Sleep(100 * time.Microsecond)
			newname := mostRecentFilename(path)

			if newname != finm {
				fmt.Println("SWITCHING FROM", finm, "TO", newname)
				finm = newname

				tailfile, err = tail.TailFile(finm, tail.Config{Follow: true})

				if err != nil {
					panic(err)
				}

				tf = tailfile.Lines
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
