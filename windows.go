package main

import (
	"crypto/rand"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
)

type savedWindow struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	AlwaysOnTop bool   `json:"alwaysOnTop"`
	Open        bool   `json:"open"`
}

type managedWindow interface {
	Show()
	Restore()
	Focus()
	Close()
	SetTitle(string)
	SetAlwaysOnTop(bool)
	Navigate(string)
}

type windowManager struct {
	windows     []savedWindow
	open        map[string]managedWindow
	create      func(savedWindow, func()) managedWindow
	changed     func()
	viewChanged func()
	dismiss     func()
	quitting    bool
	loadError   string
	saveError   string
	editID      string
	name        string
	url         string
	alwaysOnTop bool
	formError   string
	urlInput    ui.Handle
}

func newWindowManager(windows []savedWindow, create func(savedWindow, func()) managedWindow) *windowManager {
	m := &windowManager{windows: windows, open: make(map[string]managedWindow), create: create, viewChanged: func() {}}
	m.beginNew()
	return m
}

func normalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("Enter a website URL.")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || strings.ContainsAny(u.Host, " \t\r\n") {
		return "", fmt.Errorf("Enter a valid http:// or https:// website URL.")
	}
	return u.String(), nil
}

func (m *windowManager) index(id string) int {
	return slices.IndexFunc(m.windows, func(w savedWindow) bool { return w.ID == id })
}

func (m *windowManager) beginNew() {
	m.editID = ""
	m.name, m.url, m.formError = "", "", ""
	m.alwaysOnTop = false
	m.urlInput.Focus()
	m.viewChanged()
}

func (m *windowManager) beginEdit(id string) {
	if i := m.index(id); i >= 0 {
		w := m.windows[i]
		m.editID = id
		m.name, m.url, m.alwaysOnTop = w.Name, w.URL, w.AlwaysOnTop
		m.formError = ""
		m.urlInput.Focus()
		m.viewChanged()
	}
}

func (m *windowManager) cancelEdit() {
	m.formError = ""
	m.viewChanged()
	m.dismiss()
}

func (m *windowManager) viewSize() (int, int) {
	if m.loadError != "" || m.saveError != "" {
		return 580, 480
	}
	height := nativeURLDialogHeight
	if m.editID != "" {
		height = 264
	}
	if m.formError != "" && (m.editID != "" || nativeURLDialogHeight == 116) {
		height += 28
	}
	return 440, height
}

func (m *windowManager) saveDraft() {
	if m.loadError != "" {
		return
	}
	rawURL, err := normalizeURL(m.url)
	if err != nil {
		m.formError = err.Error()
		m.urlInput.Focus()
		m.viewChanged()
		return
	}
	name := strings.TrimSpace(m.name)
	if name == "" {
		u, _ := url.Parse(rawURL)
		name = u.Hostname()
	}
	if m.editID == "" {
		w := savedWindow{ID: rand.Text(), Name: name, URL: rawURL, AlwaysOnTop: m.alwaysOnTop, Open: true}
		m.windows = append(m.windows, w)
		m.show(w.ID)
	} else if i := m.index(m.editID); i >= 0 {
		previous := m.windows[i]
		w := &m.windows[i]
		w.Name, w.URL, w.AlwaysOnTop = name, rawURL, m.alwaysOnTop
		if native := m.open[w.ID]; native != nil {
			native.SetTitle(w.Name)
			native.SetAlwaysOnTop(w.AlwaysOnTop)
			if previous.URL != w.URL {
				native.Navigate(w.URL)
			}
		}
	}
	m.formError = ""
	m.changed()
	m.viewChanged()
	m.dismiss()
}

func (m *windowManager) show(id string) {
	i := m.index(id)
	if i < 0 {
		return
	}
	w := m.open[id]
	if w == nil {
		w = m.create(m.windows[i], func() { m.closed(id) })
		m.open[id] = w
	}
	w.Restore()
	w.Show()
	w.Focus()
	if !m.windows[i].Open {
		m.windows[i].Open = true
		m.changed()
	}
}

func (m *windowManager) restore() {
	for _, w := range m.windows {
		if w.Open {
			m.show(w.ID)
		}
	}
}

func (m *windowManager) close(id string) {
	if w := m.open[id]; w != nil {
		w.Close()
	}
}

func (m *windowManager) closed(id string) {
	delete(m.open, id)
	// Quitting preserves the workspace; a deliberate close only shelves this window.
	if !m.quitting {
		if i := m.index(id); i >= 0 {
			m.windows[i].Open = false
			m.changed()
		}
	}
}

func (m *windowManager) setFloating(id string, floating bool) {
	if i := m.index(id); i >= 0 {
		m.windows[i].AlwaysOnTop = floating
		if w := m.open[id]; w != nil {
			w.SetAlwaysOnTop(floating)
		}
		m.changed()
	}
}

func (m *windowManager) remove(id string) {
	if i := m.index(id); i >= 0 {
		m.windows = slices.Delete(m.windows, i, i+1)
		m.close(id)
		m.changed()
	}
}
