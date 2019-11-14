package pilot

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const (
	FLUENTDBIT_BASE      = "/fluent-bit"
	FLUENTDBIT_EXEC_CMD  = FLUENTDBIT_BASE + "/bin/fluent-bit"
	FLUENTDBIT_CONF_BASE = FLUENTDBIT_BASE + "/etc"
	FLUENTDBIT_CONF_DIR  = FLUENTDBIT_CONF_BASE + "/conf"
	FLUENTDBIT_CONF_FILE = FLUENTDBIT_CONF_BASE + "/fluent-bit.conf"
)

var fluentBit *exec.Cmd

// FluentdBitPiloter for fluentd-bit plugin
type FluentdBitPiloter struct {
	name string
}

// NewFluentdPiloter returns a FluentdPiloter instance
func NewFluentdBitPiloter() (Piloter, error) {
	return &FluentdPiloter{
		name: PILOT_FLUENT_BIT,
	}, nil
}

// Start starting and watching fluentd process
func (p *FluentdBitPiloter) Start() error {
	if fluentBit != nil {
		pid := fluentd.Process.Pid
		log.Infof("fluent-bit started, pid: %v", pid)
		return fmt.Errorf(ERR_ALREADY_STARTED)
	}

	log.Info("starting fluent-bit")

	fluentBit = exec.Command(FLUENTDBIT_EXEC_CMD,
		"-c", FLUENTDBIT_CONF_FILE)
	fluentBit.Stderr = os.Stderr
	fluentBit.Stdout = os.Stdout
	err := fluentd.Start()
	if err != nil {
		log.Errorf("fluent-bit start fail: %v", err)
	}

	go func() {
		err := fluentBit.Wait()
		if err != nil {
			log.Errorf("fluentd bit exited: %v", err)
			if exitError, ok := err.(*exec.ExitError); ok {
				processState := exitError.ProcessState
				log.Errorf("fluent-bit exited pid: %v", processState.Pid())
			}
		}

		// try to restart fluentd
		log.Warningf("fluent-bit exited and try to restart")
		fluentd = nil
		p.Start()
	}()
	return err
}

// Stop log collection
func (p *FluentdBitPiloter) Stop() error {
	return nil
}

// Reload reload configuration file
func (p *FluentdBitPiloter) Reload() error {
	if fluentBit == nil {
		err := fmt.Errorf("fluent-bit have not started")
		log.Error(err)
		return err
	}

	log.Info("reload fluent-bit")
	ch := make(chan struct{})
	go func(pid int) {
		command := fmt.Sprintf("pgrep -P %d", pid)
		childId := shell(command)
		log.Infof("before reload childId : %s", childId)
		fluentBit.Process.Signal(syscall.SIGHUP)
		time.Sleep(5 * time.Second)
		afterChildId := shell(command)
		log.Infof("after reload childId : %s", afterChildId)
		if childId == afterChildId {
			log.Infof("kill childId : %s", childId)
			shell("kill -9 " + childId)
		}
		close(ch)
	}(fluentBit.Process.Pid)
	<-ch
	return nil
}

// GetConfPath returns log configuration path
func (p *FluentdBitPiloter) GetConfPath(container string) string {
	return fmt.Sprintf("%s/%s.conf", FLUENTDBIT_CONF_DIR, container)
}

// GetConfHome returns configuration directory
func (p *FluentdBitPiloter) GetConfHome() string {
	return FLUENTDBIT_CONF_DIR
}

// Name returns plugin name
func (p *FluentdBitPiloter) Name() string {
	return p.name
}

// OnDestroyEvent watching destroy event
func (p *FluentdBitPiloter) OnDestroyEvent(container string) error {
	log.Info("refactor in the future!!!")
	return nil
}

// GetBaseConf returns plugin root directory
func (p *FluentdBitPiloter) GetBaseConf() string {
	return FLUENTDBIT_CONF_BASE
}
