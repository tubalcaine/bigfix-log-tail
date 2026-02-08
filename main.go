package main

// Package main provides a utility to tail and monitor log files in real-time.
// It uses the fsnotify package to watch for changes in the log files and the
// nxadm/tail package to read the log file contents.
import (
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/nxadm/tail"
)

// LogTailer manages the state for tailing log files
type LogTailer struct {
	microsecDelay int
	cancelFunc    context.CancelFunc
}

func main() {
	logpath := ""
	tailer := &LogTailer{}

	flag.IntVar(&tailer.microsecDelay, "usec", 150, "Microsecond delay to wait for next line (default 150)")
	flag.Parse()

	switch runtime.GOOS {
	case "darwin":
		logpath = "/Library/Application Support/BigFix/BES Agent/__BESData/__Global/Logs"
	case "linux":
		logpath = "/var/opt/BESClient/__BESData/__Global/Logs"
	case "windows":
		logpath = "C:\\Program Files (x86)\\BigFix Enterprise\\BES Client\\__BESData\\__Global\\Logs"
	default:
		logpath = "/var/opt/BESClient/__BESData/__Global/Logs"
	}

	args := flag.Args()
	if len(args) == 1 {
		logpath = args[0]
	} else if len(args) > 1 {
		log.Printf("Cannot specify more than one path to watch.")
		os.Exit(1)
	}

	_, err := os.Stat(logpath)

	if err != nil {
		log.Printf("Error opening directory [%s]: %v. Does it exist?\n", logpath, err)
		os.Exit(1)
	}

	log.Printf("Tailing current log in %s with %d microsecond delay", logpath, tailer.microsecDelay)

	// Create a root context that can be canceled on signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down gracefully...", sig)
		cancel() // Cancel the root context
	}()

	tailer.tailLatestFile(ctx, logpath)
}

func (lt *LogTailer) tailLatestFile(ctx context.Context, dir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}

	defer watcher.Close()

	// Track the currently tailed file
	var currentFile string

	// Get the latest file in directory and tail it initially
	currentFile = getLatestFile(dir)
	if currentFile != "" {
		// Cancel previous tail if exists
		if lt.cancelFunc != nil {
			lt.cancelFunc()
		}
		childCtx, cancel := context.WithCancel(ctx)
		lt.cancelFunc = cancel
		go tailFile(currentFile, childCtx)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				// Parent context canceled (signal received)
				log.Printf("Context canceled, stopping file watcher")
				if lt.cancelFunc != nil {
					lt.cancelFunc()
				}
				return
			case event := <-watcher.Events:
				// Check for any event that might indicate a new file or modified file
				if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Chmod) != 0 {
					// Get the latest file after this event
					latestFile := getLatestFile(dir)

					// Only switch if we found a file and it's different from the current one
					if latestFile != "" && latestFile != currentFile {
						log.Printf("Detected newer log file: %s", latestFile)

						// Cancel previous tail
						if lt.cancelFunc != nil {
							lt.cancelFunc()
						}

						childCtx, cancel := context.WithCancel(ctx)
						lt.cancelFunc = cancel
						currentFile = latestFile
						go tailFile(currentFile, childCtx)
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

func tailFile(file string, ctx context.Context) {
	log.Printf("Now tailing %s", file)
	// Print the last 10 lines before tailing
	printLastNLines(file, 10)
	seekInfo := tail.SeekInfo{Offset: 0, Whence: io.SeekEnd}
	isWin := os.PathListSeparator != '/'

	t, err := tail.TailFile(file, tail.Config{Poll: isWin, Follow: true, Location: &seekInfo})
	if err != nil {
		log.Fatal(err)
	}

	// Make sure we properly clean up when this function exits
	defer func() {
		t.Stop()
		t.Cleanup()
		log.Printf("Stopped tailing %s", file)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-t.Lines:
			if !ok {
				// Channel closed
				return
			}
			if line.Err != nil {
				log.Printf("Error reading line: %v", line.Err)
				continue
			}
			log.Printf("%s", line.Text)
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
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Scanner error: %v", err)
	}
}
