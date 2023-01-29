package process

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/cbodonnell/stack/pkg/logging"
)

type Process struct {
	Name     string
	path     string
	args     []string
	WorkDir  string
	stdout   io.Writer
	stderr   io.Writer
	cmd      *exec.Cmd
	ErrChan  chan error
	ExitChan chan struct{}
	stopping bool
}

func New(name string, path string, args []string, workDir string) *Process {
	return &Process{
		Name:     name,
		path:     path,
		args:     args,
		WorkDir:  workDir,
		stdout:   os.Stdout,
		stderr:   os.Stderr,
		cmd:      nil,
		ErrChan:  make(chan error),
		ExitChan: make(chan struct{}),
		stopping: false,
	}
}

func (f *Process) Start() error {
	log.Printf("starting %s", f.Name)

	if f.cmd != nil {
		return fmt.Errorf("%s already running", f.Name)
	}

	f.cmd = exec.Command(f.path, f.args...)
	f.cmd.Dir = f.WorkDir
	f.cmd.Env = os.Environ()
	f.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdout, _ := f.cmd.StdoutPipe()
	stderr, _ := f.cmd.StderrPipe()
	go logging.LogReaderWithPrefix(stdout, fmt.Sprintf("%s stdout: ", f.Name))
	go logging.LogReaderWithPrefix(stderr, fmt.Sprintf("%s stderr: ", f.Name))

	if err := f.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %s", f.Name, err)
	}

	log.Printf("%s started", f.Name)

	return nil
}

func (f *Process) Wait() error {
	if f.cmd == nil {
		return fmt.Errorf("%s not running", f.Name)
	}

	if err := f.cmd.Wait(); err != nil {
		if f.stopping {
			return nil
		}
		return fmt.Errorf("%s exited unexpectedly with error: %s", f.Name, err)
	}

	return fmt.Errorf("%s exited unexpectedly with no error", f.Name)
}

func (f *Process) StartAndWait() {
	go func() {
		if err := f.Start(); err != nil {
			f.ErrChan <- fmt.Errorf("failed to start %s: %s", f.Name, err)
		}

		if err := f.Wait(); err != nil {
			f.ErrChan <- fmt.Errorf("%s exited unexpectedly: %s", f.Name, err)
		}

		f.ExitChan <- struct{}{}
	}()
}

func (f *Process) Stop() error {
	log.Printf("stopping %s", f.Name)
	f.stopping = true
	defer func() {
		f.stopping = false
	}()

	if f.cmd == nil {
		return fmt.Errorf("%s not running", f.Name)
	}

	waitForExit := true
	if err := syscall.Kill(-f.cmd.Process.Pid, syscall.SIGINT); err != nil {
		if err.Error() != "no such process" {
			return fmt.Errorf("failed to send interrupt signal to %s: %s", f.Name, err)
		}
		log.Printf("%s already exited", f.Name)
		waitForExit = false
	}

	if waitForExit {
		select {
		case <-time.After(5 * time.Second):
			log.Printf("%s did not exit gracefully after 5 seconds, killing...", f.Name)
			syscall.Kill(-f.cmd.Process.Pid, syscall.SIGKILL)
			<-f.ExitChan
			log.Printf("%s killed", f.Name)
		case <-f.ExitChan:
			log.Printf("%s exited gracefully", f.Name)
		}
	}

	f.cmd = nil

	return nil
}

func (f *Process) Restart() error {
	log.Printf("restarting %s", f.Name)

	if err := f.Stop(); err != nil {
		return fmt.Errorf("failed to stop %s: %s", f.Name, err)
	}

	f.StartAndWait()

	return nil
}
