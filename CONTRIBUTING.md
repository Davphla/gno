# Contributing to Gno

Thank you for contributing to Gno! This guide will help you get started.

## Important Resources

- **[Documentation](https://docs.gno.land)** - comprehensive documentation for Gno
- **[Go Package Docs](https://gnolang.github.io/gno/github.com/gnolang/gno.html)** - API reference documentation
- **[Awesome Gno](https://github.com/gnoverse/awesome-gno)** - curated list of Gno resources
- **[Discord](https://discord.gg/YFtMjWwUN7)** - join our community for discussions and support

## Getting Started

### Prerequisites

- Go 1.25+
- Unix environment (Linux/macOS/WSL2)
- `make` command

### Setup

```bash
git clone https://github.com/gnolang/gno.git
cd gno
make install
```

If `gno` and `gnokey` commands are not found, see [Go's official
documentation](https://go.dev/doc/tutorial/compile-install) for configuring your
PATH.

### Testing

Run all tests:
```bash
make test
```

Test specific Gno code:
```bash
gno test ./examples/... -v
```

## Your First Contribution

Start from an issue labelled
[good first issue](https://github.com/gnolang/gno/labels/%F0%9F%97%BA%EF%B8%8Fgood%20first%20issue%F0%9F%97%BA%EF%B8%8F)
or [help wanted](https://github.com/gnolang/gno/labels/help%20wanted), and
comment on it before writing code. An issue labelled WIP or META is a
coordination thread, not a task.

## Project Structure

See the [README](./README.md) for project structure overview. Most important
directories have their own README explaining their purpose and how to
contribute.

**Tip**: Look at recent commits to understand typical file modifications:
```bash
git log --oneline -10
```

## Making Changes

### Submitting Pull Requests

1. **Open as draft first** - Start with a draft PR to run initial checks
2. **Check CI results** - We have extensive CI to catch issues early
3. **Write the body** - What breaks, what the change does, which issue it closes
4. **Move to ready** - Once CI passes and you've self-reviewed

Our CI is designed to help both you and maintainers identify potential side
effects of changes. Use it as a guide to improve your PR.

### Git Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat(scope): add new feature`
- `fix(scope): fix issue`
- `docs(scope): update documentation`

### Rebasing

Avoid rebasing after opening your PR for review. Maintainers handle the final
squash/merge. Add new commits to address feedback instead of force-pushing.

Using merge commits instead of rebase allows reviewers
to better review changes only since their last review.
To disable rebase when using `git pull` on the repository, run:

	git config pull.rebase false

This will be disabled only for the git repository you're currently on.

### Code Style

- Read [PHILOSOPHY.md](./PHILOSOPHY.md) to understand our approach
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use existing patterns in the codebase
- Run `make fmt` and `make lint` before committing

### Editor Setup

See [docs/builders/editor-setup.md](./docs/builders/editor-setup.md) for
configuring your editor with LSP support, autocompletion, and formatting
for `.gno` files.

## Reporting Issues

Before opening an issue:
1. Search existing issues
2. Include reproduction steps, version info, and logs (see `gno bug`)

## Quality Standards

Your contribution should:
- Solve a real problem
- Include tests
- Update documentation if needed
- Follow existing patterns

## What We Do Not Merge

A pull request that changes text without changing behaviour is closed:

- Typo, grammar and wording edits where the original was already unambiguous.
- Formatting that `make fmt` does not produce.
- A `.md` file restating one that already exists.
- A stub constant or a `TODO` that closes an issue on paper.

Volume earns nothing. What raises GovDAO rank is
[Notable contribution](https://github.com/gnolang/gno/labels/%F0%9F%8F%86Notable%20contribution%F0%9F%8F%86).

Fold a typo fix into the pull request that does the work.

## Communication

See the [Community section](./README.md#community) in our README for Discord,
GitHub discussions, and other communication channels.

## Architecture Decision Records (ADRs)

Non-trivial changes should include an ADR documenting the context and
reasoning behind the work. ADRs go in the component's `adr/` folder:

- `gnovm/adr/` — VM, interpreter, type-checker, transpiler
- `gno.land/adr/` — node, SDK, keeper, RPC, genesis
- `tm2/adr/` — consensus, p2p, mempool, crypto

See [AGENTS.md](./AGENTS.md#architecture-decision-records-adrs) for format
details.

## AI-Assisted Contributions

AI coding agents (Claude, Copilot, etc.) are welcome tools for contributing
to Gno. **A human is always responsible for AI-assisted work.** Contributions
must be submitted under the responsible human's GitHub account. If an
autonomous agent submits work independently, it must clearly disclose its
human owner in the PR description.

AI-assisted PRs must include an ADR documenting the context the AI operated
under (see above), unless the change is trivial (bug fixes, formatting,
simple tests, docs-only). The human is responsible for reviewing the output
for correctness, style, and security. All the same standards apply: CI must
pass, tests must be included, conventional commits must be used.

A pull request whose body advertises an autonomous agent or a bounty payout,
and names no human owner, is closed.

If using AI, point your agent at [AGENTS.md](./AGENTS.md).

## Documentation Philosophy

`docs/` is optimized for humans, readable by agents.
`AGENTS.md` is optimized for agents, readable by humans.

When writing documentation, keep human docs in `docs/` — narrative,
examples, context. Avoid bloating `AGENTS.md` with content better suited
for `docs/`. Conversely, a few lines of concise, direct rules in
`AGENTS.md` is far better than pointing agents to multiple human-oriented
documents they'd have to parse in full.

---

For more documentation, see the [docs](./docs/) folder.

> For maintainers with merge access, see [RELEASING.md](RELEASING.md) for internal processes (versioning, branching, release workflow).
