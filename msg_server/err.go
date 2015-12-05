

package main

import (
	"errors"
)

var (
	NOTOPIC         = errors.New("NO TOPIC")
	CMD_NOT_CORRECT = errors.New("CMD NOT CORRECT")
)