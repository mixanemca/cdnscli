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

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cloudflare/cloudflare-go"
	"github.com/mixanemca/cfdnscli/app"
)

// Ensure that model fulfils the tea.Model interface at compile time.
var _ tea.Model = (*Model)(nil)

// Custom tea.Msg to switching between zones and rrset tables.
type switchTableToRRSetCmd string

// Model represents model for implements bubbletea.Model interface
type Model struct {
	width      int
	height     int
	current    *table.Model
	rrsetCache map[string][]cloudflare.DNSRecord

	ClientTimeout time.Duration

	ZonesTable table.Model
	RRSetTable table.Model
	TableStyle table.Styles
	BaseStyle  lipgloss.Style
	ViewStyle  lipgloss.Style
}

type Theme struct {
	Primary   lipgloss.AdaptiveColor
	Secondary lipgloss.AdaptiveColor
	Highlight lipgloss.AdaptiveColor
	Border    lipgloss.AdaptiveColor
	Green     lipgloss.AdaptiveColor
	Red       lipgloss.AdaptiveColor
}

var Color = Theme{
	Primary:   lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"},
	Secondary: lipgloss.AdaptiveColor{Light: "#969B86", Dark: "#696969"},
	Highlight: lipgloss.AdaptiveColor{Light: "#8b2def", Dark: "#8b2def"},
	Border:    lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"},
	Green:     lipgloss.AdaptiveColor{Light: "#00FF00", Dark: "#00FF00"},
	Red:       lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF0000"},
}

func NewModel() *Model {
	var m Model

	m.rrsetCache = make(map[string][]cloudflare.DNSRecord)
	m.BaseStyle = lipgloss.NewStyle()
	m.ViewStyle = lipgloss.NewStyle()

	return &m
}

// This command will be executed immediately when the program starts.
func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		a, _ := app.New()
		ctx, cancel := context.WithTimeout(context.Background(), m.ClientTimeout)
		defer cancel()

		zones, _ := a.Zones().List(ctx)
		rows := []table.Row{}
		for _, zone := range zones {
			rows = append(rows, table.Row{
				zone.Name,
				strings.Join(zone.NameServers, ", "),
			})
		}
		m.ZonesTable.SetRows(rows)

		return m
	}
}

func (m *Model) View() string {
	m.current = &m.ZonesTable
	table := m.viewZones()
	if m.RRSetTable.Focused() {
		m.current = &m.RRSetTable
		table = m.viewRRSet()
	}
	// Sets the width of the column to the width of the terminal (m.width) and adds padding of 1 unit on the top.
	// Render is a method from the lipgloss package that applies the defined style and returns a function that can render styled content.
	column := m.BaseStyle.Width(m.width).Padding(1, 0, 0, 0).Render
	// Set the content to match the terminal dimensions (m.width and m.height).
	content := m.BaseStyle.
		Width(m.width).
		Height(m.height).
		Render(
			// Vertically join multiple elements aligned to the left.
			lipgloss.JoinVertical(lipgloss.Left,
				column(m.viewHeader()),
				column(table),
			),
			lipgloss.JoinVertical(lipgloss.Right,
				column(m.viewMenu()),
			),
		)

	return content
}

// Takes a tea.Msg as input and uses a type switch to handle different types of messages.
// Each case in the switch statement corresponds to a specific message type.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.switchTable("Zones")
		// Moves the focus up in the process table if the table is focused.
		case "up", "k":
			m.current.MoveUp(1)
		// Moves the focus down in the process table if the table is focused.
		case "down", "j":
			m.current.MoveDown(1)
		// Quits the program by returning the tea.Quit command.
		case "q", "ctrl+c":
			return m, tea.Quit
		}
		switch msg.Type {
		case tea.KeyEnter, tea.KeySpace:
			return m, m.handleEnter(msg)
		}

	// Custom messages
	case switchTableToRRSetCmd:
		m.switchTable("RRSet")
		selectedRow := m.ZonesTable.SelectedRow()
		if len(selectedRow) > 0 {
			if _, ok := m.rrsetCache[selectedRow[0]]; !ok {
				m.updateRRSet(selectedRow[0])
			}
		}
	}
	// If the message type does not match any of the handled cases, the model is returned unchanged, and no new command is issued.
	return m, nil
}

// Uses lipgloss.JoinVertical and lipgloss.JoinHorizontal to arrange the header content.
func (m *Model) viewHeader() string {
	return m.ViewStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top,
			fmt.Sprintf("CloudFlare DNS CLI\n"),
		),
	)
}

func (m *Model) viewMenu() string {
	return fmt.Sprintf("Press Enter to select, Esc to return, arrow keys to move, / to find, Ctrl+C or q to exit")
}

func (m *Model) viewZones() string {
	return m.ViewStyle.Render(m.ZonesTable.View())
}

func (m *Model) viewRRSet() string {
	return m.ViewStyle.Render(m.RRSetTable.View())
}

func (m *Model) handleEnter(tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return switchTableToRRSetCmd("RRSet")
	}
}

// updateRRSet updates resource records set by given zone name.
func (m *Model) updateRRSet(zone string) {
	a, _ := app.New()
	ctx, cancel := context.WithTimeout(context.Background(), m.ClientTimeout)
	defer cancel()

	rrset, _ := a.Zones().ListRecordsByZoneName(ctx, zone, cloudflare.ListDNSRecordsParams{})
	m.rrsetCache[zone] = rrset
	rows := []table.Row{}
	for _, rr := range rrset {
		rows = append(rows, table.Row{
			rr.Name,
			strconv.Itoa(rr.TTL),
			rr.Type,
			rr.Content,
		})
	}
	m.RRSetTable.SetRows(rows)
}

// switchTable switches focus between zones and rrset tables
func (m *Model) switchTable(name string) {
	switch name {
	case "Zones":
		m.ZonesTable.Focus()
		m.RRSetTable.Blur()
		m.current = &m.ZonesTable
	case "RRSet":
		m.ZonesTable.Blur()
		m.RRSetTable.Focus()
		m.current = &m.RRSetTable
	}
}
