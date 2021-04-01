package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

func Open(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("invalid os")
	}
}
