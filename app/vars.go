package app

import "os"

var Debug bool

func init() {
	Debug = os.Getenv("DEBUG_APP") == "true"
}
