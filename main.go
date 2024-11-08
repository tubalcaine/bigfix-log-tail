package main

// Package main provides a utility to tail and monitor log files in real-time.
// It uses the fsnotify package to watch for changes in the log files and the
// nxadm/tail package to read the log file contents.
import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nxadm/tail"
)

// A package level reference to the current
// tail goroutine in order to close it on starting next one
var g_curTailChan chan bool = nil
var g_usecDelay int = 150

func main() {
	logpath := ""

	flag.IntVar(&g_usecDelay, "usec", 150, "Microsecond delay to wait for next line (default 150)")
	flag.Parse()

	if os.PathSeparator == '/' {
		// This doesn't handle MacOS at all!
		logpath = "/var/opt/BESClient/__BESData/__Global/Logs"
	} else {
		logpath = "C:\\Program Files (x86)\\BigFix Enterprise\\BES Client\\__BESData\\__Global\\Logs"
	}

	args := flag.Args()
	if len(args) == 1 {
		logpath = args[0]
	} else if len(args) > 1 {
		log.Printf("Cannot specify more than one path to watch.")
		os.Exit(1)
		os.Exit(2)
	}

	_, err := os.Stat(logpath)

	if err != nil {
		log.Printf("Error opening directory [%s]: %v. Does it exist?\n", logpath, err)
		os.Exit(1)
	}

	log.Printf("Tailing current log in %s with %d microsecond delay", logpath, g_usecDelay)
	fmt.Println("Tailing current log in", logpath, " with ", g_usecDelay, " microsecond delay")
	tailLatestFile(logpath)
}

func tailLatestFile(dir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}

	defer watcher.Close()

	// Get the latest file in directory and tail it initially
	file := getLatestFile(dir)
	if file != "" {
		if g_curTailChan != nil {
			g_curTailChan <- true
		}
		terminate := make(chan bool)
		g_curTailChan = terminate
		go tailFile(file, terminate)
	}

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				// Check if event was a create event
				if event.Op&fsnotify.Create == fsnotify.Create {
					file := event.Name
					if file != "" {
						if g_curTailChan != nil {
							g_curTailChan <- true
						}
						terminate := make(chan bool)
						g_curTailChan = terminate
						go tailFile(file, terminate)
					}
				}
			case err := <-watcher.Errors:
				log.Printf("Watcher error: %v", err)
			}
		}
	}()

	err = watcher.Add(dir)
	if err != nil {
		log.Printf("Error adding watcher to directory: %v", err)
		os.Exit(1)
	}
	<-done
}

func getLatestFile(dir string) string {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	sort.Slice(files, func(i, j int) bool {
		infoI, err := files[i].Info()
		if err != nil {
			log.Fatal(err)
		}
		infoJ, err := files[j].Info()
		if err != nil {
			log.Fatal(err)
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	for _, file := range files {
		if !file.IsDir() {
			return filepath.Join(dir, file.Name())
		}
	}

	return ""
}

func tailFile(file string, terminate chan bool) {
	fmt.Printf("\nNow tailing %s\n", file)
	// Print the last 10 lines before tailing
	printLastNLines(file, 10)
	seekInfo := tail.SeekInfo{Offset: 0, Whence: io.SeekEnd}
	isWin := os.PathListSeparator != '/'

	t, err := tail.TailFile(file, tail.Config{Poll: isWin, Follow: true, Location: &seekInfo})
	if err != nil {
		log.Fatal(err)
	}

tLoop:
	for {
		select {
		case term := <-terminate:
			if term {
				t.Cleanup()
				break tLoop
			}
		case line := <-t.Lines:
			if line.Err != nil {
				log.Printf("Error reading line: %v", line.Err)
				continue
			}
			log.Printf("%s", line.Text)
			fmt.Println(line.Text)
		default:
			// This is just here to keep this from eating too much CPU
			time.Sleep(75 * time.Microsecond)
		}
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
	for _, line := range lines {
		log.Printf("%s", line)
		fmt.Println(line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Scanner error: %v", err)
	}
}
