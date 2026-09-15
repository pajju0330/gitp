# gitp

Thin Git CLI wrapper: pick a named global config profile, then `exec` real `git`.

```bash
gitp --profile personal push -u origin main
# → GIT_CONFIG_GLOBAL=~/.gitconfig.personal git push -u origin main
```

## Install

```bash
make install          # → $(go env GOPATH)/bin/gitp
# or
go install ./cmd/gitp
```

Ensure `$(go env GOPATH)/bin` is on your `PATH`.

Homebrew (local tap / formula in-repo):

```bash
brew install --formula ./Formula/gitp.rb
```

## Quick start

```bash
# 1. Create a profile from your current global config
gitp profile create personal --from default

# 2. Edit identity / SSH
gitp profile edit personal
# or bootstrap SSH key + core.sshCommand:
gitp profile init-ssh personal
# then add the printed pubkey to https://github.com/settings/keys

# 3. Set a default (optional)
gitp profile use personal

# 4. Use it
gitp push
gitp -v status          # see which profile is active
gitp whoami
gitp doctor
```

## Profile selection (highest wins)

1. `--profile` / `-p` on the command line (also overrides `GIT_CONFIG_GLOBAL`)
2. Repo-local `gitp.profile` (`gitp use personal`)
3. `GITP_PROFILE` environment variable
4. `default_profile` in `~/.gitp/config`
5. Git’s normal `~/.gitconfig`

## Commands

### Passthrough

Any non-builtin args go to git:

```bash
gitp [--profile NAME] [-v] <git-command> [args...]
```

### Builtins

| Command | Purpose |
|---------|---------|
| `gitp whoami` | Active profile, email, `core.sshCommand` |
| `gitp doctor` | Diagnose profile / SSH / remote auth (SSH vs HTTPS) |
| `gitp profile list` | List `~/.gitconfig.*` (`*` = default) |
| `gitp profile show NAME` | Print profile file |
| `gitp profile edit NAME` | Open in `$EDITOR` |
| `gitp profile create NAME [--from PROFILE\|default]` | New profile file |
| `gitp profile use NAME` | Set global default in `~/.gitp/config` |
| `gitp profile use --clear` | Clear global default |
| `gitp profile init-ssh NAME` | Create `~/.ssh/id_ed25519_NAME` + set `core.sshCommand` |
| `gitp use NAME` | Bind profile to **this repo** (`git config gitp.profile`) |
| `gitp use --clear` | Unbind repo profile |
| `gitp sync` | `git fetch --prune` + `git rebase @{u}` |
| `gitp ship` | `git push -u origin HEAD` + `gh pr create --fill` |
| `gitp completion bash\|zsh\|fish` | Print shell completion script |

## SSH note

`user.email` does **not** choose which GitHub account authenticates. For SSH remotes (`git@github.com:...`), set `core.sshCommand` per profile (or run `gitp profile init-ssh`).

HTTPS remotes use credential helpers / `gh` auth instead — `gitp doctor` will warn.

## Completions

```bash
# zsh
gitp completion zsh > ~/.zsh/completions/_gitp

# bash
gitp completion bash >> ~/.bashrc

# fish
gitp completion fish > ~/.config/fish/completions/gitp.fish
```

## Docs

- [docs/USAGE.md](docs/USAGE.md) — full usage guide
- [docs/PROFILES.md](docs/PROFILES.md) — profile files, SSH, HTTPS

## Develop

```bash
make test
make build
./bin/gitp --help
```
