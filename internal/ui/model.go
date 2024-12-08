/*
Copyright © 2024 Michael Bruskov <mixanemca@yandex.ru>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cloudflare/cloudflare-go"
	"github.com/mixanemca/cfdnscli/internal/app"
	"github.com/mixanemca/cfdnscli/internal/ui/theme"
)

const (
	checkMark = "✓"
	crossMark = "𐄂"
)

const (
	zonesTable = "zones"
	rrsetTable = "rrset"
)

const (
	tableStatusRecord  = "record"
	tableStatusRecords = "records"
	tableStatusZone    = "zone"
	tableStatusZones   = "zones"
)

const (
	headerHeight = 3
	statusHeight = 1
	menuHeight   = 1
)

// Ensure that model fulfils the tea.Model interface at compile time.
var _ tea.Model = (*Model)(nil)

// Custom tea.Msg to switching between zones and rrset tables.
type (
	switchTableToRRSetCmd string
	dataLoadedMsg         struct{}
	dataLoadingMsg        struct{}
	updateRRSetMsg        struct{}
)

// Model represents model for implements bubbletea.Model interface
type Model struct {
	width      int
	height     int
	spinner    spinner.Model
	loading    bool
	current    *table.Model
	rrsetCache map[string][]cloudflare.DNSRecord

	ClientTimeout time.Duration

	ZonesTable table.Model
	RRSetTable table.Model
	TableStyle table.Styles
	ViewStyle  lipgloss.Style
}

func NewModel() *Model {
	var m Model

	m.rrsetCache = make(map[string][]cloudflare.DNSRecord)
	m.ViewStyle = lipgloss.NewStyle().
		Padding(0, 0).
		Width(m.width)
	m.spinner = spinner.New()
	m.spinner.Spinner = spinner.Points
	m.loading = true

	return &m
}

// This command will be executed immediately when the program starts.
// Implements tea.Model interface.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick, // Start the spinner
		func() tea.Msg {
			a, _ := app.New()
			ctx, cancel := context.WithTimeout(context.Background(), m.ClientTimeout)
			defer cancel()

			zones, _ := a.Zones().List(ctx)
			rows := []table.Row{}
			cmds := []tea.Cmd{} // Commands list for async updating

			for _, zone := range zones {
				rows = append(rows, table.Row{
					zone.Name,
					strings.Join(zone.NameServers, ", "),
				})
				cmds = append(cmds, m.updateRRSet(zone.Name))
			}
			m.ZonesTable.SetRows(rows)
			m.current = &m.ZonesTable

			// Return the command that runs all async updates
			return tea.Batch(cmds...)()
		})
}

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
// Implements tea.Model interface.
func (m *Model) View() string {
	table := m.viewZones()
	if m.RRSetTable.Focused() {
		m.current = &m.RRSetTable
		table = m.viewRRSet()
	}

	return m.ViewStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			m.viewHeader(),
			table,
			m.viewStatusBar(),
			m.viewMenu(),
		),
	)
}

// Takes a tea.Msg as input and uses a type switch to handle different types of messages.
// Each case in the switch statement corresponds to a specific message type.
// Implements tea.Model interface.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	// message is sent when the window size changes
	// save to reflect the new dimensions of the terminal window.
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

	// message is sent when a key is pressed.
	case tea.KeyMsg:
		switch msg.String() {
		// Toggles the focus state of the process table
		case "esc":
			m.switchTable(zonesTable)
		// Moves the focus up in the process table if the table is focused.
		case "up", "k":
			m.current.MoveUp(1)
		// Moves the focus down in the process table if the table is focused.
		case "down", "j":
			m.current.MoveDown(1)
		// Reload RRSet
		case "r":
			return m, func() tea.Msg { return dataLoadingMsg{} }
		// Quits the program by returning the tea.Quit command.
		case "q", "ctrl+c":
			return m, tea.Quit
		}
		switch msg.Type {
		case tea.KeyEnter, tea.KeySpace:
			return m, m.handleEnter(msg)
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	// Custom messages
	case dataLoadingMsg:
		m.loading = true
		return m, func() tea.Msg { return updateRRSetMsg{} }

	case updateRRSetMsg:
		zone := m.ZonesTable.SelectedRow()
		return m, m.updateRRSet(zone[0])

	case dataLoadedMsg:
		m.loading = false
		return m, nil // stop spinner

	case switchTableToRRSetCmd:
		m.switchTable(rrsetTable)
	}

	if m.loading {
		m.spinner, cmd = m.spinner.Update(msg)
	}

	// If the message type does not match any of the handled cases, the model is returned unchanged, and no new command is issued.
	return m, cmd
}

func (m *Model) viewHeader() string {
	headerStyle := lipgloss.NewStyle().
		Padding(1, 1).
		Width(m.width).
		Height(headerHeight)

	return headerStyle.Render("CloudFlare DNS CLI")
}

func (m *Model) viewMenu() string {
	// return fmt.Sprintf("Press Enter to select, Esc to return, arrow keys to move, / to find, Ctrl+C or q to exit\n")
	menuStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Width(m.width).
		Height(menuHeight)

	menu := []string{
		"[↑/↓/←/→] Navigate",
		"[Enter] Edit",
		"[Esc] Exit edit",
		"[r] Reload",
		"[q] Quit",
	}

	return menuStyle.Render(strings.Join(menu, " | "))
}

func (m *Model) viewStatusBar() string {
	statusStyle := lipgloss.NewStyle().
		Foreground(theme.Color.Secondary).
		Padding(0, 1).
		Width(m.width).
		Height(statusHeight)

	if m.loading {
		return statusStyle.Render(fmt.Sprintf("Loading... %s", m.spinner.View()))
	}
	rows := len(m.RRSetTable.Rows())
	table := tableStatusRecords
	if rows == 1 {
		table = tableStatusRecord
	}
	if m.ZonesTable.Focused() {
		rows = len(m.ZonesTable.Rows())
		table = tableStatusZones
		if rows == 1 {
			table = tableStatusZone
		}
	}

	return statusStyle.Render(fmt.Sprintf("Loaded %d %s", rows, table))
}

func (m *Model) viewZones() string {
	return m.ZonesTable.View()
}

func (m *Model) viewRRSet() string {
	var rrset []cloudflare.DNSRecord

	selectedRow := m.ZonesTable.SelectedRow()
	if len(selectedRow) > 0 {
		if _, ok := m.rrsetCache[selectedRow[0]]; ok {
			rrset = m.rrsetCache[selectedRow[0]]
		}
	}

	rows := []table.Row{}
	for _, rr := range rrset {
		rows = append(rows, table.Row{
			rr.Name,
			strconv.Itoa(rr.TTL),
			rr.Type,
			boolToCheckMark(cloudflare.Bool(rr.Proxied)),
			rr.Content,
		})
	}
	m.RRSetTable.SetRows(rows)

	return m.RRSetTable.View()
}

func (m *Model) handleEnter(tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return switchTableToRRSetCmd(rrsetTable)
	}
}

// updateRRSet updates resource records set by given zone name.
func (m *Model) updateRRSet(zone string) tea.Cmd {
	return func() tea.Msg {
		a, _ := app.New()
		ctx, cancel := context.WithTimeout(context.Background(), m.ClientTimeout)
		defer cancel()

		rrset, _ := a.Zones().ListRecordsByZoneName(ctx, zone, cloudflare.ListDNSRecordsParams{})
		m.rrsetCache[zone] = rrset

		// Return message that data is loaded.
		return dataLoadedMsg{}
	}
}

// switchTable switches focus between zones and rrset tables
func (m *Model) switchTable(name string) {
	switch name {
	case zonesTable:
		m.ZonesTable.Focus()
		m.RRSetTable.Blur()
		m.current = &m.ZonesTable
	case rrsetTable:
		m.ZonesTable.Blur()
		m.RRSetTable.Focus()
		m.current = &m.RRSetTable
	}
}

func boolToCheckMark(b bool) string {
	if b {
		return checkMark
	}
	return crossMark
}
