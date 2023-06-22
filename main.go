package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/hpcloud/tail"
	"github.com/schollz/seeker"
)

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

	info, err := os.Stat(logpath)

	if err != nil {
		fmt.Printf("Error opening directory [%s]. Does it exist?", logpath)
	}

	fmt.Println("Tailing current log in", logpath)
	tailLatestFile(logpath)
}

func tailLatestFile(dir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					file := getLatestFile(dir)
					if file != "" {
						go tailFile(file)
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func getLatestFile(dir string) string {
	// existing logic...
}

func tailFile(file string) {
	// Print the last 10 lines before tailing
	printLastNLines(file, 10)

	t, err := tail.TailFile(file, tail.Config{Follow: true})
	if err != nil {
		log.Fatal(err)
	}

	for line := range t.Lines {
		fmt.Println(line.Text)
	}
}

func printLastNLines(file string, n int) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	s := seeker.New(f)
	r := bufio.NewReader(s)

	var lines []string

	// Seek to the end of the file
	_, err = s.Seek(0, io.SeekEnd)
	if err != nil {
		log.Fatal(err)
	}

	for {
		line, err := r.ReadString('\n')
		if err != nil && line == "" {
			break
		}
		line = strings.TrimRight(line, "\n")
		lines = append([]string{line}, lines...)

		if len(lines) > n {
			lines = lines[:n]
		}

		_, err = s.Seek(-2*int64(len(line)+1), io.SeekCurrent)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, line := range lines {
		fmt.Println(line)
	}
}
