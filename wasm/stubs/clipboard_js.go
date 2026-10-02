//go:build js

package clipboard

import "errors"

func readAll() (string, error) { return "", errors.New("clipboard unavailable") }

func writeAll(string) error { return errors.New("clipboard unavailable") }
