//go:build js

package config

import (
	"fmt"
	"os"
	"syscall/js"
)

const storageKey = "wxterm.config"

// In the browser the config lives in localStorage. Use Reflect.get through Call:
// Call converts JavaScript exceptions to recoverable Go panics, whereas Get
// lets a throwing localStorage getter escape directly into JavaScript.
func browserStorage() js.Value {
	return js.Global().Get("Reflect").Call("get", js.Global(), "localStorage")
}

// Convert storage access and quota exceptions to errors.
func readConfig() (data []byte, err error) {
	defer recoverJS(&err)
	v := browserStorage().Call("getItem", storageKey)
	if v.IsNull() {
		return nil, os.ErrNotExist
	}
	return []byte(v.String()), nil
}

func writeConfig(data []byte) (err error) {
	defer recoverJS(&err)
	browserStorage().Call("setItem", storageKey, string(data))
	return nil
}

func recoverJS(err *error) {
	if r := recover(); r != nil {
		*err = fmt.Errorf("browser storage unavailable: %v", r)
	}
}
