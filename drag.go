package main

import (
	"github.com/ebitengine/purego/objc"
	"github.com/egoist/mygo"
)

func enableOptionDrag(w *mygo.Window) {
	mygo.RunOnMain(func() {
		const (
			leftMouseDownMask = uint64(1 << 1)
			optionFlag        = 1 << 19
		)

		native := objc.ID(w.NativeHandle())
		events := objc.ID(objc.GetClass("NSEvent"))
		eventWindow := objc.RegisterName("window")
		modifierFlags := objc.RegisterName("modifierFlags")
		drag := objc.RegisterName("performWindowDragWithEvent:")

		handler := objc.NewBlock(func(_ objc.Block, event objc.ID) objc.ID {
			if event.Send(eventWindow) != native || event.Send(modifierFlags)&optionFlag == 0 {
				return event
			}
			native.Send(drag, event)
			return 0
		})
		defer handler.Release()

		monitor := events.Send(objc.RegisterName("addLocalMonitorForEventsMatchingMask:handler:"), leftMouseDownMask, handler)
		w.OnClosed(func() {
			events.Send(objc.RegisterName("removeMonitor:"), monitor)
		})
	})
}
