---
description: Regenerate everything from the spec and report what changed in the provider's surface
---

Rerun the whole pipeline and account for the difference. Useful after changing
`generator_config.yml` or the generators, and as a way to find out what upstream
did since the last run.

```bash
make generate      # spec → schemas → registry → mocks
make examples      # a usage example for anything that has none
make docs          # the registry documentation, from the provider's own schema
```

Then read the diff rather than trusting it:

```bash
git status --short
git diff --stat docs examples
```

`internal/gen/`, `internal/registry/registry_gen.go`, `internal/bwanclient/mock/`,
`openapi*.yaml|json`, `generator_config_gen.yml`, `provider_code_spec_gen.json`
and `provider_schema_gen.json` are gitignored build outputs — expected to change,
nothing to commit. `docs/` and `examples/` are committed and are the visible
change.

Two things worth checking by hand, because nothing else will:

- **Stale generated directories.** The generators write but never delete, so a
  removed object leaves `internal/gen/resources/resource_<name>/` behind. It
  compiles and is dead. CI never sees it (fresh checkout), so remove it locally
  or `make clean && make generate`.
- **Attributes changing required-ness or type.** A `docs/` diff moving an
  argument between Required, Optional and Read-Only is a practitioner-visible
  change even when nothing failed. Call those out.

Finish with:

```bash
make ci
```

Report the resource and data source count before and after, every attribute that
changed shape, and anything that stopped being exposed.
