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

package popup

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SaveActionMsg struct {
	Fields []string // Поля, которые были отредактированы
}

type CancelMsg struct{}

type Model struct {
	ColumnNames []string               // Названия столбцов
	Fields      []string               // Поля для редактирования
	Cursor      int                    // Текущий индекс поля
	CharPos     int                    // Позиция курсора в текущем поле
	IsActive    bool                   // Активно ли окно
	Title       string                 // Заголовок окна
	SaveAction  func([]string) tea.Msg // Действие при сохранении
	CancelMsg   tea.Msg                // Сообщение при отмене
}

// Стилизация через lipgloss
var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF79C6")).MarginBottom(1)
	// columnStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9")).PaddingBottom(1)
	columnStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9"))
	fieldStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
	cursorStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5555"))
	helpTextStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).MarginTop(1)
	borderStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).Margin(1).BorderForeground(lipgloss.Color("#BD93F9"))
	highlightedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F1FA8C"))
)

var menu = []string{
	"[↑/↓/←/→] Navigate",
	"[Enter] Save",
	"[Esc] Exit edit",
}

func New(columnNames []string, fields []string, title string, saveAction func([]string) tea.Msg, cancelMsg tea.Msg) Model {
	return Model{
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

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.IsActive {
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
		case tea.KeyEnter: // Сохранение изменений
			m.IsActive = false
			return m, func() tea.Msg {
				// Сохранение данных и закрытие редактора
				return SaveActionMsg{Fields: m.Fields}
			}
		case tea.KeyEsc: // Отмена
			m.IsActive = false
			return m, func() tea.Msg {
				// Отмена редактирования и возвращение состояния
				return CancelMsg{}
			}
		case tea.KeyBackspace: // Удаление символа
			if m.CharPos > 0 {
				// Удалить символ перед курсором
				field := m.Fields[m.Cursor]
				m.Fields[m.Cursor] = field[:m.CharPos-1] + field[m.CharPos:]
				m.CharPos-- // Сдвинуть курсор влево
			}
		case tea.KeyDelete: // Удаление символа после курсора
			if m.CharPos < len(m.Fields[m.Cursor]) {
				field := m.Fields[m.Cursor]
				m.Fields[m.Cursor] = field[:m.CharPos] + field[m.CharPos+1:]
			}
		default:
			// Добавление символов в текущее поле
			field := m.Fields[m.Cursor]
			m.Fields[m.Cursor] = field[:m.CharPos] + msg.String() + field[m.CharPos:]
			m.CharPos++ // Сдвинуть курсор вправо
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.IsActive {
		return ""
	}

	// Заголовок окна
	// Центрирование заголовка
	windowWidth := 50 // Ширина окна, настраивается по вашему желанию
	header := lipgloss.Place(
		windowWidth,
		1,               // Высота строки
		lipgloss.Center, // Выравнивание по горизонтали
		lipgloss.Top,    // Выравнивание по вертикали
		titleStyle.Render(fmt.Sprintf("--- %s ---", m.Title)),
	)

	// Поля для редактирования
	var fields strings.Builder
	for i, field := range m.Fields {
		columnName := m.ColumnNames[i] // Название столбца для текущего поля
		if i == m.Cursor {
			// Отображение курсора в текущем поле
			fields.WriteString(fmt.Sprintf(" > %s: %s\n", columnStyle.Render(columnName), fieldStyle.Render(field)))
		} else {
			fields.WriteString(fmt.Sprintf("   %s: %s\n", columnStyle.Render(columnName), fieldStyle.Render(field)))
		}
	}

	// Подсказка
	help := helpTextStyle.Render(strings.Join(menu, " | "))

	// Итоговый вывод
	content := lipgloss.JoinVertical(lipgloss.Top, header, fields.String(), help)

	return borderStyle.Render(content)
}
