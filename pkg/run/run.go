package run

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/cbodonnell/stack/pkg/config"
	"github.com/cbodonnell/stack/pkg/process"
	"github.com/fsnotify/fsnotify"
)

func Run(cfg *config.EnvironmentConfig) error {
	log.Printf("running %s", cfg.Name)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	exitChan := make(chan error)

	procs := make(map[string]*process.Process)
	for _, p := range cfg.Proccesses {
		proc := process.New(p.Name, p.Command, p.Args, p.WorkDir)
		procs[p.Name] = proc
		go func(p config.ProcessConfig) {
			proc.StartAndWait()
			for {
				err := func(proc *process.Process) error {
					watcher, err := fsnotify.NewWatcher()
					if err != nil {
						return fmt.Errorf("failed to create watcher: %s", err)
					}
					defer watcher.Close()

					// Add the directory you want to watch for changes
					err = watcher.Add(proc.WorkDir)
					if err != nil {
						return fmt.Errorf("failed to add watcher: %s", err)
					}

					log.Printf("watching %s", proc.Name)

					// Watch for events
					select {
					case <-watcher.Events:
						log.Printf("detected change in %s", proc.Name)
					case err := <-watcher.Errors:
						return fmt.Errorf("failed to watch %s: %s", proc.Name, err)
					}

					if err := proc.Restart(); err != nil {
						return fmt.Errorf("failed to restart %s: %s", proc.Name, err)
					}

					return nil
				}(proc)
				if err != nil {
					exitChan <- fmt.Errorf("failed to run %s: %s", proc.Name, err)
					return
				}
				time.Sleep(1 * time.Second)
			}
		}(p)
	}

	select {
	case <-interrupt:
		log.Println("interrupted, stopping processes")
		done := make(chan struct{})
		go func() {
			for _, proc := range procs {
				proc.Stop()
			}
			close(done)
		}()
		select {
		case <-done:
			return errors.New("interrupted")
		case <-time.After(5 * time.Second):
			return errors.New("interrupted, timed out")
		}
	case err := <-exitChan:
		return fmt.Errorf("error: %s", err)
	}
}
