package test

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

const windows = "windows"

func killPortProcess(targetPort int) error {
	processes, err := getProcessesForPort(targetPort)
	if err != nil {
		return err
	}

	for _, pid := range processes {
		fmt.Printf("Killing process on port %d with PID %d\n", targetPort, pid)

		switch runtime.GOOS {
		case windows:
			killCmd := exec.Command("taskkill", "/F", "/PID", fmt.Sprint(pid))
			if err := killCmd.Run(); err != nil {
				return err
			}
		default:
			process, err := os.FindProcess(pid)
			if err != nil {
				return err
			}
			err = process.Signal(syscall.SIGTERM)
		}
	}
	return nil
}

func getProcessesForPort(targetPort int) ([]int, error) {
	var cmd *exec.Cmd
	processes := make([]int, 0)
	switch runtime.GOOS {
	case windows:
		cmd = exec.Command("cmd", "/c", "netstat", "-ano", "|", "findstr", fmt.Sprintf(":%d", targetPort))
	default:
		cmd = exec.Command("sh", "-c", fmt.Sprintf("lsof -i tcp:%d | awk 'NR!=1 {print $2}'", targetPort))
	}
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		pidStr := scanner.Text()
		pidStr = strings.TrimSpace(strings.Split(pidStr, " ")[0]) // Extracting PID from the output

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue // Not a valid number, skip
		}
		processes = append(processes, pid)
	}

	return processes, scanner.Err()
}
