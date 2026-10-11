package main

import "github.com/egoist/mygo/ui"

func (m *windowManager) view(c *ui.Context) {
	c.SetTheme(ui.DarkTheme())
	ui.Column(c).Fill().Padding(20).Gap(12).Children(func() {
		if m.loadError != "" {
			ui.Text(c, "Couldn’t load your windows").FontSize(18).Bold()
			ui.Text(c, m.loadError).TextColor(c.Theme().Danger)
			return
		}
		if m.saveError != "" {
			ui.Text(c, m.saveError).TextColor(c.Theme().Danger)
			ui.Button(c, "Retry saving").OnClick(m.changed)
		}
		m.editor(c)
	})
}

func (m *windowManager) editor(c *ui.Context) {
	c.OnShortcut(0, ui.KeyEscape, m.cancelEdit)
	action := "Open"
	if m.editID == "" {
		input := ui.TextInput(c.Key("url"), &m.url).Bind(&m.urlInput).
			Label("URL").Placeholder("URL").Height(36).FontSize(14).OnSubmit(m.saveDraft)
		if m.formError != "" {
			input.Border(1, c.Theme().Danger)
			ui.Text(c, m.formError).FontSize(12).TextColor(c.Theme().Danger)
		}
	} else {
		action = "Save changes"
		ui.Text(c, "Edit window").FontSize(18).Bold()
		ui.Field(c, "URL", func() {
			ui.TextInput(c.Key("url"), &m.url).Bind(&m.urlInput).OnSubmit(m.saveDraft)
		}).Error(m.formError)
		ui.Field(c, "Name", func() {
			ui.TextInput(c.Key("name"), &m.name).OnSubmit(m.saveDraft)
		})
		ui.Checkbox(c, &m.alwaysOnTop, "Always on top")
	}
	ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
		ui.Button(c, "Cancel").Height(28).OnClick(m.cancelEdit)
		ui.PrimaryButton(c, action).Height(28).MinWidth(72).OnClick(m.saveDraft)
	})
}
