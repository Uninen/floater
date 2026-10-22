package main

import (
	"path/filepath"
	"testing"
)

func TestRestoreRefreshesOpenWindowsOnce(t *testing.T) {
	created := make(map[string]*testWindow)
	m := newWindowManager([]savedWindow{
		{ID: "first", Open: true},
		{ID: "second", Open: true},
		{ID: "closed"},
	}, func(w savedWindow, _ func()) managedWindow {
		win := &testWindow{}
		created[w.ID] = win
		return win
	})
	m.changed = func() {}

	m.restore()
	if len(created) != 2 || created["closed"] != nil {
		t.Fatalf("startup opened %v; want only first and second", created)
	}
	for _, id := range []string{"first", "second"} {
		win := created[id]
		if win == nil {
			t.Fatalf("startup did not restore %s", id)
		}
		if win.refreshes != 0 {
			t.Fatalf("%s refreshed before its first page load", id)
		}
		win.finishLoad()
		if win.refreshes != 1 {
			t.Fatalf("%s: got %d startup refreshes; want 1", id, win.refreshes)
		}
		win.finishLoad()
		m.show(id)
		win.finishLoad()
		if win.refreshes != 1 {
			t.Errorf("%s: later loads or window shows must not refresh again", id)
		}
	}

	m.show("closed")
	win := created["closed"]
	if win == nil {
		t.Fatal("could not open the closed window")
	}
	win.finishLoad()
	if win.refreshes != 0 {
		t.Error("opening a closed window later must not trigger a startup refresh")
	}
}

func TestQuitRestoresCurrentPages(t *testing.T) {
	store := &windowStore{path: filepath.Join(t.TempDir(), "windows.json")}
	created := make(map[string]*testWindow)
	create := func(w savedWindow, _ func()) managedWindow {
		win := &testWindow{url: w.URL}
		created[w.ID] = win
		return win
	}
	m := newWindowManager([]savedWindow{
		{ID: "first", URL: "https://example.com/", Open: true},
		{ID: "second", URL: "https://example.org/", Open: true},
		{ID: "closed", URL: "https://example.net/"},
	}, create)
	m.changed = func() {
		if err := store.save(1, m.windows); err != nil {
			t.Fatal(err)
		}
	}
	m.restore()
	m.changed()
	want := map[string]string{
		"first":  "https://example.com/articles/last?view=full#details",
		"second": "https://example.org/app/settings#profile",
	}
	for id, url := range want {
		created[id].url = url
	}
	m.quit()
	for id := range m.open {
		m.closed(id)
	}

	windows, err := store.load()
	if err != nil {
		t.Fatal(err)
	}
	created = make(map[string]*testWindow)
	restarted := newWindowManager(windows, create)
	restarted.changed = func() {}
	restarted.restore()
	for id, url := range want {
		if win := created[id]; win == nil || win.url != url {
			t.Errorf("%s did not restore its last page %q: got %+v", id, url, win)
		}
	}
	if created["closed"] != nil {
		t.Error("restart opened a window that was closed")
	}
}

type testWindow struct {
	url          string
	loadFinished func()
	refreshes    int
}

func (*testWindow) Show()                  {}
func (*testWindow) Restore()               {}
func (*testWindow) Focus()                 {}
func (*testWindow) Close()                 {}
func (*testWindow) SetTitle(string)        {}
func (*testWindow) SetAlwaysOnTop(bool)    {}
func (*testWindow) Navigate(string)        {}
func (w *testWindow) URL() string          { return w.url }
func (w *testWindow) ReloadIgnoringCache() { w.refreshes++ }

func (w *testWindow) OnDidFinishLoad(fn func()) func() {
	w.loadFinished = fn
	return func() { w.loadFinished = nil }
}

func (w *testWindow) finishLoad() {
	if w.loadFinished != nil {
		w.loadFinished()
	}
}
