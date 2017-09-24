package filebeat

import (
	"github.com/rswestmoreland/rebeat/input"
)

func init() {
	input.Register("filebeat", New)
}
