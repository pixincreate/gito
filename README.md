# gito

A thin, opinionated CLI wrapper around GitHub's API for pull request workflows.

## Why?

`gh` is great. `curl` is great. Remembering the exact API endpoint, headers, and JSON shape every time you want to fetch review comments... not so great.

`gito` gives you:

- **One interface, two backends.** Has `gh`? Uses `gh`. Doesn't? Falls back to `curl` + `GITHUB_PAT`. You don't think about it.
- **Human-readable review comments.** No more piping JSON through `jq` gymnastics. Comments come out formatted and ready to read.
- **Less typing.** `gito pull comments 10` beats `gh api -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2022-11-28" /repos/you/your-repo/pulls/10/comments | jq '.[] | {author: .user.login, body: .body, diff: .diff_hunk}'` any day.

## Install

### `go install`

```bash
go install github.com/pixincreate/gito@latest
```

### Build it yourself

```bash
git clone https://github.com/pixincreate/gito.git
cd gito
make build
# binary lands in dist/gito
```

## Setup

### Option A: Already have `gh`?

```bash
gh auth login
```

Done. `gito` will detect it automatically.

### Option B: Token in env

```bash
export GITHUB_PAT=ghp_your_token_here
```

### What happens at runtime

```
gh installed and authenticated?
  yes --> use gh
  no  --> print yellow warning, try curl
           GITHUB_PAT set?
             yes --> use curl
             no  --> error, tell you what to do
```

You can skip auto-detection entirely:

```bash
gito --backend=gh pr list
gito --backend=curl pr list
```

## Commands

Every command supports `--repo owner/repo`. Skip it and gito reads your git remote.

### Pull requests

```bash
# Create
gito pr create "Fix the thing" --base main --head fix/thing --label bug --assignee you

# List (default: open, limit 30)
gito pr list
gito pr list --state closed --limit 10

# View details
gito pr view 42

# Merge
gito pr merge 42
gito pr merge 42 --method squash --delete-branch

# Close without merging
gito pr close 42

# Raw diff
gito pr diff 42

# Who's been asked to review
gito pr reviewers 42
```

### Review comments

This is the main reason gito exists.

```bash
# Fetch all review comments on a PR
gito pull comments 10

# Save to file
gito pull comments 10 --output comments.txt
```

Output looks like this:

```
Author: Copilot
PR Number: 10
Diff: @@ -33,13 +33,26 @@ pub use convert::{ConvertOptions, Format, convert, convert_with_options};
Review comment: `is_jsonc_path()` currently treats the entire string as the extension...
URL: https://github.com/user-name/repo-name/pull/10#discussion_r2896771811
Created At: 2026-03-06T16:42:47Z
Author Association: CONTRIBUTOR
--------------------------------------------------------------------------------
Author: Copilot
PR Number: 10
Diff: @@ -0,0 +1,421 @@
Review comment: The tests validate `is_jsonc_path("file") == false`, but they don't cover...
URL: https://github.com/user-name/repo-name/pull/10#discussion_r2896771824
Created At: 2026-03-06T16:42:47Z
Author Association: CONTRIBUTOR
--------------------------------------------------------------------------------
```

No JSON. No jq. Just text you can actually read.

### Reply to a comment

```bash
gito pull replies 10 2896771811 --body "Good catch, fixed."
```

## Repo auto-detection

All commands figure out `owner/repo` from your git remote when you don't pass `--repo`. If you're in a repo directory, just run the command.

```bash
cd ~/code/my-project
gito pr list              # knows it's your-username/my-project
```

## All the flags

```bash
gito --help
gito pr --help
gito pr create --help
gito pull --help
gito pull comments --help
```

## License

[ CC0-1.0 license ](LICENSE)
