package tui

import (
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
)

const escByte = '\x1b'

var keyTypeToSequence = map[tea.KeyType][]byte{
	tea.KeyUp:       {escByte, '[', 'A'},
	tea.KeyDown:     {escByte, '[', 'B'},
	tea.KeyRight:    {escByte, '[', 'C'},
	tea.KeyLeft:     {escByte, '[', 'D'},
	tea.KeyHome:     {escByte, '[', 'H'},
	tea.KeyEnd:      {escByte, '[', 'F'},
	tea.KeyPgUp:     {escByte, '[', '5', '~'},
	tea.KeyPgDown:   {escByte, '[', '6', '~'},
	tea.KeyInsert:   {escByte, '[', '2', '~'},
	tea.KeyDelete:   {escByte, '[', '3', '~'},
	tea.KeyF1:       {escByte, 'O', 'P'},
	tea.KeyF2:       {escByte, 'O', 'Q'},
	tea.KeyF3:       {escByte, 'O', 'R'},
	tea.KeyF4:       {escByte, 'O', 'S'},
	tea.KeyF5:       {escByte, '[', '1', '5', '~'},
	tea.KeyF6:       {escByte, '[', '1', '7', '~'},
	tea.KeyF7:       {escByte, '[', '1', '8', '~'},
	tea.KeyF8:       {escByte, '[', '1', '9', '~'},
	tea.KeyF9:       {escByte, '[', '2', '0', '~'},
	tea.KeyF10:      {escByte, '[', '2', '1', '~'},
	tea.KeyF11:      {escByte, '[', '2', '3', '~'},
	tea.KeyF12:      {escByte, '[', '2', '4', '~'},
	tea.KeyEnter:    {'\r'},
	tea.KeyTab:      {'\t'},
	tea.KeyBackspace: {'\x7f'},
	tea.KeyEscape:   {escByte},
	tea.KeySpace:    {' '},
}

var ctrlKeyBytes = map[tea.KeyType]byte{
	tea.KeyCtrlA: 0x01,
	tea.KeyCtrlB: 0x02,
	tea.KeyCtrlC: 0x03,
	tea.KeyCtrlD: 0x04,
	tea.KeyCtrlE: 0x05,
	tea.KeyCtrlF: 0x06,
	tea.KeyCtrlG: 0x07,
	tea.KeyCtrlH: 0x08,
	tea.KeyCtrlK: 0x0b,
	tea.KeyCtrlL: 0x0c,
	tea.KeyCtrlN: 0x0e,
	tea.KeyCtrlO: 0x0f,
	tea.KeyCtrlP: 0x10,
	tea.KeyCtrlQ: 0x11,
	tea.KeyCtrlR: 0x12,
	tea.KeyCtrlS: 0x13,
	tea.KeyCtrlT: 0x14,
	tea.KeyCtrlU: 0x15,
	tea.KeyCtrlV: 0x16,
	tea.KeyCtrlW: 0x17,
	tea.KeyCtrlX: 0x18,
	tea.KeyCtrlY: 0x19,
	tea.KeyCtrlZ: 0x1a,
}

func KeyMsgToBytes(msg tea.KeyMsg) []byte {
	if seq, exists := keyTypeToSequence[msg.Type]; exists {
		return prependAlt(msg, seq)
	}

	if b, exists := ctrlKeyBytes[msg.Type]; exists {
		return []byte{b}
	}

	if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
		return runeBytes(msg)
	}

	return nil
}

func prependAlt(msg tea.KeyMsg, seq []byte) []byte {
	if !msg.Alt {
		return seq
	}
	result := make([]byte, 0, len(seq)+1)
	result = append(result, escByte)
	result = append(result, seq...)
	return result
}

func runeBytes(msg tea.KeyMsg) []byte {
	buf := make([]byte, 0, len(msg.Runes)*utf8.UTFMax)
	for _, r := range msg.Runes {
		encoded := make([]byte, utf8.UTFMax)
		n := utf8.EncodeRune(encoded, r)
		buf = append(buf, encoded[:n]...)
	}
	if msg.Alt {
		result := make([]byte, 0, len(buf)+1)
		result = append(result, escByte)
		result = append(result, buf...)
		return result
	}
	return buf
}
