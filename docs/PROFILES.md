# Profiles

## Files

| Path | Role |
|------|------|
| `~/.gitconfig.NAME` | Profile global git config |
| `~/.gitp/config` | gitp settings (`default_profile = …`) |
| repo `.git/config` → `gitp.profile` | Per-repo binding |

A profile file is a normal gitconfig:

```ini
[user]
    name = Prajwal
    email = personal@example.com
[core]
    sshCommand = ssh -i ~/.ssh/id_ed25519_personal -o IdentitiesOnly=yes
[push]
    autoSetupRemote = true
```

## Creating profiles

```bash
gitp profile create personal --from default   # copy ~/.gitconfig
gitp profile create work --from personal
```

## SSH (recommended for `git@github.com` remotes)

GitHub account selection for SSH is **not** `user.email`. It is the key offered by SSH.

Per profile:

```bash
gitp profile init-ssh personal
# → ~/.ssh/id_ed25519_personal
# → sets core.sshCommand in ~/.gitconfig.personal
# → prints pubkey to add at https://github.com/settings/keys
```

Equivalent manual config:

```ini
[core]
    sshCommand = ssh -i ~/.ssh/id_ed25519_personal -o IdentitiesOnly=yes
```

`IdentitiesOnly=yes` stops ssh-agent from offering your work key first.

Verify:

```bash
gitp --profile personal doctor
ssh -i ~/.ssh/id_ed25519_personal -o IdentitiesOnly=yes -T git@github.com
# Hi your-personal-username!
```

## HTTPS remotes

If `origin` is `https://github.com/...`:

- Auth uses credential helper / token / `gh`, not `core.sshCommand`
- Keep a matching `credential.helper` in the profile if needed
- `gitp doctor` reports HTTPS and shows `gh auth status` when `gh` is installed

For multi-account HTTPS, prefer SSH profiles or separate `gh` hosts — gitp does not switch `gh` accounts automatically.

## Default vs repo vs flag

```bash
gitp profile use work      # default for all repos
gitp use personal          # this repo only
gitp -p client status      # one-shot override
```

## Environment

| Variable | Effect |
|----------|--------|
| `GITP_PROFILE` | Default profile (beats `~/.gitp/config`) |
| `GIT_CONFIG_GLOBAL` | Overridden when gitp selects a profile |
| `EDITOR` | Used by `gitp profile edit` |
