package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/DaiYuANg/jumpa/internal/cli/api"
	"github.com/samber/lo"
)

type Tab int

const (
	TabOverview Tab = iota
	TabHosts
	TabRequests
	TabSessions
	TabGateways
)

type LaunchRequest struct {
	Target      string
	GatewayHost string
	GatewayPort string
}

type Options struct {
	Principal   string
	GatewayAddr string
	Me          api.Me
	AltScreen   bool
}

type snapshot struct {
	Overview api.Overview
	Hosts    []api.Host
	Requests []api.AccessRequest
	Sessions []api.Session
	Gateways []api.Gateway
}

type loadMsg struct {
	data snapshot
	err  error
}

type Model struct {
	client      *api.Client
	principal   string
	gatewayAddr string
	me          api.Me
	altScreen   bool
	active      Tab
	selected    map[Tab]int
	data        snapshot
	loading     bool
	err         error
	message     string
	width       int
	height      int
	launch      *LaunchRequest
}

func New(client *api.Client, opts Options) Model {
	return Model{
		client:      client,
		principal:   strings.TrimSpace(opts.Principal),
		gatewayAddr: strings.TrimSpace(opts.GatewayAddr),
		me:          opts.Me,
		altScreen:   opts.AltScreen,
		active:      TabHosts,
		selected: map[Tab]int{
			TabOverview: 0,
			TabHosts:    0,
			TabRequests: 0,
			TabSessions: 0,
			TabGateways: 0,
		},
		loading: true,
		message: "loading control-plane data",
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadCmd()
}

func (m Model) loadCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		overview, err := m.client.Overview(ctx)
		if err != nil {
			return loadMsg{err: err}
		}
		hosts, err := m.client.Hosts(ctx)
		if err != nil {
			return loadMsg{err: err}
		}
		requests, err := m.client.AccessRequests(ctx, api.AccessRequestQuery{
			Page:     1,
			PageSize: 50,
		})
		if err != nil {
			return loadMsg{err: err}
		}
		sessions, err := m.client.Sessions(ctx)
		if err != nil {
			return loadMsg{err: err}
		}
		gateways, err := m.client.Gateways(ctx)
		if err != nil {
			return loadMsg{err: err}
		}

		return loadMsg{data: snapshot{
			Overview: overview,
			Hosts:    hosts,
			Requests: requests.Items,
			Sessions: sessions,
			Gateways: gateways,
		}}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case loadMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.data = msg.data
			m.message = "data refreshed"
		} else {
			m.message = "refresh failed"
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab", "right", "l":
			m.active = (m.active + 1) % 5
			return m, nil
		case "shift+tab", "left", "h":
			m.active = (m.active + 4) % 5
			return m, nil
		case "j", "down":
			m.move(1)
			return m, nil
		case "k", "up":
			m.move(-1)
			return m, nil
		case "r":
			m.loading = true
			m.err = nil
			m.message = "refreshing"
			return m, m.loadCmd()
		case "enter", "c":
			if req := m.currentLaunch(); req != nil {
				m.launch = req
				return m, tea.Quit
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) move(delta int) {
	max := m.currentCount()
	if max <= 0 {
		m.selected[m.active] = 0
		return
	}
	next := m.selected[m.active] + delta
	if next < 0 {
		next = max - 1
	}
	if next >= max {
		next = 0
	}
	m.selected[m.active] = next
}

func (m Model) currentCount() int {
	switch m.active {
	case TabHosts:
		return len(m.data.Hosts)
	case TabRequests:
		return len(m.data.Requests)
	case TabSessions:
		return len(m.data.Sessions)
	case TabGateways:
		return len(m.data.Gateways)
	default:
		return 1
	}
}

func (m Model) currentLaunch() *LaunchRequest {
	if m.active != TabHosts || len(m.data.Hosts) == 0 {
		return nil
	}
	idx := m.selected[TabHosts]
	if idx < 0 || idx >= len(m.data.Hosts) {
		return nil
	}
	host := m.data.Hosts[idx]
	if !host.JumpEnabled {
		return nil
	}
	target := fmt.Sprintf("%s#%s", m.principal, host.Name)
	gatewayHost, gatewayPort := SplitGatewayAddress(m.gatewayAddr)
	return &LaunchRequest{
		Target:      target,
		GatewayHost: gatewayHost,
		GatewayPort: gatewayPort,
	}
}

func (m Model) View() tea.View {
	if m.width == 0 {
		m.width = 120
	}

	var body string
	if m.loading {
		body = "Loading..."
	} else if m.err != nil {
		body = "Error: " + m.err.Error()
	} else {
		body = m.renderActive()
	}

	header := renderHeader(m)
	footer := renderFooter(m)
	view := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, header, body, footer))
	view.AltScreen = m.altScreen
	return view
}

func (m Model) LaunchRequest() *LaunchRequest {
	return m.launch
}

func (m Model) renderActive() string {
	switch m.active {
	case TabOverview:
		return renderOverview(m)
	case TabHosts:
		return renderSelectableList(m.selected[TabHosts], lo.Map(m.data.Hosts, func(it api.Host, _ int) string {
			state := "off"
			if it.JumpEnabled {
				state = "on"
			}
			return fmt.Sprintf("%-24s %-21s %-5s jump:%s auth:%s", it.Name, fmt.Sprintf("%s:%d", it.Address, it.Port), it.Protocol, state, it.Authentication)
		}))
	case TabRequests:
		return renderSelectableList(m.selected[TabRequests], lo.Map(m.data.Requests, func(it api.AccessRequest, _ int) string {
			return fmt.Sprintf("%-12s %-22s %-14s %-10s %s", it.Status, it.HostName, it.HostAccount, it.Protocol, it.RequestedAt.Local().Format("01-02 15:04"))
		}))
	case TabSessions:
		return renderSelectableList(m.selected[TabSessions], lo.Map(m.data.Sessions, func(it api.Session, _ int) string {
			return fmt.Sprintf("%-12s %-22s %-14s %-12s %s", it.Status, it.HostName, it.HostAccount, it.PrincipalName, it.StartedAt.Local().Format("01-02 15:04"))
		}))
	case TabGateways:
		return renderSelectableList(m.selected[TabGateways], lo.Map(m.data.Gateways, func(it api.Gateway, _ int) string {
			return fmt.Sprintf("%-18s %-8s %-10s %-24s %s", it.NodeName, it.Zone, it.EffectiveStatus, it.AdvertiseAddr, strings.Join(it.Tags, ","))
		}))
	default:
		return ""
	}
}

func renderHeader(m Model) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render("jumpa cli")
	tabStyle := lipgloss.NewStyle().PaddingRight(2)
	activeStyle := tabStyle.Bold(true).Foreground(lipgloss.Color("229"))
	tabs := []string{"Overview", "Hosts", "Requests", "Sessions", "Gateways"}
	rendered := lo.Map(tabs, func(it string, idx int) string {
		if Tab(idx) == m.active {
			return activeStyle.Render(it)
		}
		return tabStyle.Render(it)
	})
	meta := fmt.Sprintf(" user:%s principal:%s gateway:%s ", m.me.Email, m.principal, m.gatewayAddr)
	return lipgloss.JoinVertical(lipgloss.Left, title, strings.Join(rendered, " "), meta)
}

func renderFooter(m Model) string {
	help := "tab/h/l switch  j/k move  r refresh  enter/c connect-host  q quit"
	status := m.message
	if launch := m.currentLaunch(); launch != nil && m.active == TabHosts {
		status = "ssh " + launch.Target + "@" + launch.GatewayHost + " -p " + launch.GatewayPort
	}
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	return style.Render(help + "\n" + status)
}

func renderSelectableList(selected int, items []string) string {
	if len(items) == 0 {
		return "No data."
	}
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("62"))
	lines := lo.Map(items, func(it string, idx int) string {
		prefix := "  "
		line := prefix + it
		if idx == selected {
			line = "> " + it
			return selectedStyle.Render(line)
		}
		return line
	})
	return strings.Join(lines, "\n")
}

func renderOverview(m Model) string {
	lines := []string{
		fmt.Sprintf("Product: %s", m.data.Overview.ProductName),
		fmt.Sprintf("Database: %s", m.data.Overview.DatabaseDriver),
		fmt.Sprintf("Bastion Enabled: %t", m.data.Overview.BastionEnabled),
		fmt.Sprintf("SSH Listen: %s", m.data.Overview.SSHListenAddr),
		fmt.Sprintf("Gateway Target: %s", m.gatewayAddr),
		fmt.Sprintf("Identity Modes: %s", strings.Join(m.data.Overview.IdentityModes, ", ")),
		fmt.Sprintf("Password Ready: %t", m.data.Overview.PasswordAuthReady),
		fmt.Sprintf("Recording Dir: %s", m.data.Overview.RecordingDir),
		fmt.Sprintf("Capability Notes: %s", strings.Join(m.data.Overview.CapabilityNotes, " | ")),
	}
	return strings.Join(lines, "\n")
}
