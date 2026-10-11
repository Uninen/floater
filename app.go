package main

import (
	_ "embed"
	"log"
	"path/filepath"
	"slices"
	"sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

//go:embed resources/tray.png
var trayIcon []byte

type floater struct {
	*windowManager
	store        *windowStore
	editorWindow *mygo.Window
	urlForm      nativeURLForm
	editorSize   [2]int
	tray         *mygo.Tray
	revision     uint64
	saves        sync.WaitGroup
}

type nativeURLForm interface {
	Show(string)
	Hide()
	Reset(string)
}

type nativeWindow struct{ *mygo.Window }

func (w nativeWindow) Navigate(url string) { w.Page().LoadURL(url) }

func createPage(w savedWindow, closed func()) managedWindow {
	win := mygo.NewWindow(mygo.WindowOptions{
		Title:           w.Name,
		Width:           1024,
		Height:          768,
		URL:             w.URL,
		StateKey:        "page-" + w.ID,
		Frameless:       true,
		AlwaysOnTop:     w.AlwaysOnTop,
		SkipTaskbar:     true,
		AutoHideMenuBar: true,
		Page: mygo.PageOptions{
			PreloadScript: `
				document.addEventListener("DOMContentLoaded", () => {
					const style = document.createElement("style");
					style.textContent = "* { scrollbar-width: none !important; } *::-webkit-scrollbar { display: none !important; }";
					document.head.append(style);
				}, { once: true });
			`,
		},
	})
	enableOptionDrag(win)
	win.OnClosed(closed)
	return nativeWindow{win}
}

func (a *floater) ready() {
	width, height := a.viewSize()
	a.editorWindow = mygo.NewWindow(mygo.WindowOptions{
		Title:           "Floater",
		Width:           width,
		Height:          height,
		UseContentSize:  true,
		BackgroundColor: "#18181b",
		StateKey:        "editor",
		Hidden:          true,
		SkipTaskbar:     true,
		DisableResize:   true,
		Content:         ui.View(a.view),
	})
	a.dismiss = a.editorWindow.Hide
	a.changed = a.persist
	a.viewChanged = a.layoutEditor
	a.urlForm = newNativeURLForm(a.editorWindow, func(url string) {
		a.url = url
		a.saveDraft()
		if a.formError != "" {
			mygo.Dialog.Error("Invalid URL", a.formError)
			a.urlForm.Show(a.url)
		}
	})
	a.editorWindow.OnClose(func(e *mygo.CloseEvent) {
		if !a.quitting {
			e.PreventDefault()
			a.cancelEdit()
		}
	})
	var err error
	a.tray, err = mygo.NewTray(mygo.TrayOptions{
		Icon:           trayIcon,
		IconIsTemplate: true,
		ToolTip:        "Floater — your floating windows",
		Menu:           mygo.NewMenu(a.trayItems()),
	})
	if err != nil {
		log.Fatal("menu bar: ", err)
	}
	if a.loadError == "" {
		a.restore()
	}
	a.refresh()
	if len(a.windows) == 0 || a.loadError != "" {
		a.showEditor()
	}
}

func (a *floater) layoutEditor() {
	if a.urlForm != nil {
		if a.editID == "" && a.loadError == "" && a.saveError == "" {
			a.urlForm.Show(a.url)
		} else {
			a.urlForm.Hide()
		}
	}
	width, height := a.viewSize()
	size := [2]int{width, height}
	if size == a.editorSize {
		return
	}
	a.editorSize = size
	a.editorWindow.SetContentSize(width, height)
}

func (a *floater) showEditor() {
	a.layoutEditor()
	a.editorWindow.Restore()
	a.editorWindow.Show()
	a.editorWindow.Focus()
	a.editorWindow.Invalidate()
}

func (a *floater) newWindow() {
	a.beginNew()
	if a.urlForm != nil {
		a.urlForm.Reset("")
	}
	a.showEditor()
}

func (a *floater) refresh() {
	a.layoutEditor()
	a.tray.SetMenu(mygo.NewMenu(a.trayItems()))
	mygo.App.SetMenu(a.appMenu())
	a.editorWindow.Invalidate()
}

func (a *floater) persist() {
	a.refresh()
	a.revision++
	revision := a.revision
	windows := slices.Clone(a.windows)
	a.saves.Add(1)
	go func() {
		defer a.saves.Done()
		err := a.store.save(revision, windows)
		if err != nil {
			log.Println("save windows:", err)
		}
		a.editorWindow.Update(func() {
			if revision != a.revision {
				return
			}
			a.saveError = ""
			if err != nil {
				a.saveError = err.Error()
				if !a.quitting {
					a.showEditor()
				}
			}
		})
	}()
}

func (a *floater) trayItems() []*mygo.MenuItem {
	items := []*mygo.MenuItem{
		{Label: "New window…", Disabled: a.loadError != "", Click: func(*mygo.MenuItem, *mygo.Window) { a.newWindow() }},
		mygo.Separator(),
	}
	items = append(items, a.windowItems()...)
	return append(items, mygo.Separator(), &mygo.MenuItem{Role: mygo.RoleQuit})
}

func (a *floater) windowItems() []*mygo.MenuItem {
	if len(a.windows) == 0 {
		return []*mygo.MenuItem{{Label: "No windows", Disabled: true}}
	}
	var items []*mygo.MenuItem
	for _, w := range a.windows {
		label := "Open"
		if a.open[w.ID] != nil {
			label = "Show window"
		}
		items = append(items, &mygo.MenuItem{Label: w.Name, Submenu: []*mygo.MenuItem{
			{Label: label, Click: func(*mygo.MenuItem, *mygo.Window) { a.show(w.ID) }},
			{Label: "Close window", Disabled: a.open[w.ID] == nil, Click: func(*mygo.MenuItem, *mygo.Window) { a.close(w.ID) }},
			{Label: "Always on top", Type: mygo.MenuItemCheckbox, Checked: w.AlwaysOnTop,
				Click: func(item *mygo.MenuItem, _ *mygo.Window) { a.setFloating(w.ID, item.Checked) }},
			mygo.Separator(),
			{Label: "Edit…", Click: func(*mygo.MenuItem, *mygo.Window) {
				a.beginEdit(w.ID)
				a.showEditor()
			}},
			{Label: "Remove", Click: func(*mygo.MenuItem, *mygo.Window) { a.remove(w.ID) }},
		}})
	}
	return items
}

func (a *floater) appMenu() *mygo.Menu {
	return mygo.NewMenu([]*mygo.MenuItem{
		{Role: mygo.RoleAppMenu},
		{Label: "File", Submenu: []*mygo.MenuItem{
			{Label: "New window…", Accelerator: "CmdOrCtrl+N", Disabled: a.loadError != "", Click: func(*mygo.MenuItem, *mygo.Window) { a.newWindow() }},
			mygo.Separator(),
			{Role: mygo.RoleClose},
		}},
		{Role: mygo.RoleEditMenu},
		{Label: "Window", Submenu: a.windowItems()},
	})
}

func loadApp() (*floater, error) {
	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		return nil, err
	}
	store := &windowStore{path: filepath.Join(dir, "windows.json")}
	windows, loadErr := store.load()
	a := &floater{windowManager: newWindowManager(windows, createPage), store: store}
	if loadErr != nil {
		a.loadError = store.path + ": " + loadErr.Error()
	}
	return a, nil
}
