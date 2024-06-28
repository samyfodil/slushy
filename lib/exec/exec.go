package slushy

import (
	"bytes"
	"os/exec"
	"sync"
	"sync/atomic"

	"github.com/taubyte/tau/pkg/starlark"
)

type module struct {
	mu      sync.Mutex
	execs   map[int]*exec.Cmd
	buffers map[int]*commandBuffers
	nextID  int64
}

type commandBuffers struct {
	stdout bytes.Buffer
	stderr bytes.Buffer
}

func New() starlark.Module {
	return &module{
		execs:   make(map[int]*exec.Cmd),
		buffers: make(map[int]*commandBuffers),
	}
}

func (e *module) Name() string {
	return "exec"
}

func (e *module) generateUniqueID() int {
	return int(atomic.AddInt64(&e.nextID, 1))
}

func (e *module) E_New(command string) int {
	e.mu.Lock()
	defer e.mu.Unlock()

	cmd := exec.Command("sh", "-c", command)
	cmdID := e.generateUniqueID()
	e.execs[cmdID] = cmd
	e.buffers[cmdID] = &commandBuffers{}

	return cmdID
}

func (e *module) E_Output(cmdId int) string {
	e.mu.Lock()
	buffers, exists := e.buffers[cmdId]
	e.mu.Unlock()
	if !exists {
		return ""
	}

	return buffers.stdout.String()
}

func (e *module) E_ErrorOutput(cmdId int) string {
	e.mu.Lock()
	buffers, exists := e.buffers[cmdId]
	e.mu.Unlock()
	if !exists {
		return ""
	}

	return buffers.stderr.String()
}

func (e *module) E_ReturnCode(cmdId int) int {
	e.mu.Lock()
	cmd, exists := e.execs[cmdId]
	e.mu.Unlock()
	if !exists {
		return -1
	}

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return -1
	}

	return 0
}

func (e *module) E_Run(cmdId int) int {
	e.mu.Lock()
	cmd, exists := e.execs[cmdId]
	buffers, bufExists := e.buffers[cmdId]
	e.mu.Unlock()
	if !exists || !bufExists {
		return -1
	}

	cmd.Stdout = &buffers.stdout
	cmd.Stderr = &buffers.stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return -1
	}

	return 0
}

func (e *module) E_Start(cmdId int) {
	e.mu.Lock()
	cmd, exists := e.execs[cmdId]
	buffers, bufExists := e.buffers[cmdId]
	e.mu.Unlock()
	if !exists || !bufExists {
		return
	}

	cmd.Stdout = &buffers.stdout
	cmd.Stderr = &buffers.stderr

	cmd.Start()
}

func (e *module) E_Wait(cmdId int) int {
	e.mu.Lock()
	cmd, exists := e.execs[cmdId]
	e.mu.Unlock()
	if !exists {
		return -1
	}

	err := cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return -1
	}

	return 0
}

func (e *module) E_Kill(cmdId int) {
	e.mu.Lock()
	cmd, exists := e.execs[cmdId]
	e.mu.Unlock()
	if !exists {
		return
	}

	cmd.Process.Kill()
}
