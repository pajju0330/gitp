# gitp usage

## Flags

| Flag | Meaning |
|------|---------|
| `-p`, `--profile NAME` | Force profile `~/.gitconfig.NAME` |
| `-v`, `--verbose` | Print selected profile/source/path to stderr |
| `-h`, `--help` | Help |
| `-V`, `--version` | Version |

Flags may appear anywhere before or among git args. `--profile` always wins over repo binding, env, and default.

## Profile resolution

```
--profile / -p
    ↓ (else)
repo git config gitp.profile     # set by: gitp use NAME
    ↓ (else)
$GITP_PROFILE
    ↓ (else)
~/.gitp/config  default_profile = …
    ↓ (else)
plain git (~/.gitconfig)
```

When a profile is selected, gitp sets `GIT_CONFIG_GLOBAL` to that file (overriding any prior value) and execs `git`.

## Builtins vs passthrough

If the first non-flag argument is a builtin (`whoami`, `doctor`, `profile`, `use`, `sync`, `ship`, `completion`), gitp handles it and does **not** exec git for that command (except where the builtin itself runs git subprocesses).

Anything else is forwarded:

```bash
gitp -p work commit -m "msg"   # → git commit -m "msg" with work profile
```

## Identity commands

```bash
gitp whoami
# profile:  personal (flag)
# config:   /Users/you/.gitconfig.personal
# name:     …
# email:    …
# ssh:      ssh -i ~/.ssh/id_ed25519_personal -o IdentitiesOnly=yes

gitp doctor
# [ok]/fail]/warn]/info] lines for profile, identity, SSH key, GitHub auth, remote URL
```

## Profile management

```bash
gitp profile list
gitp profile show personal
gitp profile edit personal
gitp profile create client --from work
gitp profile use personal          # write ~/.gitp/config
gitp profile use --clear
gitp profile init-ssh personal     # key + core.sshCommand
```

## Repo binding

```bash
cd ~/code/my-oss-project
gitp use personal                  # git config --local gitp.profile personal
gitp push                          # no --profile needed
gitp use --clear
```

Global default (`profile use`) and repo bind (`use`) are different:

- `gitp profile use X` → every repo (unless overridden)
- `gitp use X` → this repo only

## Workflow helpers

```bash
gitp sync     # fetch --prune && rebase @{u}
gitp ship     # push -u origin HEAD && gh pr create --fill
```

`ship` requires [GitHub CLI](https://cli.github.com/) for the PR step; push still runs without it.

## Verbose debugging

```bash
gitp -v --profile personal push
# stderr: gitp: profile=personal source=flag config=/Users/you/.gitconfig.personal
```

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Runtime error (missing profile, doctor failures, git failure via builtins) |
| 2 | Usage / parse error |

Passthrough `exec` preserves git’s own exit code when the process is replaced.
