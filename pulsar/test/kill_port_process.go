package test

import (
	"fmt"
	"os"
	"syscall"

	"github.com/cakturk/go-netstat/netstat"
)

func killPortProcess(targetPort int) error {
	socks6, err := netstat.TCP6Socks(netstat.NoopFilter)
	if err != nil {
		return err
	}
	socks, err := netstat.TCPSocks(netstat.NoopFilter)
	if err != nil {
		return err
	}
	for _, sock := range append(socks6, socks...) {
		if sock.LocalAddr.Port == uint16(targetPort) {
			pid := sock.Process.Pid
			process, err := os.FindProcess(pid)
			if err != nil {
				return err
			}
			fmt.Println("Killing process of port", pid, targetPort)

			// Send a SIGTERM signal to the process
			err = process.Signal(syscall.SIGTERM)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}
