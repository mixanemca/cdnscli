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

// Package ui holds cfdnscli UI.
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
	"github.com/mixanemca/cfdnscli/internal/app"
	"github.com/mixanemca/cfdnscli/internal/models"
	"github.com/mixanemca/cfdnscli/internal/ui/popup"
	"github.com/mixanemca/cfdnscli/internal/ui/theme"
	overlay "github.com/rmhubbert/bubbletea-overlay"
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

// Messages to control the popup window
type editRowMsg struct {
	row table.Row
}

type (
	saveRowMsg    struct{}
	cancelEditMsg struct{}
)

// Model represents model for implements bubbletea.Model interface.
type Model struct {
	width      int
	height     int
	spinner    spinner.Model
	loading    bool
	current    *table.Model
	rrsetCache map[string][]models.DNSRecord

    // editing
    popup      *popup.Model
	showPopup  bool
	overlay    *overlay.Model
	editRow    table.Row
	editBuffer []string
	cursor     int

	ClientTimeout time.Duration

	ZonesTable table.Model
	RRSetTable table.Model
	TableStyle table.Styles
	ViewStyle  lipgloss.Style
}

// NewModel creates new Model for UI.
func NewModel() *Model {
	var m Model

	m.rrsetCache = make(map[string][]models.DNSRecord)
	m.ViewStyle = lipgloss.NewStyle().
		Padding(0, 0).
		Width(m.width)
	m.spinner = spinner.New()
	m.spinner.Spinner = spinner.Points
	m.loading = true

	return &m
}

// Init This command will be executed immediately when the program starts.
// Implements tea.Model interface.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick, // Start the spinner
		func() tea.Msg {
			a, _ := app.New()
			ctx, cancel := context.WithTimeout(context.Background(), m.ClientTimeout)
			defer cancel()

			zones, _ := a.Provider().ListZones(ctx)
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

    if m.showPopup {
        // Render popup window over the base scene using overlay
        if m.overlay == nil {
            bg := &backgroundViewModel{parent: m}
            m.overlay = overlay.New(m.popup, bg, overlay.Center, overlay.Center, 0, 0)
        }
        return m.overlay.View()
    }

    return m.renderBase(table)
}

// Update Takes a tea.Msg as input and uses a type switch to handle different types of messages.
// Each case in the switch statement corresponds to a specific message type.
// Implements tea.Model interface.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.showPopup {
        updated, cmd := m.popup.Update(msg)
        if pm, ok := updated.(*popup.Model); ok {
            m.popup = pm
        }
        if !m.popup.IsActive { // Close popup when it becomes inactive
			m.showPopup = false
			return m, cmd
		}
		return m, cmd
	}

	switch msg := msg.(type) {
    // Window size changed: save to reflect new terminal dimensions
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

    // Key pressed
	case tea.KeyMsg:
		switch msg.String() {
        // Toggle focus between zones and records
		case "esc":
			if m.showPopup {
				m.switchTable(rrsetTable)
				return m, nil
			}
			m.switchTable(zonesTable)
        // Move focus up in the current table
		case "up", "k":
			m.current.MoveUp(1)
        // Move focus down in the current table
		case "down", "j":
			m.current.MoveDown(1)
        // Open popup editor for the selected record
		case "e":
			// If RRSet is focused, open record editor; if Zones is focused, show NameServers popup
			if m.RRSetTable.Focused() && m.current.Cursor() >= 0 {
				rows := m.current.Rows()
				cursor := m.current.Cursor()
				if cursor < len(rows) {
					row := rows[cursor]
					if len(row) >= 5 { // Name, TTL, Type, Proxied, Content
						// Convert current Proxied value to boolean string
						proxiedStr := "false"
						if row[3] == checkMark {
							proxiedStr = "true"
						}
						initial := []string{row[0], row[1], row[2], proxiedStr, row[4]}
						m.showPopup = true
						m.overlay = nil // recreate overlay on render
						m.popup = popup.New(
							[]string{"Name", "TTL", "Type", "Proxied", "Content"},
							initial,
							"Resource record editing",
							func(fields []string) tea.Msg {
								return popup.SaveActionMsg{Fields: fields}
							},
							popup.CancelMsg{},
						)
					}
				}
			} else if m.ZonesTable.Focused() && m.ZonesTable.Cursor() >= 0 {
				rows := m.ZonesTable.Rows()
				cursor := m.ZonesTable.Cursor()
				if cursor < len(rows) {
					row := rows[cursor]
					if len(row) >= 2 { // Name, NameServers
						zoneName := row[0]
						nameServers := row[1]
						m.showPopup = true
						m.overlay = nil
					// Build initial list from comma-separated string
					parts := strings.Split(nameServers, ",")
					var initial []string
					for _, p := range parts { if s := strings.TrimSpace(p); s != "" { initial = append(initial, s) } }
					m.popup = popup.NewNameServersEditor(initial, fmt.Sprintf("Zone: %s — NameServers", zoneName))
					}
				}
			}
		// Reload RRSet
		case "r":
			return m, func() tea.Msg { return dataLoadingMsg{} }
		// Quits the program by returning the tea.Quit command.
		case "q", "ctrl+c":
			return m, tea.Quit
		}
		switch msg.Type {
		case tea.KeyEnter:
			// Replicate 'e' behavior for Enter: open editor depending on focused table
			if m.RRSetTable.Focused() && m.current.Cursor() >= 0 {
				rows := m.current.Rows()
				cursor := m.current.Cursor()
				if cursor < len(rows) {
					row := rows[cursor]
					if len(row) >= 5 {
						proxiedStr := "false"
						if row[3] == checkMark { proxiedStr = "true" }
						initial := []string{row[0], row[1], row[2], proxiedStr, row[4]}
						m.showPopup = true
						m.overlay = nil
						m.popup = popup.New(
							[]string{"Name", "TTL", "Type", "Proxied", "Content"},
							initial,
							"Resource record editing",
							func(fields []string) tea.Msg { return popup.SaveActionMsg{Fields: fields} },
							popup.CancelMsg{},
						)
					}
				}
				return m, nil
			}
			if m.ZonesTable.Focused() && m.ZonesTable.Cursor() >= 0 {
				rows := m.ZonesTable.Rows()
				cursor := m.ZonesTable.Cursor()
				if cursor < len(rows) {
					row := rows[cursor]
					if len(row) >= 2 {
						zoneName := row[0]
						nameServers := row[1]
						parts := strings.Split(nameServers, ",")
						var initial []string
						for _, p := range parts { if s := strings.TrimSpace(p); s != "" { initial = append(initial, s) } }
						m.showPopup = true
						m.overlay = nil
						m.popup = popup.NewNameServersEditor(initial, fmt.Sprintf("Zone: %s — NameServers", zoneName))
					}
				}
				return m, nil
			}
			return m, nil
		case tea.KeySpace:
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

    case editRowMsg:
        // Open editing window
		m.showPopup = true
		m.editRow = msg.row
		m.editBuffer = append([]string{}, msg.row...)
		m.cursor = 0
		return m, nil

	case saveRowMsg:
        // Save changes
		copy(m.editRow, m.editBuffer)
		m.showPopup = false
		return m, nil

	case cancelEditMsg:
        // Cancel editing
		m.showPopup = false
		return m, nil
    // Handle save and cancel messages from popup
	case popup.SaveActionMsg:
        // Save changes to the table and perform UpdateRR
		m.updateTableRow(m.current.Cursor(), msg.Fields)
		return m, m.updateRRFromFields(msg.Fields)
	case popup.SaveNameServersMsg:
		// Update zones table NameServers column with joined values
		if m.ZonesTable.Focused() {
			cursor := m.ZonesTable.Cursor()
			rows := m.ZonesTable.Rows()
			if cursor >= 0 && cursor < len(rows) {
				joined := strings.Join(msg.Servers, ", ")
				rows[cursor][1] = joined
				m.ZonesTable.SetRows(rows)
			}
		}
		m.popup.IsActive = false
		m.showPopup = false
		m.overlay = nil
		return m, nil
	case popup.CancelMsg:
        // Cancel without persisting changes
		m.popup.IsActive = false
	}

	if m.loading {
		m.spinner, cmd = m.spinner.Update(msg)
	}

    // If the message type does not match any of the handled cases, return unchanged with no command
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
	menuStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Width(m.width).
		Height(menuHeight)

	menu := []string{
		"[↑/↓/←/→] Navigate",
		"[Enter] Edit",
		"[Esc] Exit edit",
		"[r] Reload",
		"[e] Edit",
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
		return statusStyle.Render(fmt.Sprintf("Loading %s", m.spinner.View()))
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
	var rrset []models.DNSRecord

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
			boolToCheckMark(rr.Proxied),
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

// updateRRSet updates resource records set for the given zone name.
func (m *Model) updateRRSet(zone string) tea.Cmd {
	return func() tea.Msg {
		a, _ := app.New()
		ctx, cancel := context.WithTimeout(context.Background(), m.ClientTimeout)
		defer cancel()

		rrset, _ := a.Provider().ListRecords(ctx, models.ListDNSRecordsParams{
			ZoneName: zone,
		})
		m.rrsetCache[zone] = rrset

        // Return a message to indicate that data has been loaded
		return dataLoadedMsg{}
	}
}

// switchTable switches focus between zones and records tables
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

func (m *Model) updateTableRow(index int, newRow table.Row) {
	var rrset []models.DNSRecord

	rows := m.current.Rows()
	if index >= 0 && index < len(rows) {
        rows[index] = newRow
		m.current.SetRows(rows) // Переназначаем строки таблице
		// cache update
		selectedRow := m.ZonesTable.SelectedRow()
		if len(selectedRow) > 0 {
			if _, ok := m.rrsetCache[selectedRow[0]]; ok {
				rrset = m.rrsetCache[selectedRow[0]]
			}
            for i := range rrset {
                if rrset[i].Name == newRow[0] {
                    rrset[i].TTL, _ = strconv.Atoi(newRow[1])
                    rrset[i].Type = newRow[2]
                    // newRow[3] is "true"/"false"; convert to bool
                    rrset[i].Proxied = strings.ToLower(newRow[3]) == "true"
                    rrset[i].Content = newRow[4]
                    break
                }
            }
			m.rrsetCache[selectedRow[0]] = rrset
		}
	}
}

// renderBase renders base UI without overlays
func (m *Model) renderBase(table string) string {
    return m.ViewStyle.Render(
        lipgloss.JoinVertical(lipgloss.Left,
            m.viewHeader(),
            table,
            m.viewStatusBar(),
            m.viewMenu(),
        ),
    )
}

// backgroundViewModel adapts base view to tea.Model for overlay background
type backgroundViewModel struct{ parent *Model }

func (b *backgroundViewModel) Init() tea.Cmd                 { return nil }
func (b *backgroundViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return b, nil }
func (b *backgroundViewModel) View() string {
    table := b.parent.viewZones()
    if b.parent.RRSetTable.Focused() {
        table = b.parent.viewRRSet()
    }
    return b.parent.renderBase(table)
}

// updateRRFromFields builds DNSRecord and performs UpdateRR via provider
func (m *Model) updateRRFromFields(fields []string) tea.Cmd {
    return func() tea.Msg {
        a, _ := app.New()
        ctx, cancel := context.WithTimeout(context.Background(), m.ClientTimeout)
        defer cancel()

        // Find selected zone and record by name
        zoneRow := m.ZonesTable.SelectedRow()
        if len(zoneRow) == 0 {
            return nil
        }
        zoneName := zoneRow[0]
        var target models.DNSRecord
        if rrset, ok := m.rrsetCache[zoneName]; ok {
            for _, r := range rrset {
                if r.Name == fields[0] {
                    target = r
                    break
                }
            }
        }

        ttl, _ := strconv.Atoi(fields[1])
        proxied := strings.ToLower(fields[3]) == "true"

        // Build updated record
        target.Name = fields[0]
        target.TTL = ttl
        target.Type = fields[2]
        target.Proxied = proxied
        target.Content = fields[4]

        // Perform update
        if _, err := a.Provider().UpdateRR(ctx, zoneName, target); err != nil {
            // keep UI responsive; could add error handling UI later
            return nil
        }

        // Close popup
        m.popup.IsActive = false
        m.showPopup = false
        m.overlay = nil
        return nil
    }
}
