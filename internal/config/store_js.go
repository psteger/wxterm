//go:build js

package config

import (
	"fmt"
	"os"
	"syscall/js"
)

const storageKey = "wxterm.config"

// In the browser the config lives in localStorage. syscall/js panics on
// JavaScript exceptions (storage blocked, quota), so convert them to errors.
func readConfig() (data []byte, err error) {
	defer recoverJS(&err)
	v := js.Global().Get("localStorage").Call("getItem", storageKey)
	if v.IsNull() {
		return nil, os.ErrNotExist
	}
	return []byte(v.String()), nil
}

func writeConfig(data []byte) (err error) {
	defer recoverJS(&err)
	js.Global().Get("localStorage").Call("setItem", storageKey, string(data))
	return nil
}

func recoverJS(err *error) {
	if r := recover(); r != nil {
		*err = fmt.Errorf("browser storage unavailable: %v", r)
	}
}
