//go:build !darwin

package main

import "github.com/egoist/mygo"

const nativeURLDialogHeight = 116

func newNativeURLForm(*mygo.Window, func(string)) nativeURLForm { return nil }
