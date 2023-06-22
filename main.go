package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/fsnotify/fsnotify"
	"github.com/nxadm/tail"
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

func getLatestFile(dir string) string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	for _, file := range files {
		if !file.IsDir() {
			return filepath.Join(dir, file.Name())
		}
	}

	return ""
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

	scanner := bufio.NewScanner(f)

	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}

	if scanner.Err() != nil {
		log.Fatal(scanner.Err())
	}

	for _, line := range lines {
		fmt.Println(line)
	}
}
