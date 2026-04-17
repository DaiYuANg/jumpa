package sshclient

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func (c *Client) prepareInteractiveSession(session *ssh.Session, requested *Terminal) (func(), error) {
	if requested != nil {
		return c.prepareConfiguredTerminal(session, requested)
	}
	return c.prepareLocalTerminal(session)
}

func (c *Client) prepareConfiguredTerminal(session *ssh.Session, requested *Terminal) (func(), error) {
	terminal := normalizeTerminal(requested)
	if terminal == nil {
		return nil, nil
	}

	if err := session.RequestPty(terminal.Term, terminal.Height, terminal.Width, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		return nil, err
	}

	var restoreRaw func()
	if terminal.MakeRaw {
		file, ok := c.stdin.(interface{ Fd() uintptr })
		if ok {
			fd := int(file.Fd())
			if term.IsTerminal(fd) {
				state, err := term.MakeRaw(fd)
				if err != nil {
					return nil, err
				}
				restoreRaw = func() {
					_ = term.Restore(fd, state)
				}
			}
		}
	}

	stopWindowSync := forwardWindowSizes(session, terminal.Resize)
	return func() {
		if stopWindowSync != nil {
			stopWindowSync()
		}
		if restoreRaw != nil {
			restoreRaw()
		}
	}, nil
}

func (c *Client) prepareLocalTerminal(session *ssh.Session) (func(), error) {
	file, ok := c.stdin.(interface{ Fd() uintptr })
	if !ok {
		return nil, nil
	}

	fd := int(file.Fd())
	if !term.IsTerminal(fd) {
		return nil, nil
	}

	width, height, err := term.GetSize(fd)
	if err != nil {
		width, height = 80, 24
	}

	termType := strings.TrimSpace(os.Getenv("TERM"))
	if termType == "" {
		termType = "xterm-256color"
	}

	if err := session.RequestPty(termType, height, width, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		return nil, err
	}

	state, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}

	stopWindowSync := forwardWindowChanges(session, fd)
	return func() {
		stopWindowSync()
		_ = term.Restore(fd, state)
	}, nil
}

func forwardWindowSizes(session *ssh.Session, updates <-chan WindowSize) func() {
	if updates == nil {
		return nil
	}

	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			case size, ok := <-updates:
				if !ok {
					return
				}
				if size.Width <= 0 || size.Height <= 0 {
					continue
				}
				_ = session.WindowChange(size.Height, size.Width)
			}
		}
	}()

	return func() {
		close(done)
	}
}

func forwardWindowChanges(session *ssh.Session, fd int) func() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGWINCH)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			case <-signals:
				width, height, err := term.GetSize(fd)
				if err == nil {
					_ = session.WindowChange(height, width)
				}
			}
		}
	}()

	signals <- syscall.SIGWINCH
	return func() {
		signal.Stop(signals)
		close(done)
	}
}

func normalizeTerminal(requested *Terminal) *Terminal {
	if requested == nil {
		return nil
	}

	termType := strings.TrimSpace(requested.Term)
	if termType == "" {
		termType = "xterm-256color"
	}

	width := requested.Width
	if width <= 0 {
		width = 80
	}

	height := requested.Height
	if height <= 0 {
		height = 24
	}

	return &Terminal{
		Term:    termType,
		Width:   width,
		Height:  height,
		MakeRaw: requested.MakeRaw,
		Resize:  requested.Resize,
	}
}
