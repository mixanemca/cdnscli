/*
Copyright © 2024-2026 Michael Bruskov <mixanemca@yandex.ru>

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
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var recordColumns = []string{"Name", "TTL", "Type", "Proxied", "Content"}

func newRecordEditor(fields []string) *Model {
	return New(recordColumns, fields, "test", nil, CancelMsg{})
}

func press(m *Model, keys ...tea.KeyType) tea.Cmd {
	var cmd tea.Cmd
	for _, k := range keys {
		_, cmd = m.Update(tea.KeyMsg{Type: k})
	}
	return cmd
}

func typeText(m *Model, text string) {
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(text)})
}

// replaceCurrentField opens the editor for the field under the cursor and
// replaces its content with text, leaving the editor open.
func replaceCurrentField(m *Model, text string) {
	press(m, tea.KeyEnter)
	for range m.textBuf {
		press(m, tea.KeyBackspace)
	}
	typeText(m, text)
}

func TestTextEdit_CtrlS_AppliesValueAndSaves(t *testing.T) {
	m := newRecordEditor([]string{"foo.example.com", "3600", "A", "false", "192.0.2.1"})
	m.Cursor = 1 // TTL
	replaceCurrentField(m, "300")

	cmd := press(m, tea.KeyCtrlS)
	require.NotNil(t, cmd, "Ctrl+S inside the field editor must save the form")

	msg, ok := cmd().(SaveActionMsg)
	require.True(t, ok)
	assert.Equal(t, []string{"foo.example.com", "300", "A", "false", "192.0.2.1"}, msg.Fields)
	assert.False(t, m.IsActive)
	assert.False(t, m.inTextEdit)
}

func TestTextEdit_CtrlS_InvalidValueKeepsEditorOpen(t *testing.T) {
	m := newRecordEditor([]string{"foo.example.com", "3600", "A", "false", "192.0.2.1"})
	m.Cursor = 1
	replaceCurrentField(m, "soon")

	cmd := press(m, tea.KeyCtrlS)
	assert.Nil(t, cmd)
	assert.True(t, m.inTextEdit)
	assert.True(t, m.IsActive)
	assert.Contains(t, m.textErr, "TTL must be a number")
	assert.Equal(t, "3600", m.Fields[1], "field must keep the old value")
}

func TestTextEdit_Enter_AcceptsAutoTTL(t *testing.T) {
	m := newRecordEditor([]string{"foo.example.com", "3600", "A", "false", "192.0.2.1"})
	m.Cursor = 1
	replaceCurrentField(m, "auto")

	press(m, tea.KeyEnter)
	assert.False(t, m.inTextEdit)
	assert.Equal(t, "Auto", m.Fields[1], "typed value is normalized to the display form")
}

func TestTextEdit_ProxiedRecordRejectsNumericTTL(t *testing.T) {
	m := newRecordEditor([]string{"foo.example.com", "Auto", "A", "true", "192.0.2.1"})
	m.Cursor = 1
	replaceCurrentField(m, "300")

	cmd := press(m, tea.KeyCtrlS)
	assert.Nil(t, cmd)
	assert.True(t, m.inTextEdit)
	assert.Contains(t, m.textErr, "Proxied records use TTL Auto")
	assert.Equal(t, "Auto", m.Fields[1])
}

func TestBoolSelect_EnablingProxiedSetsTTLAuto(t *testing.T) {
	m := newRecordEditor([]string{"foo.example.com", "3600", "A", "false", "192.0.2.1"})
	m.Cursor = 3 // Proxied
	press(m, tea.KeyEnter)
	require.True(t, m.inBoolSelect)

	press(m, tea.KeyUp, tea.KeyEnter) // select "true" and apply
	assert.Equal(t, "true", m.Fields[3])
	assert.Equal(t, "Auto", m.Fields[1])
	assert.False(t, m.inBoolSelect)
}

func TestBoolSelect_DisablingProxiedKeepsTTL(t *testing.T) {
	m := newRecordEditor([]string{"foo.example.com", "Auto", "A", "true", "192.0.2.1"})
	m.Cursor = 3
	press(m, tea.KeyEnter)
	press(m, tea.KeyDown, tea.KeyEnter) // select "false" and apply
	assert.Equal(t, "false", m.Fields[3])
	assert.Equal(t, "Auto", m.Fields[1])
}

func TestValidateInput_TTL(t *testing.T) {
	assert.Empty(t, validateInput("ttl", "300", "A"))
	assert.Empty(t, validateInput("ttl", "Auto", "A"))
	assert.Empty(t, validateInput("ttl", "AUTO", "A"))
	assert.NotEmpty(t, validateInput("ttl", "", "A"))
	assert.NotEmpty(t, validateInput("ttl", "5m", "A"))
}
