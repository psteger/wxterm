//go:build js && wasm

package config

import (
	"os"
	"strings"
	"syscall/js"
	"testing"
)

func installTestStorage(t *testing.T, descriptor js.Value) {
	t.Helper()
	global := js.Global()
	object := global.Get("Object")
	original := object.Call("getOwnPropertyDescriptor", global, "localStorage")
	descriptor.Set("configurable", true)
	object.Call("defineProperty", global, "localStorage", descriptor)
	t.Cleanup(func() {
		if original.IsUndefined() {
			global.Get("Reflect").Call("deleteProperty", global, "localStorage")
		} else {
			object.Call("defineProperty", global, "localStorage", original)
		}
	})
}

func TestBrowserStorageRoundTrip(t *testing.T) {
	descriptor := js.Global().Get("Function").New(`
		const data = new Map();
		return {value: {
			getItem(key) { return data.has(key) ? data.get(key) : null; },
			setItem(key, value) { data.set(key, value); }
		}};
	`).Invoke()
	installTestStorage(t, descriptor)
	if _, err := readConfig(); !os.IsNotExist(err) {
		t.Fatalf("missing config: got %v, want os.ErrNotExist", err)
	}
	want := &Config{UseFahrenheit: true}
	if err := want.Save(); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil || got == nil || !got.UseFahrenheit {
		t.Fatalf("round trip: got %+v, %v", got, err)
	}
}

func TestBrowserStorageExceptions(t *testing.T) {
	for name, script := range map[string]string{
		"blocked getter": `return {get() { throw new Error("storage blocked"); }};`,
		"blocked methods": `return {value: {
			getItem() { throw new Error("storage blocked"); },
			setItem() { throw new Error("quota exceeded"); }
		}};`,
	} {
		t.Run(name, func(t *testing.T) {
			descriptor := js.Global().Get("Function").New(script).Invoke()
			installTestStorage(t, descriptor)
			_, readErr := readConfig()
			writeErr := writeConfig([]byte(`{}`))
			for operation, err := range map[string]error{"read": readErr, "write": writeErr} {
				if err == nil || !strings.Contains(err.Error(), "browser storage unavailable") {
					t.Errorf("%s: got %v, want storage error", operation, err)
				}
			}
		})
	}
}
