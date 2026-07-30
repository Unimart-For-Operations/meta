package platform

import "os/exec"

// OpenBrowser opens the given URL in the default browser.
// On macOS it uses "open", on Linux it tries "xdg-open".
func OpenBrowser(url string) error {
	if IsDarwin() {
		return exec.Command("open", url).Start()
	}
	return exec.Command("xdg-open", url).Start()
}
