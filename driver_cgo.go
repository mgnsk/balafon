//go:build cgo

package balafon

import (
	// Register the rtmidi driver.
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)
