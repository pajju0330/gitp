package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// SetRepoProfile sets or unsets local gitp.profile. Empty name unsets.
func SetRepoProfile(name string, env []string) error {
	bin, err := Finder("git")
	if err != nil {
		return fmt.Errorf("git not found in PATH: %w", err)
	}
	var cmd *exec.Cmd
	if name == "" {
		cmd = exec.Command(bin, "config", "--local", "--unset-all", RepoProfileKey)
	} else {
		cmd = exec.Command(bin, "config", "--local", RepoProfileKey, name)
	}
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		// git config --unset exits 5 when the key is absent
		if name == "" {
			if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 5 {
				return nil
			}
		}
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}
