package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func killPortProcess(targetPort int) {
	procDir := "/proc"
	targetProcess := "containers-rootlessport"

	files, err := os.ReadDir(procDir)
	if err != nil {
		fmt.Println("Error reading /proc directory:", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		pid := file.Name()
		cmdlinePath := filepath.Join(procDir, pid, "cmdline")
		cmdline, err := os.ReadFile(cmdlinePath)
		if err != nil {
			continue
		}

		if strings.Contains(string(cmdline), targetProcess) {
			netPath := filepath.Join(procDir, pid, "net", "tcp")
			netData, err := os.ReadFile(netPath)
			if err != nil {
				continue
			}

			if strings.Contains(string(netData), fmt.Sprintf(":%04X", targetPort)) {
				fmt.Printf("PID of '%s' listening on port %d: %s\n", targetProcess, targetPort, pid)
				return
			}
		}
	}

	fmt.Printf("No process named '%s' found listening on port %d.\n", targetProcess, targetPort)
}
