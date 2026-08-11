---
description: Run everything CI runs, and check the things CI cannot, before pushing
---

Everything CI runs, in CI's order:

```bash
make ci
```

That is `generate`, `fmt-check`, `terraform-check`, `vet`, `lint`, `test`,
`provider`, `docs-check`. If it passes, CI passes — with one exception worth
stating: `make generate` downloads the spec again on the runner, so upstream can
change between your run and CI's.

`docs-check` compares `docs/` against git, so it fails while regenerated docs are
uncommitted even though nothing is wrong. Commit them and rerun rather than
assuming the failure is spurious — read the list it prints to be sure that is
what it is.

Then the things `make ci` does not cover:

- **Is the working tree what you think it is?** `git status --short`. Generated
  outputs are ignored; `docs/` and `examples/` are not, and are part of the change.
- **Does the commit explain why?** The house style is prose about the reason and
  the failure it prevents, not a summary of the diff. `git log` is the reference.
- **Did you touch a generated file by hand?** `internal/gen/`,
  `internal/registry/registry_gen.go`, `internal/bwanclient/mock/`. Any edit
  there is lost on the next `make generate`; it belongs in the generator.
- **Behaviour a test would have caught.** New runtime behaviour without a test in
  `internal/genresource` or `internal/provider`, or a new generator rule without
  one in `tools/tfgen` or `tools/tfdocs`, is the gap to close before pushing.

If pushing to a branch CI watches, note that a newer push cancels the run in
flight (`cancel-in-progress`) — a "cancelled" conclusion on the previous commit
is that, not a failure.
