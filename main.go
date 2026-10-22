package main

import (
	"log"

	"github.com/egoist/mygo"
)

func main() {
	mygo.Theme.SetSource(mygo.ThemeDark)
	mygo.App.SetActivationPolicy(mygo.ActivationPolicyAccessory)
	if !mygo.App.RequestSingleInstanceLock() {
		return
	}
	app, err := loadApp()
	if err != nil {
		log.Fatal(err)
	}
	mygo.App.SetMenu(app.appMenu())
	mygo.App.OnWindowAllClosed(func() {})
	mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) { app.quit() })
	mygo.App.OnSecondInstance(func([]string, string) { app.newWindow() })
	mygo.App.WhenReady(app.ready)
	err = mygo.App.Run()
	app.saves.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
