package main

import (
	"log"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego/objc"
	"github.com/egoist/mygo"
)

const urlFormHeight = 88

type cocoaPoint struct{ X, Y float64 }
type cocoaSize struct{ Width, Height float64 }
type cocoaRect struct {
	Origin cocoaPoint
	Size   cocoaSize
}

func cocoa(object objc.ID, selector string, args ...any) objc.ID {
	return object.Send(objc.RegisterName(selector), args...)
}

func cocoaClass(name string) objc.ID { return objc.ID(objc.GetClass(name)) }

func cocoaString(text string) objc.ID {
	value := cocoa(cocoaClass("NSString"), "alloc")
	value = cocoa(value, "initWithBytes:length:encoding:", unsafe.StringData(text), len(text), 4)
	runtime.KeepAlive(text)
	return cocoa(value, "autorelease")
}

func cocoaText(value objc.ID) string {
	pointer := objc.Send[*byte](value, objc.RegisterName("UTF8String"))
	length := cocoa(value, "lengthOfBytesUsingEncoding:", 4)
	return string(unsafe.Slice(pointer, int(length)))
}

var urlFormClass objc.Class
var urlFormTargets = map[objc.ID]*appKitURLForm{}

type appKitURLForm struct {
	window, original, view, field, button, target objc.ID
	submit                                        func(string)
	visible                                       bool
}

func newAppKitURLForm(window *mygo.Window, submit func(string)) *appKitURLForm {
	if urlFormClass == 0 {
		var err error
		urlFormClass, err = objc.RegisterClass("FloaterURLForm", objc.GetClass("NSObject"), nil, nil, []objc.MethodDef{
			{Cmd: objc.RegisterName("openURL:"), Fn: func(self objc.ID, _ objc.SEL, _ objc.ID) {
				form := urlFormTargets[self]
				cocoa(form.window, "makeFirstResponder:", objc.ID(0))
				form.submit(cocoaText(cocoa(form.field, "stringValue")))
			}},
			{Cmd: objc.RegisterName("control:textView:doCommandBySelector:"), Fn: func(self objc.ID, _ objc.SEL, _, _ objc.ID, command objc.SEL) bool {
				if command != objc.RegisterName("cancelOperation:") {
					return false
				}
				cocoa(urlFormTargets[self].window, "performClose:", objc.ID(0))
				return true
			}},
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	native := objc.ID(window.NativeHandle())
	f := &appKitURLForm{window: native, submit: submit}
	// Keep MyGo’s content alive while AppKit owns the window’s content view.
	f.original = cocoa(cocoa(native, "contentView"), "retain")
	f.view = cocoa(cocoa(cocoaClass("NSView"), "alloc"), "initWithFrame:", cocoaRect{Size: cocoaSize{440, urlFormHeight}})
	f.target = cocoa(objc.ID(urlFormClass), "new")
	urlFormTargets[f.target] = f
	f.field = cocoa(cocoaClass("NSTextField"), "textFieldWithString:", cocoaString(""))
	cocoa(f.field, "setFrame:", cocoaRect{Origin: cocoaPoint{20, 32}, Size: cocoaSize{312, 24}})
	cocoa(f.field, "setPlaceholderString:", cocoaString("URL"))
	cocoa(f.field, "setBezelStyle:", 1)
	cocoa(f.field, "setDelegate:", f.target)
	cocoa(f.view, "addSubview:", f.field)
	f.button = cocoa(cocoaClass("NSButton"), "buttonWithTitle:target:action:", cocoaString("Open"), f.target, objc.RegisterName("openURL:"))
	cocoa(f.button, "setFrame:", cocoaRect{Origin: cocoaPoint{344, 28}, Size: cocoaSize{76, 32}})
	cocoa(f.button, "setBezelStyle:", 1)
	cocoa(f.button, "setKeyEquivalent:", cocoaString("\r"))
	cocoa(f.view, "addSubview:", f.button)
	window.OnClosed(func() {
		delete(urlFormTargets, f.target)
		cocoa(f.target, "release")
		cocoa(f.view, "release")
		cocoa(f.original, "release")
	})
	return f
}

func (f *appKitURLForm) Show(url string) {
	if !f.visible {
		f.Reset(url)
		cocoa(f.window, "setContentView:", f.view)
		cocoa(f.window, "setDefaultButtonCell:", cocoa(f.button, "cell"))
		f.visible = true
	}
	cocoa(f.window, "makeFirstResponder:", f.field)
}

func (f *appKitURLForm) Hide() {
	if f.visible {
		cocoa(f.window, "setDefaultButtonCell:", objc.ID(0))
		cocoa(f.window, "setContentView:", f.original)
		cocoa(f.window, "makeFirstResponder:", f.original)
		f.visible = false
	}
}

func (f *appKitURLForm) Reset(url string) {
	cocoa(f.field, "setStringValue:", cocoaString(url))
}
