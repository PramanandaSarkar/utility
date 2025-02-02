package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Task struct {
	Name     string
	Duration time.Duration
}

const historyFile = "timer_history.log"

var (
	taskQueue []Task
	queueMux  sync.Mutex
)

func parseDuration(input string) (time.Duration, error) {
	fs := flag.NewFlagSet("durationFlags", flag.ContinueOnError)
	h := fs.Int("h", 0, "Hours")
	m := fs.Int("m", 0, "Minutes")
	s := fs.Int("s", 0, "Seconds")

	args := strings.Fields(input)
	if err := fs.Parse(args); err != nil {
		return 0, err
	}

	if *h < 0 || *m < 0 || *s < 0 {
		return 0, fmt.Errorf("negative values not allowed")
	}

	return time.Duration(*h)*time.Hour +
		time.Duration(*m)*time.Minute +
		time.Duration(*s)*time.Second, nil
}

func startTimer(task string, duration time.Duration) {
	endTime := time.Now().Add(duration)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	fmt.Printf("\nStarting %s timer for %s\n", task, duration.Round(time.Second))

	for {
		select {
		case <-ticker.C:
			remaining := time.Until(endTime).Round(time.Second)
			if remaining <= 0 {
				fmt.Printf("\r%s: \033[32mCompleted!\033[0m\n", task)
				return
			}
			fmt.Printf("\r%s: %-10s remaining", task, remaining)
		}
	}
}

func logHistory(task Task) error {
	file, err := os.OpenFile(historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	entry := fmt.Sprintf("%s|%s|%s\n",
		task.Name,
		task.Duration.String(),
		time.Now().Format("2006-01-02 15:04:05"),
	)

	_, err = file.WriteString(entry)
	return err
}

func showHistory() error {
	file, err := os.Open(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No history available")
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	fmt.Println("\nTask History:")
	fmt.Println("----------------------------------------")
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "|")
		if len(parts) != 3 {
			continue
		}
		fmt.Printf("Task: %s\nDuration: %s\nCompleted: %s\n\n",
			parts[0], parts[1], parts[2])
	}
	return scanner.Err()
}

func handleInput(cmdCh chan<- string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cmdCh <- scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Input error: %v\n", err)
	}
	close(cmdCh)
}

func main() {
	historyFlag := flag.Bool("history", false, "Show timer history")
	flag.Parse()

	if *historyFlag {
		if err := showHistory(); err != nil {
			fmt.Printf("Error showing history: %v\n", err)
		}
		return
	}

	cmdCh := make(chan string)
	go handleInput(cmdCh)

	// fmt.Println("Timer App - Enter commands ('add', 'exit', or task duration)")
	// fmt.Println("Format: add <task name> [flags]")
	// fmt.Println("Example: add 'Study Session' -m 25 -s 30")
	fmt.Print("$")

	for {
		queueMux.Lock()
		hasTasks := len(taskQueue) > 0
		queueMux.Unlock()

		if hasTasks {
			queueMux.Lock()
			task := taskQueue[0]
			taskQueue = taskQueue[1:]
			queueMux.Unlock()

			done := make(chan struct{})
			go func() {
				startTimer(task.Name, task.Duration)
				close(done)
			}()

			if err := logHistory(task); err != nil {
				fmt.Printf("Error logging history: %v\n", err)
			}

			// Wait for timer completion or new commands
			for {
				select {
				case cmd, ok := <-cmdCh:
					if !ok {
						return
					}
					processCommand(cmd)
				case <-done:
					goto NextTask
				}
			}
		NextTask:
		} else {
			select {
			case cmd, ok := <-cmdCh:
				if !ok {
					return
				}
				processCommand(cmd)
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func processCommand(cmd string) {
	if strings.ToLower(cmd) == "exit" {
		fmt.Println("Exiting...")
		os.Exit(0)
	}

	if !strings.HasPrefix(cmd, "add ") {
		fmt.Println("Unknown command. Use 'add <task> [flags]' or 'exit'")
		return
	}

	args := strings.Fields(cmd)[1:]
	var flagsIndex int
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flagsIndex = i
			break
		}
	}

	if flagsIndex == 0 {
		fmt.Println("Invalid command format. Use: add <task name> [flags]")
		return
	}

	taskName := strings.Join(args[:flagsIndex], " ")
	durationStr := strings.Join(args[flagsIndex:], " ")

	duration, err := parseDuration(durationStr)
	if err != nil {
		fmt.Printf("Error parsing duration: %v\n", err)
		return
	}

	if duration <= 0 {
		fmt.Println("Duration must be positive")
		return
	}

	queueMux.Lock()
	taskQueue = append(taskQueue, Task{Name: taskName, Duration: duration})
	queueMux.Unlock()

	fmt.Printf("Added task: %s (%s)\n", taskName, duration.Round(time.Second))
}