package tui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
)

const (
	defaultTermCols = 80
	defaultTermRows = 24
	ptyReadBufSize  = 4096
	ptyTermEnvVar   = "TERM=xterm-256color"
	scrollStep      = 3
)

type PTYSessionConfig struct {
	ThreadID string
	Command  string
	Args     []string
	SendMsg  func(PTYOutputMsg)
}

type PTYSession struct {
	threadID     string
	command      string
	args         []string
	cmd          *exec.Cmd
	ptmx         *os.File
	emulator     *vt.SafeEmulator
	mu           sync.Mutex
	running      bool
	exitCode     int
	scrollOffset int
	sendMsg      func(PTYOutputMsg)
}

func NewPTYSession(config PTYSessionConfig) *PTYSession {
	return &PTYSession{
		threadID: config.ThreadID,
		command:  config.Command,
		args:     config.Args,
		emulator: vt.NewSafeEmulator(defaultTermCols, defaultTermRows),
		sendMsg:  config.SendMsg,
	}
}

func (ps *PTYSession) Start() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.running {
		return fmt.Errorf("pty session already running")
	}

	ps.cmd = exec.Command(ps.command, ps.args...)
	ps.cmd.Env = append(os.Environ(), ptyTermEnvVar)

	size := &pty.Winsize{
		Rows: uint16(ps.emulator.Height()),
		Cols: uint16(ps.emulator.Width()),
	}

	ptmx, err := pty.StartWithSize(ps.cmd, size)
	if err != nil {
		return fmt.Errorf("start pty: %w", err)
	}

	ps.ptmx = ptmx
	ps.running = true

	go ps.readLoop()
	go ps.responseLoop()

	return nil
}

func (ps *PTYSession) readLoop() {
	buf := make([]byte, ptyReadBufSize)
	for {
		bytesRead, err := ps.ptmx.Read(buf)
		if bytesRead > 0 {
			ps.emulator.Write(buf[:bytesRead])
			ps.sendMsg(PTYOutputMsg{ThreadID: ps.threadID})
		}
		if err != nil {
			break
		}
	}

	ps.mu.Lock()
	ps.running = false
	if ps.cmd.ProcessState != nil {
		ps.exitCode = ps.cmd.ProcessState.ExitCode()
	}
	ps.mu.Unlock()

	ps.sendMsg(PTYOutputMsg{ThreadID: ps.threadID, Exited: true})
}

func (ps *PTYSession) responseLoop() {
	buf := make([]byte, ptyReadBufSize)
	for {
		bytesRead, err := ps.emulator.Read(buf)
		if bytesRead > 0 {
			ps.writeResponseToPTY(buf[:bytesRead])
		}
		if err != nil {
			return
		}
	}
}

func (ps *PTYSession) writeResponseToPTY(data []byte) {
	ps.mu.Lock()
	ptmx := ps.ptmx
	ps.mu.Unlock()

	if ptmx == nil {
		return
	}
	ptmx.Write(data)
}

func (ps *PTYSession) WriteBytes(data []byte) error {
	ps.mu.Lock()
	isRunning := ps.running
	ps.mu.Unlock()

	if !isRunning {
		return fmt.Errorf("pty session not running")
	}

	_, err := ps.ptmx.Write(data)
	if err != nil {
		return fmt.Errorf("write to pty: %w", err)
	}
	return nil
}

func (ps *PTYSession) Resize(width int, height int) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.emulator.Resize(width, height)

	if ps.ptmx == nil {
		return
	}

	size := &pty.Winsize{
		Rows: uint16(height),
		Cols: uint16(width),
	}
	_ = pty.Setsize(ps.ptmx, size)
}

func (ps *PTYSession) Render() string {
	return ps.emulator.Render()
}

func (ps *PTYSession) Stop() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.ptmx != nil {
		ps.ptmx.Close()
	}

	hasProcess := ps.cmd != nil && ps.cmd.Process != nil
	if hasProcess {
		ps.cmd.Process.Kill()
		ps.cmd.Wait()
	}

	ps.running = false
}

func (ps *PTYSession) Running() bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.running
}

func (ps *PTYSession) ExitCode() int {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.exitCode
}

func (ps *PTYSession) ThreadID() string {
	return ps.threadID
}

func (ps *PTYSession) ScrollUp(lines int) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	maxOffset := ps.emulator.ScrollbackLen()
	ps.scrollOffset += lines
	if ps.scrollOffset > maxOffset {
		ps.scrollOffset = maxOffset
	}
}

func (ps *PTYSession) ScrollDown(lines int) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.scrollOffset -= lines
	if ps.scrollOffset < 0 {
		ps.scrollOffset = 0
	}
}

func (ps *PTYSession) ScrollToBottom() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.scrollOffset = 0
}

func (ps *PTYSession) ScrollOffset() int {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	return ps.scrollOffset
}

func (ps *PTYSession) IsScrolledUp() bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	return ps.scrollOffset > 0
}

var _ io.Writer = (*PTYSession)(nil)

func (ps *PTYSession) Write(data []byte) (int, error) {
	return len(data), ps.WriteBytes(data)
}
