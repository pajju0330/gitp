package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/prajwal/gitp/internal/config"
	"github.com/prajwal/gitp/internal/git"
	"github.com/prajwal/gitp/internal/profile"
)

// Context carries IO and the resolved profile environment for builtins.
type Context struct {
	Stdout      io.Writer
	Stderr      io.Writer
	Env         []string
	ProfileName string
	ProfilePath string
	Source      string // how profile was selected
	Verbose     bool
}

func logVerbose(c Context, format string, args ...any) {
	if c.Verbose {
		fmt.Fprintf(c.Stderr, "gitp: "+format+"\n", args...)
	}
}

// Whoami prints the active identity.
func Whoami(c Context) int {
	name := git.ConfigGet("user.name", c.Env)
	email := git.ConfigGet("user.email", c.Env)
	ssh := git.ConfigGet("core.sshCommand", c.Env)

	if c.ProfileName == "" {
		fmt.Fprintf(c.Stdout, "profile:  (none — using git default global config)\n")
	} else {
		fmt.Fprintf(c.Stdout, "profile:  %s (%s)\n", c.ProfileName, c.Source)
		fmt.Fprintf(c.Stdout, "config:   %s\n", c.ProfilePath)
	}
	fmt.Fprintf(c.Stdout, "name:     %s\n", dash(name))
	fmt.Fprintf(c.Stdout, "email:    %s\n", dash(email))
	fmt.Fprintf(c.Stdout, "ssh:      %s\n", dash(ssh))
	return 0
}

func dash(s string) string {
	if s == "" {
		return "(unset)"
	}
	return s
}

// Doctor diagnoses common profile/auth problems.
func Doctor(c Context) int {
	issues := 0
	print := func(status, msg string) {
		fmt.Fprintf(c.Stdout, "[%s] %s\n", status, msg)
	}

	if c.ProfileName == "" {
		print("warn", "no profile selected (flag / repo / GITP_PROFILE / default)")
	} else {
		print("ok", fmt.Sprintf("profile %q via %s → %s", c.ProfileName, c.Source, c.ProfilePath))
	}

	email := git.ConfigGet("user.email", c.Env)
	name := git.ConfigGet("user.name", c.Env)
	if email == "" || name == "" {
		print("fail", "user.name / user.email incomplete")
		issues++
	} else {
		print("ok", fmt.Sprintf("identity %s <%s>", name, email))
	}

	sshCmd := git.ConfigGet("core.sshCommand", c.Env)
	if sshCmd == "" {
		print("warn", "core.sshCommand unset — git will use default SSH keys (easy to hit wrong GitHub account)")
	} else {
		print("ok", "core.sshCommand = "+sshCmd)
		if key := extractIdentityFile(sshCmd); key != "" {
			if _, err := os.Stat(expandHome(key)); err != nil {
				print("fail", "SSH identity file missing: "+key)
				issues++
			} else {
				print("ok", "SSH identity file exists: "+key)
				auth := sshAuthCheck(sshCmd)
				if auth != "" {
					print("ok", "GitHub SSH auth: "+auth)
				} else {
					print("warn", "could not verify GitHub SSH auth (ssh -T failed or timed out)")
				}
			}
		}
	}

	remote, _ := git.Output([]string{"remote", "get-url", "origin"}, c.Env)
	if remote == "" {
		print("warn", "no origin remote (skip remote checks)")
	} else {
		print("ok", "origin = "+remote)
		switch {
		case strings.HasPrefix(remote, "git@") || strings.HasPrefix(remote, "ssh://"):
			print("ok", "remote uses SSH — auth = SSH key / core.sshCommand (not user.email)")
		case strings.HasPrefix(remote, "https://") || strings.HasPrefix(remote, "http://"):
			print("warn", "remote uses HTTPS — auth = credential helper / token, not SSH key")
			helper := git.ConfigGet("credential.helper", c.Env)
			if helper == "" {
				print("warn", "credential.helper unset")
			} else {
				print("ok", "credential.helper = "+helper)
			}
			if _, err := exec.LookPath("gh"); err == nil {
				out, _ := exec.Command("gh", "auth", "status").CombinedOutput()
				line := firstLine(string(out))
				if line != "" {
					print("info", "gh: "+line)
				}
			}
		}
	}

	if issues > 0 {
		print("fail", fmt.Sprintf("%d issue(s) found", issues))
		return 1
	}
	print("ok", "doctor found no hard failures")
	return 0
}

func extractIdentityFile(sshCmd string) string {
	fields := strings.Fields(sshCmd)
	for i, f := range fields {
		if (f == "-i" || f == "-oIdentityFile") && i+1 < len(fields) {
			return fields[i+1]
		}
		if strings.HasPrefix(f, "-i") && len(f) > 2 {
			return f[2:]
		}
	}
	return ""
}

func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

func sshAuthCheck(sshCmd string) string {
	// sshCmd is like: ssh -i KEY -o IdentitiesOnly=yes
	fields := strings.Fields(sshCmd)
	args := append(fields[1:], "-o", "BatchMode=yes", "-o", "ConnectTimeout=5", "-T", "git@github.com")
	bin := fields[0]
	cmd := exec.Command(bin, args...)
	out, _ := cmd.CombinedOutput()
	s := string(out)
	// GitHub prints: Hi USERNAME! You've successfully authenticated...
	if i := strings.Index(s, "Hi "); i >= 0 {
		rest := s[i+3:]
		if j := strings.Index(rest, "!"); j >= 0 {
			return strings.TrimSpace(rest[:j])
		}
	}
	if strings.Contains(s, "Permission denied") {
		return ""
	}
	return ""
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}

// Profile dispatches profile subcommands.
func Profile(c Context, args []string) int {
	if len(args) == 0 {
		fmt.Fprintf(c.Stderr, "usage: gitp profile list|show|edit|create|use|init-ssh\n")
		return 2
	}
	switch args[0] {
	case "list":
		names, err := profile.List()
		if err != nil {
			fmt.Fprintf(c.Stderr, "gitp: %v\n", err)
			return 1
		}
		settings, _ := config.Load()
		for _, n := range names {
			mark := "  "
			if n == settings.FileDefault {
				mark = "* "
			}
			path, _ := profile.PathFor(n)
			fmt.Fprintf(c.Stdout, "%s%s\t%s\n", mark, n, path)
		}
		if len(names) == 0 {
			fmt.Fprintf(c.Stdout, "(no profiles — create with: gitp profile create NAME)\n")
		}
		return 0

	case "show":
		if len(args) < 2 {
			fmt.Fprintf(c.Stderr, "usage: gitp profile show NAME\n")
			return 2
		}
		if err := profile.Show(args[1], c.Stdout); err != nil {
			fmt.Fprintf(c.Stderr, "gitp: %v\n", err)
			return 1
		}
		return 0

	case "edit":
		if len(args) < 2 {
			fmt.Fprintf(c.Stderr, "usage: gitp profile edit NAME\n")
			return 2
		}
		if err := profile.Edit(args[1]); err != nil {
			fmt.Fprintf(c.Stderr, "gitp: %v\n", err)
			return 1
		}
		return 0

	case "create":
		if len(args) < 2 {
			fmt.Fprintf(c.Stderr, "usage: gitp profile create NAME [--from PROFILE|default]\n")
			return 2
		}
		name := args[1]
		from := "default"
		for i := 2; i < len(args); i++ {
			if args[i] == "--from" && i+1 < len(args) {
				from = args[i+1]
				i++
			}
		}
		path, err := profile.Create(name, from)
		if err != nil {
			fmt.Fprintf(c.Stderr, "gitp: %v\n", err)
			return 1
		}
		fmt.Fprintf(c.Stdout, "created %s\n", path)
		return 0

	case "use":
		return profileUse(c, args[1:])

	case "init-ssh":
		if len(args) < 2 {
			fmt.Fprintf(c.Stderr, "usage: gitp profile init-ssh NAME\n")
			return 2
		}
		name := args[1]
		env := c.Env
		if path, err := profile.Resolve(name); err == nil {
			env = git.WithGlobalConfig(c.Env, path)
		}
		email := git.ConfigGet("user.email", env)
		pub, generated, err := profile.InitSSH(name, email)
		if err != nil {
			fmt.Fprintf(c.Stderr, "gitp: %v\n", err)
			return 1
		}
		if generated {
			fmt.Fprintf(c.Stdout, "generated key for profile %q\n", name)
		} else {
			fmt.Fprintf(c.Stdout, "using existing key for profile %q\n", name)
		}
		fmt.Fprintf(c.Stdout, "set core.sshCommand in ~/.gitconfig.%s\n", name)
		fmt.Fprintf(c.Stdout, "\nAdd this public key to the matching GitHub account:\n\n")
		data, _ := os.ReadFile(pub)
		fmt.Fprintf(c.Stdout, "%s\n", strings.TrimSpace(string(data)))
		fmt.Fprintf(c.Stdout, "\nhttps://github.com/settings/keys\n")
		return 0

	default:
		fmt.Fprintf(c.Stderr, "gitp: unknown profile subcommand %q\n", args[0])
		return 2
	}
}

func profileUse(c Context, args []string) int {
	if len(args) == 0 {
		fmt.Fprintf(c.Stderr, "usage: gitp profile use NAME | gitp profile use --clear\n")
		return 2
	}
	if args[0] == "--clear" {
		if err := config.SetDefaultProfile(""); err != nil {
			fmt.Fprintf(c.Stderr, "gitp: %v\n", err)
			return 1
		}
		fmt.Fprintf(c.Stdout, "cleared default_profile in ~/.gitp/config\n")
		return 0
	}
	name := args[0]
	if _, err := profile.Resolve(name); err != nil {
		fmt.Fprintf(c.Stderr, "gitp: %v\n", err)
		return 1
	}
	if err := config.SetDefaultProfile(name); err != nil {
		fmt.Fprintf(c.Stderr, "gitp: %v\n", err)
		return 1
	}
	fmt.Fprintf(c.Stdout, "default profile set to %q\n", name)
	return 0
}

// Use binds a profile to the current repo (or clears it).
func Use(c Context, args []string) int {
	if len(args) == 0 {
		fmt.Fprintf(c.Stderr, "usage: gitp use NAME | gitp use --clear\n")
		return 2
	}
	if args[0] == "--clear" {
		if err := git.SetRepoProfile("", c.Env); err != nil {
			fmt.Fprintf(c.Stderr, "gitp: %v\n", err)
			return 1
		}
		fmt.Fprintf(c.Stdout, "cleared repo gitp.profile\n")
		return 0
	}
	name := args[0]
	if _, err := profile.Resolve(name); err != nil {
		fmt.Fprintf(c.Stderr, "gitp: %v\n", err)
		return 1
	}
	if err := git.SetRepoProfile(name, c.Env); err != nil {
		fmt.Fprintf(c.Stderr, "gitp: %v\n", err)
		return 1
	}
	fmt.Fprintf(c.Stdout, "repo bound to profile %q (git config gitp.profile)\n", name)
	return 0
}

// Sync fetches and rebases onto upstream.
func Sync(c Context) int {
	logVerbose(c, "sync: fetch")
	if code := git.RunInherit([]string{"fetch", "--prune"}, c.Env); code != 0 {
		return code
	}
	logVerbose(c, "sync: rebase @{u}")
	return git.RunInherit([]string{"rebase", "@{u}"}, c.Env)
}

// Ship pushes the current branch and opens a PR via gh when available.
func Ship(c Context, args []string) int {
	pushArgs := append([]string{"push", "-u", "origin", "HEAD"}, args...)
	logVerbose(c, "ship: %s", strings.Join(pushArgs, " "))
	if code := git.RunInherit(pushArgs, c.Env); code != 0 {
		return code
	}
	if _, err := exec.LookPath("gh"); err != nil {
		fmt.Fprintf(c.Stderr, "gitp: pushed; install GitHub CLI (gh) to open a PR automatically\n")
		return 0
	}
	logVerbose(c, "ship: gh pr create")
	cmd := exec.Command("gh", "pr", "create", "--fill")
	cmd.Env = c.Env
	cmd.Stdin = os.Stdin
	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stderr
	if err := cmd.Run(); err != nil {
		// already has PR is fine-ish — surface message
		fmt.Fprintf(c.Stderr, "gitp: gh pr create: %v\n", err)
		return 1
	}
	return 0
}

// Completion prints shell completion scripts.
func Completion(c Context, args []string) int {
	if len(args) < 1 {
		fmt.Fprintf(c.Stderr, "usage: gitp completion bash|zsh|fish\n")
		return 2
	}
	switch args[0] {
	case "bash":
		fmt.Fprint(c.Stdout, bashCompletion)
	case "zsh":
		fmt.Fprint(c.Stdout, zshCompletion)
	case "fish":
		fmt.Fprint(c.Stdout, fishCompletion)
	default:
		fmt.Fprintf(c.Stderr, "gitp: unsupported shell %q\n", args[0])
		return 2
	}
	return 0
}

const bashCompletion = `# bash completion for gitp
_gitp() {
  local cur profiles
  cur="${COMP_WORDS[COMP_CWORD]}"
  profiles=$(ls -1 ~/.gitconfig.* 2>/dev/null | sed 's|.*/\.gitconfig\.||')
  case "${COMP_WORDS[1]}" in
    profile)
      COMPREPLY=( $(compgen -W "list show edit create use init-ssh" -- "$cur") )
      ;;
    completion)
      COMPREPLY=( $(compgen -W "bash zsh fish" -- "$cur") )
      ;;
    *)
      COMPREPLY=( $(compgen -W "whoami doctor profile use sync ship completion --profile -p --verbose -v --help --version $profiles" -- "$cur") )
      ;;
  esac
}
complete -F _gitp gitp
`

const zshCompletion = `#compdef gitp
_gitp() {
  local -a profiles builtins
  profiles=(${(f)"$(ls -1 ~/.gitconfig.* 2>/dev/null | sed 's|.*/\.gitconfig\.||')"})
  builtins=(whoami doctor profile use sync ship completion)
  _arguments \
    '(-p --profile)'{-p,--profile}'[profile name]:profile:($profiles)' \
    '(-v --verbose)'{-v,--verbose}'[verbose]' \
    '(-h --help)'{-h,--help}'[help]' \
    '(-V --version)'{-V,--version}'[version]' \
    '1:command:(${builtins} status push pull commit clone fetch rebase log diff)' \
    '*::arg:->args'
}
compdef _gitp gitp
`

const fishCompletion = `complete -c gitp -s p -l profile -d 'Profile name' -a '(ls -1 ~/.gitconfig.* 2>/dev/null | string replace -r ".*/\\.gitconfig\\." "")'
complete -c gitp -s v -l verbose -d 'Verbose'
complete -c gitp -s h -l help -d 'Help'
complete -c gitp -s V -l version -d 'Version'
complete -c gitp -n '__fish_use_subcommand' -a 'whoami doctor profile use sync ship completion'
complete -c gitp -n '__fish_seen_subcommand_from profile' -a 'list show edit create use init-ssh'
complete -c gitp -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish'
`
