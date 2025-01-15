package processes

import (
	"fmt"
	"sort"

	"github.com/shirou/gopsutil/process"
)

const (
	maxProcessCnt = 10
	threadProcess = 5
)

var errCreateDate = "failed to get createdate from process: %s"

func filterProcesses(prcs []*process.Process) ([]*process.Process, error) {
	r := make([]*process.Process, 0)
	for _, p := range prcs {
		n, err := p.Name()
		if err != nil {
			return nil, err
		}

		if n == "chat" {
			r = append(r, p)
		}
	}
	return r, nil
}

func StopJoinRoom() error {
	prcs, err := process.Processes()
	if err != nil {
		return fmt.Errorf("failed to see all processes: %s", err)
	}

	processes, err := filterProcesses(prcs)
	if err != nil {
		return fmt.Errorf("failed to get process name: %s", err)
	}

	for _, p := range processes {
		if err := p.Kill(); err != nil {
			return fmt.Errorf("failed to kil process: %s", err)
		}
	}
	return nil
}

func SheduleProcessCount() chan error {
	errCh := make(chan error)
	prcs, err := process.Processes()
	if err != nil {
		errCh <- fmt.Errorf("failed to see all processes: %s", err)
		return errCh
	}

	processes, err := filterProcesses(prcs)
	if err != nil {
		errCh <- fmt.Errorf("failed to get process name: %s", err)
		return errCh
	}

	sort.Slice(processes, func(i, j int) bool {
		t1, err := processes[i].CreateTime()
		if err != nil {
			errCh <- fmt.Errorf(errCreateDate, err)
		}

		t2, err := processes[j].CreateTime()
		if err != nil {
			errCh <- fmt.Errorf(errCreateDate, err)
		}

		return t1 < t2
	})

	for i := 0; i < len(processes)-maxProcessCnt; i++ {
		if err = processes[i].Kill(); err != nil {
			errCh <- fmt.Errorf("failed to kill unnecessary processes: %s", err)
		}
	}
	return errCh
}
