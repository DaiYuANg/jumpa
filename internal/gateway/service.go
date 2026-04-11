package gateway

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	config2 "github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/identity"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
	bastiondomain "github.com/DaiYuANg/jumpa/internal/modules/bastion/domain"
	"github.com/samber/mo"
	"golang.org/x/crypto/ssh"
)

const sshServerVersion = "SSH-2.0-jumpa"

type Service struct {
	cfg           config2.AppConfig
	log           *slog.Logger
	authenticator identity.Authenticator
	targetSvc     application.TargetService
	accessSvc     application.AccessService
	sessionSvc    application.SessionRuntimeService

	mu       sync.Mutex
	listener net.Listener
}

type execRequest struct {
	Command string
}

type subsystemRequest struct {
	Name string
}

type ptyRequest struct {
	Term    string
	Columns uint32
	Rows    uint32
	Width   uint32
	Height  uint32
	Modes   string
}

type loginTarget struct {
	Principal   string
	HostName    string
	AccountName string
}

type connectionTarget struct {
	Login   loginTarget
	Host    bastiondomain.Host
	Account mo.Option[bastiondomain.HostAccount]
}

type sessionOptions struct {
	env       map[string]string
	term      string
	columns   int
	rows      int
	hasPTY    bool
	subsystem string
	command   string
	shell     bool
}

func NewService(cfg config2.AppConfig, log *slog.Logger, authenticator identity.Authenticator, targetSvc application.TargetService, accessSvc application.AccessService, sessionSvc application.SessionRuntimeService) *Service {
	return &Service{cfg: cfg, log: log, authenticator: authenticator, targetSvc: targetSvc, accessSvc: accessSvc, sessionSvc: sessionSvc}
}

func (s *Service) Start(_ context.Context) error {
	if !s.cfg.Bastion.Enabled {
		s.log.Info("gateway disabled", slog.String("listen_addr", s.cfg.Bastion.SSH.ListenAddr))
		return nil
	}

	addr := s.cfg.Bastion.SSH.ListenAddr
	if addr == "" {
		addr = ":2222"
	}

	serverConfig, hostKeySource, err := s.newSSHServerConfig()
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.listener = ln
	s.mu.Unlock()

	descriptor := s.authenticator.Descriptor()
	s.log.Info("gateway listening",
		slog.String("listen_addr", ln.Addr().String()),
		slog.String("identity_provider", descriptor.Kind),
		slog.String("identity_backend", descriptor.Backend),
		slog.String("host_key_source", hostKeySource),
		slog.Bool("password_auth_enabled", s.authenticator.SupportsPassword()),
	)

	go s.serve(ln, serverConfig)
	return nil
}

func (s *Service) serve(ln net.Listener, serverConfig *ssh.ServerConfig) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			s.log.Error("gateway accept failed", slog.String("error", err.Error()))
			return
		}
		go s.handleConn(conn, serverConfig)
	}
}

func (s *Service) handleConn(conn net.Conn, serverConfig *ssh.ServerConfig) {
	defer func() { _ = conn.Close() }()

	serverConn, channels, requests, err := ssh.NewServerConn(conn, serverConfig)
	if err != nil {
		s.log.Warn("ssh handshake failed",
			slog.String("remote_addr", conn.RemoteAddr().String()),
			slog.String("error", err.Error()),
		)
		return
	}
	defer func() { _ = serverConn.Close() }()

	extensions := serverConn.Permissions.Extensions
	s.log.Info("gateway client authenticated",
		slog.String("remote_addr", serverConn.RemoteAddr().String()),
		slog.String("principal", extensions["principal"]),
		slog.String("target_host", extensions["target_host"]),
		slog.String("target_account", extensions["target_account"]),
		slog.String("identity_provider", extensions["provider"]),
		slog.String("identity_backend", extensions["backend"]),
	)

	go ssh.DiscardRequests(requests)

	for newChannel := range channels {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "only session channels are supported")
			continue
		}
		go s.handleSessionChannel(serverConn, newChannel)
	}
}

func (s *Service) handleSessionChannel(serverConn *ssh.ServerConn, newChannel ssh.NewChannel) {
	channel, requests, err := newChannel.Accept()
	if err != nil {
		s.log.Warn("session channel accept failed",
			slog.String("remote_addr", serverConn.RemoteAddr().String()),
			slog.String("error", err.Error()),
		)
		return
	}
	defer func() { _ = channel.Close() }()

	opts, ack, err := collectSessionOptions(requests)
	if err != nil {
		s.writeProxyError(channel, err)
		return
	}
	target, err := s.targetFromPermissions(serverConn.Permissions)
	if err != nil {
		s.writeProxyError(channel, err)
		return
	}
	bastionSession, err := s.sessionSvc.Start(context.Background(), application.StartSessionInput{
		PrincipalName: serverConn.Permissions.Extensions["principal"],
		HostID:        target.Host.ID,
		HostName:      target.Host.Name,
		HostAccountID: targetAccountIDPtr(target),
		HostAccount:   target.Login.AccountName,
		Protocol:      "ssh",
		SourceAddr:    serverConn.RemoteAddr().String(),
	})
	if err != nil {
		s.log.Error("gateway session registration failed",
			slog.String("remote_addr", serverConn.RemoteAddr().String()),
			slog.String("principal", serverConn.Permissions.Extensions["principal"]),
			slog.String("target_host", target.Host.Name),
			slog.String("error", err.Error()),
		)
		s.writeProxyError(channel, errors.New("session registration failed"))
		return
	}
	if err := s.sessionSvc.RecordEvent(context.Background(), bastionSession.ID, "channel_opened", sessionPayload(opts)); err != nil {
		s.log.Warn("gateway session event failed",
			slog.String("session_id", bastionSession.ID),
			slog.String("event_type", "channel_opened"),
			slog.String("error", err.Error()),
		)
	}
	client, err := s.openTargetClient(target, serverConn.Permissions.Extensions["login_password"])
	if err != nil {
		s.failGatewaySession(bastionSession.ID, "target_connect_failed", err)
		s.writeProxyError(channel, err)
		return
	}
	defer func() { _ = client.Close() }()

	session, err := client.NewSession()
	if err != nil {
		s.failGatewaySession(bastionSession.ID, "target_session_failed", err)
		s.writeProxyError(channel, err)
		return
	}
	defer func() { _ = session.Close() }()

	session.Stdout = channel
	session.Stderr = channel.Stderr()
	stdin, err := session.StdinPipe()
	if err != nil {
		s.failGatewaySession(bastionSession.ID, "stdin_pipe_failed", err)
		s.writeProxyError(channel, err)
		return
	}

	for _, req := range ack {
		_ = req.Reply(true, nil)
	}

	for k, v := range opts.env {
		if err := session.Setenv(k, v); err != nil {
			s.log.Debug("setenv failed", slog.String("key", k), slog.String("error", err.Error()))
		}
	}
	if opts.hasPTY {
		if err := session.RequestPty(opts.term, opts.rows, opts.columns, ssh.TerminalModes{}); err != nil {
			s.failGatewaySession(bastionSession.ID, "pty_request_failed", err)
			s.writeProxyError(channel, err)
			return
		}
	}

	switch {
	case opts.shell:
		if err := session.Shell(); err != nil {
			s.failGatewaySession(bastionSession.ID, "shell_start_failed", err)
			s.writeProxyError(channel, err)
			return
		}
	case opts.command != "":
		if err := session.Start(opts.command); err != nil {
			s.failGatewaySession(bastionSession.ID, "exec_start_failed", err)
			s.writeProxyError(channel, err)
			return
		}
	case opts.subsystem != "":
		if err := session.RequestSubsystem(opts.subsystem); err != nil {
			s.failGatewaySession(bastionSession.ID, "subsystem_start_failed", err)
			s.writeProxyError(channel, err)
			return
		}
	default:
		s.failGatewaySession(bastionSession.ID, "invalid_session_request", errors.New("no shell, exec, or subsystem request received"))
		s.writeProxyError(channel, errors.New("no shell, exec, or subsystem request received"))
		return
	}

	if err := s.sessionSvc.MarkActive(context.Background(), bastionSession.ID); err != nil {
		s.log.Warn("gateway session activation failed",
			slog.String("session_id", bastionSession.ID),
			slog.String("error", err.Error()),
		)
	}
	if err := s.sessionSvc.RecordEvent(context.Background(), bastionSession.ID, "proxy_opened", sessionPayload(opts)); err != nil {
		s.log.Warn("gateway session event failed",
			slog.String("session_id", bastionSession.ID),
			slog.String("event_type", "proxy_opened"),
			slog.String("error", err.Error()),
		)
	}

	go func() {
		_, _ = io.Copy(stdin, channel)
		_ = stdin.Close()
	}()

	waitErr := session.Wait()
	if waitErr != nil {
		s.failGatewaySession(bastionSession.ID, "proxy_closed", waitErr)
	} else {
		if err := s.sessionSvc.RecordEvent(context.Background(), bastionSession.ID, "proxy_closed", map[string]string{"exit_code": "0"}); err != nil {
			s.log.Warn("gateway session event failed",
				slog.String("session_id", bastionSession.ID),
				slog.String("event_type", "proxy_closed"),
				slog.String("error", err.Error()),
			)
		}
		if err := s.sessionSvc.Finish(context.Background(), bastionSession.ID, "closed"); err != nil {
			s.log.Warn("gateway session completion failed",
				slog.String("session_id", bastionSession.ID),
				slog.String("error", err.Error()),
			)
		}
	}
	_ = sendExitStatus(channel, exitCodeFromError(waitErr))
}

func collectSessionOptions(requests <-chan *ssh.Request) (sessionOptions, []*ssh.Request, error) {
	opts := sessionOptions{env: map[string]string{}}
	ack := make([]*ssh.Request, 0, 4)

	for req := range requests {
		switch req.Type {
		case "pty-req":
			var payload ptyRequest
			if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
				_ = req.Reply(false, nil)
				return opts, ack, err
			}
			opts.term = payload.Term
			opts.columns = int(payload.Columns)
			opts.rows = int(payload.Rows)
			opts.hasPTY = true
			ack = append(ack, req)
		case "env":
			name, value, ok := parseEnvPayload(req.Payload)
			if !ok {
				_ = req.Reply(false, nil)
				return opts, ack, errors.New("invalid env payload")
			}
			opts.env[name] = value
			ack = append(ack, req)
		case "shell":
			opts.shell = true
			ack = append(ack, req)
			return opts, ack, nil
		case "exec":
			var payload execRequest
			if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
				_ = req.Reply(false, nil)
				return opts, ack, err
			}
			opts.command = payload.Command
			ack = append(ack, req)
			return opts, ack, nil
		case "subsystem":
			var payload subsystemRequest
			if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
				_ = req.Reply(false, nil)
				return opts, ack, err
			}
			opts.subsystem = payload.Name
			ack = append(ack, req)
			return opts, ack, nil
		case "window-change":
			ack = append(ack, req)
		default:
			_ = req.Reply(false, nil)
			return opts, ack, fmt.Errorf("unsupported session request: %s", req.Type)
		}
	}

	return opts, ack, nil
}

func parseEnvPayload(payload []byte) (string, string, bool) {
	var data struct {
		Name  string
		Value string
	}
	if err := ssh.Unmarshal(payload, &data); err != nil {
		return "", "", false
	}
	return data.Name, data.Value, true
}

func (s *Service) newSSHServerConfig() (*ssh.ServerConfig, string, error) {
	signer, source, err := loadHostSigner(s.cfg.Bastion.SSH.HostKeyPath)
	if err != nil {
		return nil, "", err
	}

	cfg := &ssh.ServerConfig{
		ServerVersion: sshServerVersion,
		PasswordCallback: func(metadata ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			if !s.authenticator.SupportsPassword() {
				return nil, errors.New("password authentication is not available for the configured identity provider")
			}

			login, err := parseLoginTarget(metadata.User())
			if err != nil {
				return nil, errors.New("invalid login target")
			}

			target, err := s.resolveConnectionTarget(context.Background(), login)
			if err != nil {
				s.log.Warn("gateway target resolution failed",
					slog.String("login", metadata.User()),
					slog.String("error", err.Error()),
				)
				return nil, errors.New("target resolution failed")
			}

			authn, err := s.authenticator.AuthenticatePassword(context.Background(), identity.PasswordCredentials{
				Username:   login.Principal,
				Password:   string(password),
				RemoteAddr: metadata.RemoteAddr().String(),
			})
			if err != nil {
				s.log.Warn("gateway password authentication failed",
					slog.String("remote_addr", metadata.RemoteAddr().String()),
					slog.String("principal", login.Principal),
					slog.String("error", err.Error()),
				)
				return nil, errors.New("authentication failed")
			}

			decision, err := s.accessSvc.Authorize(context.Background(), application.AccessCheckInput{
				PrincipalName:  authn.Username,
				PrincipalEmail: stringAttribute(authn, "email"),
				HostName:       target.Host.Name,
				AccountName:    target.Login.AccountName,
				Protocol:       "ssh",
			})
			if err != nil {
				s.log.Error("gateway access check failed",
					slog.String("principal", authn.Username),
					slog.String("target_host", target.Host.Name),
					slog.String("target_account", target.Login.AccountName),
					slog.String("error", err.Error()),
				)
				return nil, errors.New("access evaluation failed")
			}
			if !decision.Allowed {
				s.log.Warn("gateway access denied",
					slog.String("principal", authn.Username),
					slog.String("target_host", target.Host.Name),
					slog.String("target_account", target.Login.AccountName),
					slog.String("reason", decision.Reason),
				)
				return nil, errors.New(decision.Reason)
			}

			return &ssh.Permissions{
				Extensions: map[string]string{
					"principal":          authn.Username,
					"provider":           authn.Provider.Kind,
					"backend":            authn.Provider.Backend,
					"policy_id":          decision.MatchedPolicyID,
					"recording_required": strconv.FormatBool(decision.RecordingRequired),
					"target_host_id":     target.Host.ID,
					"target_host":        target.Host.Name,
					"target_address":     target.Host.Address,
					"target_port":        strconv.Itoa(target.Host.Port),
					"target_account":     target.Login.AccountName,
					"target_account_id":  targetAccountID(target),
					"target_auth":        targetAuthType(target),
					"target_credref":     targetCredentialRef(target),
					"login_password":     string(password),
				},
			}, nil
		},
	}
	cfg.AddHostKey(signer)
	return cfg, source, nil
}

func (s *Service) resolveConnectionTarget(ctx context.Context, login loginTarget) (connectionTarget, error) {
	hostOpt, err := s.targetSvc.GetHostByName(ctx, login.HostName)
	if err != nil {
		return connectionTarget{}, err
	}
	if hostOpt.IsAbsent() {
		return connectionTarget{}, fmt.Errorf("host %q not found", login.HostName)
	}
	host := hostOpt.MustGet()
	if !strings.EqualFold(host.Protocol, "ssh") {
		return connectionTarget{}, fmt.Errorf("host %q does not support ssh", login.HostName)
	}
	if !host.JumpEnabled {
		return connectionTarget{}, fmt.Errorf("host %q is disabled", login.HostName)
	}

	accountName := login.AccountName
	if accountName == "" {
		accountName = login.Principal
	}

	accountOpt, err := s.targetSvc.GetHostAccountByName(ctx, host.ID, accountName)
	if err != nil {
		return connectionTarget{}, err
	}

	return connectionTarget{
		Login: loginTarget{
			Principal:   login.Principal,
			HostName:    host.Name,
			AccountName: accountName,
		},
		Host:    host,
		Account: accountOpt,
	}, nil
}

func parseLoginTarget(input string) (loginTarget, error) {
	parts := strings.Split(strings.TrimSpace(input), "#")
	if len(parts) < 2 || len(parts) > 3 {
		return loginTarget{}, errors.New("login must be principal#host or principal#host#account")
	}
	target := loginTarget{
		Principal: strings.TrimSpace(parts[0]),
		HostName:  strings.TrimSpace(parts[1]),
	}
	if len(parts) == 3 {
		target.AccountName = strings.TrimSpace(parts[2])
	}
	if target.Principal == "" || target.HostName == "" {
		return loginTarget{}, errors.New("principal and host are required")
	}
	return target, nil
}

func targetAuthType(target connectionTarget) string {
	if target.Account.IsPresent() && strings.TrimSpace(target.Account.MustGet().AuthenticationType) != "" {
		return target.Account.MustGet().AuthenticationType
	}
	return target.Host.Authentication
}

func targetCredentialRef(target connectionTarget) string {
	if target.Account.IsPresent() {
		if ref := target.Account.MustGet().CredentialRef; ref != nil {
			return *ref
		}
	}
	return ""
}

func targetAccountID(target connectionTarget) string {
	if target.Account.IsPresent() {
		return target.Account.MustGet().ID
	}
	return ""
}

func targetAccountIDPtr(target connectionTarget) *string {
	if target.Account.IsPresent() {
		value := target.Account.MustGet().ID
		if value != "" {
			return &value
		}
	}
	return nil
}

func stringAttribute(authn identity.Authentication, key string) string {
	if authn.Attributes == nil {
		return ""
	}
	value, ok := authn.Attributes.Get(key)
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func (s *Service) targetFromPermissions(permissions *ssh.Permissions) (connectionTarget, error) {
	if permissions == nil {
		return connectionTarget{}, errors.New("missing permissions")
	}
	ext := permissions.Extensions
	port, err := strconv.Atoi(ext["target_port"])
	if err != nil {
		return connectionTarget{}, err
	}

	host := bastiondomain.Host{
		ID:             ext["target_host_id"],
		Name:           ext["target_host"],
		Address:        ext["target_address"],
		Port:           port,
		Protocol:       "ssh",
		Authentication: ext["target_auth"],
		JumpEnabled:    true,
	}
	accountName := ext["target_account"]
	account := mo.None[bastiondomain.HostAccount]()
	if accountName != "" {
		var ref *string
		if value := ext["target_credref"]; value != "" {
			ref = &value
		}
		account = mo.Some(bastiondomain.HostAccount{
			ID:                 ext["target_account_id"],
			AccountName:        accountName,
			AuthenticationType: ext["target_auth"],
			CredentialRef:      ref,
		})
	}

	return connectionTarget{
		Login: loginTarget{
			Principal:   ext["principal"],
			HostName:    host.Name,
			AccountName: accountName,
		},
		Host:    host,
		Account: account,
	}, nil
}

func (s *Service) openTargetClient(target connectionTarget, loginPassword string) (*ssh.Client, error) {
	authMethod, err := s.resolveTargetAuthMethod(target, loginPassword)
	if err != nil {
		return nil, err
	}

	cfg := &ssh.ClientConfig{
		User:            target.Login.AccountName,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	address := net.JoinHostPort(target.Host.Address, strconv.Itoa(target.Host.Port))
	return ssh.Dial("tcp", address, cfg)
}

func (s *Service) resolveTargetAuthMethod(target connectionTarget, loginPassword string) (ssh.AuthMethod, error) {
	authType := strings.ToLower(strings.TrimSpace(targetAuthType(target)))
	credRef := strings.TrimSpace(targetCredentialRef(target))

	switch {
	case authType == "passthrough":
		if loginPassword == "" {
			return nil, errors.New("passthrough authentication requires login password")
		}
		return ssh.Password(loginPassword), nil
	case strings.HasPrefix(credRef, "env:"):
		envName := strings.TrimPrefix(credRef, "env:")
		secret := os.Getenv(envName)
		if secret == "" {
			return nil, fmt.Errorf("credential env %q is empty", envName)
		}
		return ssh.Password(secret), nil
	case strings.HasPrefix(credRef, "file:"):
		path := strings.TrimPrefix(credRef, "file:")
		keyBytes, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		signer, err := ssh.ParsePrivateKey(keyBytes)
		if err != nil {
			return nil, err
		}
		return ssh.PublicKeys(signer), nil
	default:
		return nil, fmt.Errorf("unsupported target credential reference %q", credRef)
	}
}

func (s *Service) failGatewaySession(sessionID, eventType string, err error) {
	payload := map[string]string{
		"error": err.Error(),
	}
	if eventType == "proxy_closed" {
		payload["exit_code"] = strconv.FormatUint(uint64(exitCodeFromError(err)), 10)
	}
	if eventErr := s.sessionSvc.RecordEvent(context.Background(), sessionID, eventType, payload); eventErr != nil {
		s.log.Warn("gateway session event failed",
			slog.String("session_id", sessionID),
			slog.String("event_type", eventType),
			slog.String("error", eventErr.Error()),
		)
	}
	if finishErr := s.sessionSvc.Finish(context.Background(), sessionID, "failed"); finishErr != nil {
		s.log.Warn("gateway session completion failed",
			slog.String("session_id", sessionID),
			slog.String("error", finishErr.Error()),
		)
	}
}

func sessionPayload(opts sessionOptions) map[string]string {
	payload := map[string]string{
		"has_pty": strconv.FormatBool(opts.hasPTY),
	}
	if opts.term != "" {
		payload["term"] = opts.term
	}
	if opts.command != "" {
		payload["command"] = opts.command
	}
	if opts.subsystem != "" {
		payload["subsystem"] = opts.subsystem
	}
	if opts.shell {
		payload["mode"] = "shell"
	} else if opts.command != "" {
		payload["mode"] = "exec"
	} else if opts.subsystem != "" {
		payload["mode"] = "subsystem"
	}
	if opts.columns > 0 {
		payload["columns"] = strconv.Itoa(opts.columns)
	}
	if opts.rows > 0 {
		payload["rows"] = strconv.Itoa(opts.rows)
	}
	return payload
}

func loadHostSigner(path string) (ssh.Signer, string, error) {
	if path != "" {
		keyBytes, err := os.ReadFile(path)
		if err == nil {
			signer, parseErr := ssh.ParsePrivateKey(keyBytes)
			if parseErr != nil {
				return nil, "", parseErr
			}
			return signer, "file", nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, "", err
		}
	}

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, "", err
	}
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, "", err
	}
	return signer, "ephemeral", nil
}

func (s *Service) writeProxyError(channel ssh.Channel, err error) {
	writeSessionMessage(channel,
		"jumpa gateway proxy error",
		err.Error(),
	)
	_ = sendExitStatus(channel, 1)
}

func writeSessionMessage(channel ssh.Channel, lines ...string) {
	for _, line := range lines {
		_, _ = channel.Write([]byte(line + "\r\n"))
	}
}

func exitCodeFromError(err error) uint32 {
	if err == nil {
		return 0
	}
	var exitErr *ssh.ExitError
	if errors.As(err, &exitErr) {
		return uint32(exitErr.ExitStatus())
	}
	return 1
}

func sendExitStatus(channel ssh.Channel, code uint32) error {
	_, err := channel.SendRequest("exit-status", false, ssh.Marshal(struct {
		Status uint32
	}{Status: code}))
	return err
}

func (s *Service) Shutdown(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener == nil {
		return nil
	}

	err := s.listener.Close()
	s.listener = nil
	if errors.Is(err, net.ErrClosed) {
		return nil
	}
	return err
}
