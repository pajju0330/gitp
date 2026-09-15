# gitp

Thin Git CLI wrapper that selects a named global config profile, then execs the real `git`.

```bash
gitp --profile personal push -u origin main
# → GIT_CONFIG_GLOBAL=~/.gitconfig.personal git push -u origin main
```

## Install

```bash
make install   # builds to bin/gitp and installs to $GOBIN or ~/go/bin
# or
go install ./cmd/gitp
```

Ensure the install directory is on your `PATH`.

## Usage

```bash
gitp [--profile|-p NAME] <git-command> [args...]
gitp --help
gitp --version
```

Examples:

```bash
gitp --profile personal clone git@github.com:me/repo.git
gitp -p work commit -m "fix"
gitp --profile=personal config --global user.email
gitp status   # uses default profile if configured, else plain git
```

## Profile selection (highest wins)

1. **`--profile` / `-p`** on the command line — always overrides
2. **`GITP_PROFILE`** environment variable
3. **`default_profile`** in `~/.gitp/config`
4. Git’s normal global config (`~/.gitconfig`)

`--profile` also overrides any existing `GIT_CONFIG_GLOBAL` in the environment.

## Setup

### 1. Create a profile file

```bash
cp ~/.gitconfig ~/.gitconfig.personal
# edit user.email, core.sshCommand, etc.
```

Example `~/.gitconfig.personal`:

```ini
[user]
    name = Prajwal
    email = personal@example.com
[core]
    sshCommand = ssh -i ~/.ssh/id_ed25519_personal
```

### 2. Optional default profile

```bash
mkdir -p ~/.gitp
echo 'default_profile = personal' > ~/.gitp/config
```

Or:

```bash
export GITP_PROFILE=personal
```

## How it works

`gitp` strips its own flags, sets `GIT_CONFIG_GLOBAL` to `~/.gitconfig.{name}`, and replaces itself with `git` via `exec`. Exit codes, TTY, pagers, and credential prompts behave like native Git.

## Develop

```bash
make test
make build
./bin/gitp --help
```
