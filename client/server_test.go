package client

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

const (
	startTimeout = 10 * time.Second
	stopTimeout  = 10 * time.Second
)

// testServer runs the MintDb server binary as a child process so tests can
// start it, stop it gracefully, or kill it to simulate a crash.
type testServer struct {
	bin     string // path to the built server binary
	dataDir string // holds mint.aof and server.log
	addr    string // host:port the server listens on

	cmd    *exec.Cmd
	exited chan struct{} // closed once the process has exited
}

// Start launches the server and waits until it accepts connections.
func (s *testServer) Start() error {
	if s.Running() {
		return errors.New("server already running")
	}
	logFile, err := os.OpenFile(filepath.Join(s.dataDir, "server.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	s.cmd = exec.Command(s.bin, "-addr", s.addr, "-data-dir", s.dataDir, "-repl=false")
	s.cmd.Stdout = logFile
	s.cmd.Stderr = logFile
	if err := s.cmd.Start(); err != nil {
		logFile.Close()
		return err
	}
	s.exited = make(chan struct{})
	go func() {
		s.cmd.Wait()
		logFile.Close()
		close(s.exited)
	}()

	deadline := time.Now().Add(startTimeout)
	for time.Now().Before(deadline) {
		if !s.Running() {
			return fmt.Errorf("server exited during startup (%v), see server.log", s.cmd.ProcessState)
		}
		if conn, err := net.DialTimeout("tcp", s.addr, 100*time.Millisecond); err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	s.Kill()
	return fmt.Errorf("server not listening on %s after %v", s.addr, startTimeout)
}

// Stop shuts the server down gracefully with SIGTERM, falling back to SIGKILL.
func (s *testServer) Stop() error {
	if !s.Running() {
		return nil
	}
	if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	select {
	case <-s.exited:
		return nil
	case <-time.After(stopTimeout):
		s.Kill()
		return fmt.Errorf("server ignored SIGTERM for %v, killed it", stopTimeout)
	}
}

// Kill sends SIGKILL, so the server gets no chance to clean up.
func (s *testServer) Kill() error {
	if !s.Running() {
		return nil
	}
	if err := s.cmd.Process.Kill(); err != nil {
		return err
	}
	<-s.exited
	return nil
}

func (s *testServer) Running() bool {
	if s.exited == nil {
		return false
	}
	select {
	case <-s.exited:
		return false
	default:
		return true
	}
}

// freeAddr returns a localhost address with a port nothing is listening on.
func freeAddr() (string, error) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer lis.Close()
	return lis.Addr().String(), nil
}
