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

// Package popup holds cfdnscli UI elements for editing.
package popup

import (
	"fmt"
    "net"
    "regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
    "github.com/mixanemca/cfdnscli/internal/ui/theme"
    overlay "github.com/rmhubbert/bubbletea-overlay"
)

// SaveActionMsg is a tea.Msg signaling that edited fields should be saved.
type SaveActionMsg struct {
	Fields []string // Поля, которые были отредактированы
}

// CancelMsg is a tea.Msg signaling that editing was canceled.
type CancelMsg struct{}
// SaveNameServersMsg is a tea.Msg signaling that NS list should be saved.
type SaveNameServersMsg struct {
    Servers []string
}
// ConfirmDeleteMsg is a tea.Msg signaling that deletion was confirmed.
type ConfirmDeleteMsg struct{}
// Model implements tea.Model and represents the popup editor state.
type Model struct {
	ColumnNames []string               // Названия столбцов
	Fields      []string               // Поля для редактирования
	Cursor      int                    // Текущий индекс поля
	CharPos     int                    // Позиция курсора в текущем поле
	IsActive    bool                   // Активно ли окно
	Title       string                 // Заголовок окна
	SaveAction  func([]string) tea.Msg // Действие при сохранении
	CancelMsg   tea.Msg                // Сообщение при отмене

    // Boolean selection mode for boolean fields
    inBoolSelect bool
    boolIndex    int // 0 => true, 1 => false

    ov *overlay.Model

    // Text edit mode for simple fields (Name, TTL, Content)
    inTextEdit bool
    textBuf    string
    textErr    string
    // Type selection mode (enum)
    inTypeSelect bool
    typeIndex    int

    // Mode-specific state: NameServers list editor
    Mode        string   // "default" | "nslist" | "confirm"
    ListValues  []string // values for list mode
    ListCursor  int      // cursor for list mode
    // Confirm dialog state
    ConfirmIndex int    // 0 => Yes, 1 => No
}

// Ensure that model fulfils the tea.Model interface at compile time.
// var _ tea.Model = (*Model)(nil)

// Styles via lipgloss
var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF79C6")).MarginBottom(1)
	// columnStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9")).PaddingBottom(1)
	columnStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9"))
	fieldStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
	// cursorStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5555"))
	helpTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).MarginTop(1)
	borderStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).Margin(1).BorderForeground(lipgloss.Color("#BD93F9"))
	// highlightedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F1FA8C"))
)

// Styles for boolean selection modal
var (
    boolTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(theme.Color.Highlight)
    boolNormalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#A0A0A0"))
    boolSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(theme.Color.Highlight)
    boolModalBorder   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(theme.Color.Border).Padding(1).Margin(1)
)

// Menu help displayed in the editor
var menu = []string{
    "[↑/↓/←/→] Navigate",
    "[Enter] Edit field",
    "[Ctrl+S] Save",
    "[Esc] Exit edit / cancel selection",
}

// New constructs a popup editor Model.
func New(columnNames []string, fields []string, title string, saveAction func([]string) tea.Msg, cancelMsg tea.Msg) *Model {
    return &Model{
		ColumnNames: columnNames,
		Fields:      fields,
		Cursor:      0,
		CharPos:     0,
		IsActive:    true,
		Title:       title,
		SaveAction:  saveAction,
		CancelMsg:   cancelMsg,
	}
}

// NewNameServersEditor constructs popup Model in list-edit mode for NS values.
func NewNameServersEditor(initial []string, title string) *Model {
    // ensure at least 2 lines
    vals := make([]string, len(initial))
    copy(vals, initial)
    for len(vals) < 2 {
        vals = append(vals, "")
    }
    return &Model{
        Mode:       "nslist",
        ListValues: vals,
        ListCursor: 0,
        IsActive:   true,
        Title:      title,
    }
}

// NewConfirmDialog constructs popup Model in confirmation mode.
func NewConfirmDialog(title string) *Model {
    return &Model{
        Mode:         "confirm",
        ConfirmIndex: 0, // default to Yes
        IsActive:     true,
        Title:        title,
    }
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if !m.IsActive {
        return m, nil
    }

    // List editor base behavior (multi-line NameServers editor)
    if m.Mode == "nslist" {
        // Handle inline text editing overlay first
        if m.inTextEdit {
            switch km := msg.(type) {
            case tea.KeyMsg:
                switch km.Type {
                case tea.KeyEnter:
                    value := strings.TrimSpace(m.textBuf)
                    if value != "" && !isHostname(value) {
                        m.textErr = "Name server must be a valid hostname"
                        return m, nil
                    }
                    if value == "" && len(m.ListValues) > 2 {
                        // delete line when confirming empty value and we have more than minimum
                        idx := m.ListCursor
                        m.ListValues = append(m.ListValues[:idx], m.ListValues[idx+1:]...)
                        if m.ListCursor >= len(m.ListValues) && m.ListCursor > 0 { m.ListCursor-- }
                        // ensure at least 2 lines remain
                        for len(m.ListValues) < 2 { m.ListValues = append(m.ListValues, "") }
                    } else {
                        m.ListValues[m.ListCursor] = value
                    }
                    m.inTextEdit = false
                    m.textBuf = ""
                    m.textErr = ""
                    m.ov = nil
                case tea.KeyEsc:
                    m.inTextEdit = false
                    m.textBuf = ""
                    m.textErr = ""
                    m.ov = nil
                case tea.KeyBackspace:
                    if len(m.textBuf) > 0 { m.textBuf = m.textBuf[:len(m.textBuf)-1] }
                case tea.KeyCtrlH:
                    if len(m.textBuf) > 0 { m.textBuf = m.textBuf[:len(m.textBuf)-1] }
                case tea.KeyRunes:
                    if len(km.Runes) > 0 { m.textBuf += string(km.Runes) }
                }
            }
            return m, nil
        }

        // Base navigation and commands for list mode
        switch km := msg.(type) {
        case tea.KeyMsg:
            switch km.Type {
            case tea.KeyUp:
                if m.ListCursor > 0 { m.ListCursor-- }
                return m, nil
            case tea.KeyDown:
                if m.ListCursor == len(m.ListValues)-1 {
                    // add a new line when moving past the last, but cap at 4
                    if len(m.ListValues) < 4 {
                        m.ListValues = append(m.ListValues, "")
                        m.ListCursor++
                        return m, nil
                    }
                    // already at max; keep cursor at last
                    return m, nil
                }
                m.ListCursor++
                return m, nil
            case tea.KeyEnter:
                // open text editor for current line
                m.inTextEdit = true
                m.textBuf = m.ListValues[m.ListCursor]
                m.textErr = ""
                return m, nil
            case tea.KeyCtrlD:
                // delete current line if more than 2 lines remain
                if len(m.ListValues) > 2 {
                    idx := m.ListCursor
                    m.ListValues = append(m.ListValues[:idx], m.ListValues[idx+1:]...)
                    if m.ListCursor >= len(m.ListValues) && m.ListCursor > 0 { m.ListCursor-- }
                    // ensure at least 2 lines remain
                    for len(m.ListValues) < 2 { m.ListValues = append(m.ListValues, "") }
                }
                return m, nil
            case tea.KeyCtrlS:
                // save non-empty trimmed values
                var out []string
                for _, v := range m.ListValues {
                    v = strings.TrimSpace(v)
                    if v != "" { out = append(out, v) }
                }
                m.IsActive = false
                return m, func() tea.Msg { return SaveNameServersMsg{Servers: out} }
            case tea.KeyEsc:
                m.IsActive = false
                return m, func() tea.Msg { return CancelMsg{} }
            }
        }
        return m, nil
    }

    // Confirm dialog mode
    if m.Mode == "confirm" {
        switch km := msg.(type) {
        case tea.KeyMsg:
            switch km.Type {
            case tea.KeyLeft, tea.KeyUp:
                if m.ConfirmIndex > 0 {
                    m.ConfirmIndex = 0
                }
            case tea.KeyRight, tea.KeyDown:
                if m.ConfirmIndex < 1 {
                    m.ConfirmIndex = 1
                }
            case tea.KeyEnter:
                // Confirm selection
                if m.ConfirmIndex == 0 {
                    // Yes - confirm deletion
                    m.IsActive = false
                    return m, func() tea.Msg { return ConfirmDeleteMsg{} }
                }
                // No - cancel
                m.IsActive = false
                return m, func() tea.Msg { return CancelMsg{} }
            case tea.KeyEsc:
                // Cancel
                m.IsActive = false
                return m, func() tea.Msg { return CancelMsg{} }
            }
        }
        return m, nil
    }

    // Handle text edit mode
    if m.inTextEdit {
        switch km := msg.(type) {
        case tea.KeyMsg:
            switch km.Type {
            case tea.KeyEnter:
                fieldName := strings.ToLower(m.ColumnNames[m.Cursor])
                rrType := strings.ToUpper(m.currentType())
                if errText := validateInput(fieldName, m.textBuf, rrType); errText != "" {
                    m.textErr = errText
                    return m, nil
                }
                m.Fields[m.Cursor] = m.textBuf
                m.inTextEdit = false
                m.textBuf = ""
                m.textErr = ""
                m.ov = nil
                m.CharPos = len(m.Fields[m.Cursor])
            case tea.KeyEsc:
                m.inTextEdit = false
                m.textBuf = ""
                m.textErr = ""
                m.ov = nil
            case tea.KeyLeft, tea.KeyRight, tea.KeyUp, tea.KeyDown:
                // ignore navigation keys in text edit modal (no cursor rendering yet)
            case tea.KeyBackspace:
                if len(m.textBuf) > 0 {
                    m.textBuf = m.textBuf[:len(m.textBuf)-1]
                }
            case tea.KeyCtrlH:
                if len(m.textBuf) > 0 {
                    m.textBuf = m.textBuf[:len(m.textBuf)-1]
                }
            case tea.KeyDelete:
                // ignore for simplicity
            case tea.KeyRunes:
                if len(km.Runes) > 0 {
                    m.textBuf += string(km.Runes)
                }
            }
        }
        return m, nil
    }

    // Handle type selection mode
    if m.inTypeSelect {
        switch km := msg.(type) {
        case tea.KeyMsg:
            switch km.Type {
            case tea.KeyUp:
                if m.typeIndex > 0 {
                    m.typeIndex--
                }
            case tea.KeyDown:
                if m.typeIndex < len(supportedTypes)-1 {
                    m.typeIndex++
                }
            case tea.KeyEnter:
                m.Fields[m.Cursor] = supportedTypes[m.typeIndex]
                m.inTypeSelect = false
                m.ov = nil
                m.CharPos = len(m.Fields[m.Cursor])
            case tea.KeyEsc:
                m.inTypeSelect = false
                m.ov = nil
            }
        }
        return m, nil
    }

    // Handle boolean selection mode
    if m.inBoolSelect && m.isBoolField(m.Cursor) {
        switch km := msg.(type) {
        case tea.KeyMsg:
            switch km.Type {
            case tea.KeyLeft, tea.KeyUp:
                if m.boolIndex > 0 {
                    m.boolIndex = 0
                }
            case tea.KeyRight, tea.KeyDown:
                if m.boolIndex < 1 {
                    m.boolIndex = 1
                }
            case tea.KeyEnter:
                // Apply selection to field and exit selection mode
                if m.boolIndex == 0 {
                    m.Fields[m.Cursor] = "true"
                } else {
                    m.Fields[m.Cursor] = "false"
                }
                m.inBoolSelect = false
                m.ov = nil
                // keep cursor position at end
                m.CharPos = len(m.Fields[m.Cursor])
            case tea.KeyEsc:
                // cancel selection without changes
                m.inBoolSelect = false
                m.ov = nil
            }
        }
        return m, nil
    }
    

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab, tea.KeyDown: // Перемещение вперёд по полям
			m.Cursor = (m.Cursor + 1) % len(m.Fields)
			m.CharPos = len(m.Fields[m.Cursor])
		case tea.KeyShiftTab, tea.KeyUp: // Перемещение назад
			m.Cursor = (m.Cursor - 1 + len(m.Fields)) % len(m.Fields)
			m.CharPos = len(m.Fields[m.Cursor]) // Переместить курсор в конец нового поля
		case tea.KeyLeft: // Перемещение курсора влево
			if m.CharPos > 0 {
				m.CharPos--
			}
		case tea.KeyRight: // Перемещение курсора вправо
			if m.CharPos < len(m.Fields[m.Cursor]) {
				m.CharPos++
			}
        case tea.KeyEnter:
            // Открыть редактор поля
            if m.isBoolField(m.Cursor) {
                m.inBoolSelect = true
                if strings.ToLower(m.Fields[m.Cursor]) == "true" { m.boolIndex = 0 } else { m.boolIndex = 1 }
                return m, nil
            }
            if m.isTypeField(m.Cursor) {
                m.inTypeSelect = true
                // init index to current value
                m.typeIndex = 0
                cur := strings.ToUpper(m.Fields[m.Cursor])
                for i, t := range supportedTypes {
                    if t == cur { m.typeIndex = i; break }
                }
                return m, nil
            }
            // текстовые поля
            m.inTextEdit = true
            m.textBuf = m.Fields[m.Cursor]
            return m, nil
		case tea.KeyCtrlS: // сохранить все изменения формы
			m.IsActive = false
			return m, func() tea.Msg { return SaveActionMsg{Fields: m.Fields} }
		case tea.KeyEsc: // выйти без сохранения
			m.IsActive = false
			return m, func() tea.Msg { return CancelMsg{} }
		default:
			// В базовом режиме до входа в редактирование игнорируем любые не-навигационные клавиши
			return m, nil
		}
	}

    return m, nil
}

// View implements tea.Model.
func (m *Model) View() string {
	if !m.IsActive {
		return ""
	}

    // List editor view
    if m.Mode == "nslist" {
        if m.inTextEdit {
            if m.ov == nil {
                bg := &baseListView{parent: m}
                fg := &textView{parent: m}
                m.ov = overlay.New(fg, bg, overlay.Center, overlay.Center, 0, 0)
            }
            return m.ov.View()
        }
        return m.viewListBase()
    }

    // Confirm dialog view
    if m.Mode == "confirm" {
        return m.viewConfirm()
    }

    // When in boolean selection, show a small overlay modal over the edit window
    if m.inBoolSelect && m.isBoolField(m.Cursor) {
        if m.ov == nil {
            bg := &baseView{parent: m}
            fg := &boolView{parent: m}
            m.ov = overlay.New(fg, bg, overlay.Center, overlay.Center, 0, 0)
        }
        return m.ov.View()
    }

    // When in text edit, show small input overlay
    if m.inTextEdit {
        if m.ov == nil {
            bg := &baseView{parent: m}
            fg := &textView{parent: m}
            m.ov = overlay.New(fg, bg, overlay.Center, overlay.Center, 0, 0)
        }
        return m.ov.View()
    }

    // When in type select, show enum overlay
    if m.inTypeSelect {
        if m.ov == nil {
            bg := &baseView{parent: m}
            fg := &typeView{parent: m}
            m.ov = overlay.New(fg, bg, overlay.Center, overlay.Center, 0, 0)
        }
        return m.ov.View()
    }

    return m.viewBase()
}

// isBoolField returns true if field at index i is boolean-typed (by column name).
func (m *Model) isBoolField(i int) bool {
    if i < 0 || i >= len(m.ColumnNames) {
        return false
    }
    name := strings.ToLower(m.ColumnNames[i])
    return name == "proxied" || name == "enabled" || name == "active"
}

// isTypeField returns true if field is the RR type.
func (m *Model) isTypeField(i int) bool {
    if i < 0 || i >= len(m.ColumnNames) {
        return false
    }
    return strings.ToLower(m.ColumnNames[i]) == "type"
}

var supportedTypes = []string{"A", "AAAA", "CNAME", "TXT", "MX", "NS", "SRV", "CAA"}

// currentType returns current RR type from fields
func (m *Model) currentType() string {
    for idx, name := range m.ColumnNames {
        if strings.ToLower(name) == "type" && idx < len(m.Fields) {
            return m.Fields[idx]
        }
    }
    return ""
}

// validateInput validates value based on field and rr type
func validateInput(fieldName, value, rrType string) string {
    switch fieldName {
    case "ttl":
        if !isNumber(value) {
            return "TTL must be a number in seconds (e.g. 60, 300, 1800)"
        }
        return ""
    case "name":
        if !isHostname(value) {
            return "Name must be a valid hostname"
        }
        return ""
    case "content":
        switch strings.ToUpper(rrType) {
        case "A":
            if !isIPv4(value) { return "Content must be a valid IPv4 address for A record" }
        case "AAAA":
            if !isIPv6(value) { return "Content must be a valid IPv6 address for AAAA record" }
        case "CNAME", "NS", "MX":
            if !isHostname(value) { return "Content must be a valid hostname" }
        default:
            // TXT, SRV, CAA — skip strict validation
            return ""
        }
        return ""
    default:
        return ""
    }
}

var hostnameRe = regexp.MustCompile(`^(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)(?:\.(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?))*\.?$`)

func isHostname(s string) bool {
    if len(s) == 0 || len(s) > 253 { return false }
    return hostnameRe.MatchString(s)
}

func isNumber(s string) bool {
    for _, r := range s {
        if r < '0' || r > '9' { return false }
    }
    return len(s) > 0
}

func isIPv4(s string) bool {
    ip := net.ParseIP(s)
    return ip != nil && ip.To4() != nil
}

func isIPv6(s string) bool {
    ip := net.ParseIP(s)
    return ip != nil && ip.To16() != nil && ip.To4() == nil
}

// stringWidth calculates display width of a string (runes length is enough here)
func stringWidth(s string) int {
    return len([]rune(s))
}

// viewBase renders the main edit window content
func (m *Model) viewBase() string {
    // Построим стилизованные строки и параллельно посчитаем ширину по "сырым" строкам
    var fieldLinesStyled []string
    maxW := stringWidth(fmt.Sprintf("--- %s ---", m.Title))
    for i, field := range m.Fields {
        columnName := m.ColumnNames[i]
        raw := fmt.Sprintf(" > %s: %s", columnName, field)
        if w := stringWidth(raw); w > maxW { maxW = w }
        if i == m.Cursor {
            fieldLinesStyled = append(fieldLinesStyled, fmt.Sprintf(" > %s: %s", columnStyle.Render(columnName), fieldStyle.Render(field)))
        } else {
            fieldLinesStyled = append(fieldLinesStyled, fmt.Sprintf("   %s: %s", columnStyle.Render(columnName), fieldStyle.Render(field)))
        }
    }

    helpPlain := strings.Join(menu, " | ")
    if w := stringWidth(helpPlain); w > maxW { maxW = w }
    if maxW < 30 { maxW = 30 }

    // Заголовок по центру относительно рассчитанной ширины
    header := lipgloss.Place(
        maxW,
        1,
        lipgloss.Center,
        lipgloss.Top,
        titleStyle.Render(fmt.Sprintf("--- %s ---", m.Title)),
    )

    helpStyled := helpTextStyle.Render(helpPlain)
    content := lipgloss.JoinVertical(lipgloss.Top, header, strings.Join(fieldLinesStyled, "\n"), helpStyled)

    boxed := borderStyle.Render(content)
    // Центрируем всю коробку в пределах рассчитанной ширины
    height := 1 + len(fieldLinesStyled) + 1 // header + fields + help
    return lipgloss.Place(maxW, height, lipgloss.Center, lipgloss.Top, boxed)
}

// baseView adapts base editor view to tea.Model for overlay background
type baseView struct{ parent *Model }

func (b *baseView) Init() tea.Cmd                                  { return nil }
func (b *baseView) Update(msg tea.Msg) (tea.Model, tea.Cmd)        { return b, nil }
func (b *baseView) View() string                                   { return b.parent.viewBase() }

// baseListView adapts list editor base view as overlay background
type baseListView struct{ parent *Model }

func (b *baseListView) Init() tea.Cmd                           { return nil }
func (b *baseListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return b, nil }
func (b *baseListView) View() string                            { return b.parent.viewListBase() }

// boolView renders the boolean selection small modal as a tea.Model foreground
type boolView struct{ parent *Model }

func (b *boolView) Init() tea.Cmd                           { return nil }
func (b *boolView) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return b, nil }
func (b *boolView) View() string {
    windowWidth := 20
    header := lipgloss.Place(
        windowWidth,
        1,
        lipgloss.Center,
        lipgloss.Top,
        boolTitleStyle.Render("Select value"),
    )

    trueLine := boolNormalStyle.Render("true")
    falseLine := boolNormalStyle.Render("false")
    if b.parent.boolIndex == 0 {
        trueLine = boolSelectedStyle.Render("true")
    } else {
        falseLine = boolSelectedStyle.Render("false")
    }
    body := lipgloss.JoinVertical(lipgloss.Top, header, trueLine, falseLine, helpTextStyle.Render("[↑/↓] Move  [Enter] Apply  [Esc] Cancel"))
    return boolModalBorder.Render(body)
}

// textView renders a simple input modal
type textView struct{ parent *Model }

func (t *textView) Init() tea.Cmd                           { return nil }
func (t *textView) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return t, nil }
func (t *textView) View() string {
    windowWidth := 40
    header := lipgloss.Place(
        windowWidth,
        1,
        lipgloss.Center,
        lipgloss.Top,
        boolTitleStyle.Render("Edit value"),
    )
    value := fieldStyle.Render(t.parent.textBuf)
    // Build hint safely for both default and nslist modes
    var hintText string
    if t.parent.Mode == "nslist" {
        hintText = "Nameserver hostname, e.g. ns1.example.com"
    } else {
        colName := ""
        if t.parent.Cursor >= 0 && t.parent.Cursor < len(t.parent.ColumnNames) {
            colName = strings.ToLower(t.parent.ColumnNames[t.parent.Cursor])
        }
        hintText = textHint(colName, strings.ToUpper(t.parent.currentType()))
    }
    hint := helpTextStyle.Render(hintText)
    errLine := ""
    if t.parent.textErr != "" {
        errStyle := lipgloss.NewStyle().Foreground(theme.Color.Red)
        errLine = errStyle.Render(t.parent.textErr)
    }
    help := helpTextStyle.Render("[Enter] Apply  [Esc] Cancel")
    body := lipgloss.JoinVertical(lipgloss.Top, header, value, hint, errLine, help)
    return boolModalBorder.Render(body)
}

// textHint returns context-aware hint for field
func textHint(fieldName, rrType string) string {
    switch fieldName {
    case "ttl":
        return "TTL in seconds, e.g. 60, 300, 1800"
    case "name":
        return "Record name (hostname), e.g. www or api.example.com"
    case "content":
        switch rrType {
        case "A":
            return "IPv4 address, e.g. 203.0.113.10"
        case "AAAA":
            return "IPv6 address, e.g. 2001:db8::1"
        case "CNAME":
            return "Canonical hostname, e.g. target.example.com"
        case "MX":
            return "Mail exchanger hostname, e.g. mail.example.com"
        case "NS":
            return "Nameserver hostname, e.g. ns1.example.com"
        default:
            return "Value for record content"
        }
    default:
        return ""
    }
}

// typeView renders a selection list for RR types
type typeView struct{ parent *Model }

func (t *typeView) Init() tea.Cmd                           { return nil }
func (t *typeView) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return t, nil }
func (t *typeView) View() string {
    windowWidth := 20
    header := lipgloss.Place(
        windowWidth,
        1,
        lipgloss.Center,
        lipgloss.Top,
        boolTitleStyle.Render("Select type"),
    )
    var lines []string
    for i, tp := range supportedTypes {
        line := boolNormalStyle.Render(tp)
        if i == t.parent.typeIndex {
            line = boolSelectedStyle.Render(tp)
        }
        lines = append(lines, line)
    }
    list := lipgloss.JoinVertical(lipgloss.Top, lines...)
    help := helpTextStyle.Render("[↑/↓] Move  [Enter] Apply  [Esc] Cancel")
    body := lipgloss.JoinVertical(lipgloss.Top, header, list, help)
    return boolModalBorder.Render(body)
}

// viewListBase renders multi-line NS editor
func (m *Model) viewListBase() string {
    var lines []string
    maxW := stringWidth(fmt.Sprintf("--- %s ---", m.Title))
    for i, v := range m.ListValues {
        prefix := "   "
        if i == m.ListCursor { prefix = " > " }
        raw := fmt.Sprintf("%sns%d: %s", prefix, i+1, v)
        if w := stringWidth(raw); w > maxW { maxW = w }
        if i == m.ListCursor {
            lines = append(lines, fmt.Sprintf(" %s", fieldStyle.Render(raw)))
        } else {
            lines = append(lines, raw)
        }
    }

    helpPlain := strings.Join([]string{"[↑/↓] Move", "[Enter] Edit", "[Ctrl+D] Delete", "[Ctrl+S] Save", "[Esc] Cancel", "(max 4 NS)"}, " | ")
    if w := stringWidth(helpPlain); w > maxW { maxW = w }
    if maxW < 30 { maxW = 30 }

    header := lipgloss.Place(maxW, 1, lipgloss.Center, lipgloss.Top, titleStyle.Render(fmt.Sprintf("--- %s ---", m.Title)))
    helpStyled := helpTextStyle.Render(helpPlain)
    content := lipgloss.JoinVertical(lipgloss.Top, header, strings.Join(lines, "\n"), helpStyled)
    boxed := borderStyle.Render(content)
    height := 1 + len(lines) + 1
    return lipgloss.Place(maxW, height, lipgloss.Center, lipgloss.Top, boxed)
}

// viewConfirm renders confirmation dialog
func (m *Model) viewConfirm() string {
    windowWidth := 40
    header := lipgloss.Place(
        windowWidth,
        1,
        lipgloss.Center,
        lipgloss.Top,
        titleStyle.Render(fmt.Sprintf("--- %s ---", m.Title)),
    )

    yesLine := boolNormalStyle.Render("Yes")
    noLine := boolNormalStyle.Render("No")
    if m.ConfirmIndex == 0 {
        yesLine = boolSelectedStyle.Render("Yes")
    } else {
        noLine = boolSelectedStyle.Render("No")
    }
    body := lipgloss.JoinVertical(lipgloss.Top, header, yesLine, noLine, helpTextStyle.Render("[←/→] Move  [Enter] Confirm  [Esc] Cancel"))
    return boolModalBorder.Render(body)
}
