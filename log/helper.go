package log

import "github.com/crazy-choose/go/usage"

func defLogName() string {
	return usage.HomeDir() + "/logs/" + usage.ExeName() + ".log"
}
