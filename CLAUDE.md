# Working in this repository

A Terraform provider whose resources, data sources, schemas, registry, docs and
examples are all generated from the BWAN v2 OpenAPI document. [`README.md`](README.md)
explains the pipeline; this file is what to know before changing anything.

Everything runs inside the project shell. `devenv shell` supplies Go, Terraform,
`golangci-lint`, the HashiCorp generators, `mockgen` and `tfplugindocs`; outside
it, `make tools` says what is missing.

## The rule everything else follows

**Nothing is written per object.** A resource or data source is an entry in
[`generator_config.yml`](generator_config.yml) and nothing else. If a change
seems to need a hand-written resource, the generator is what needs the change —
say so rather than writing the resource.

The same applies downstream: `docs/`, `examples/`, `internal/gen/`,
`internal/registry/registry_gen.go` and `internal/bwanclient/mock/` are outputs.
Never edit them by hand. `docs/` and `examples/` are committed anyway, because
the registry serves docs straight from the tagged tree and the docs are built
from the examples.

## Before anything compiles

A fresh checkout does not build: `main.go` imports packages that do not exist
until they are generated.

```bash
devenv shell
make generate        # spec download → schemas → registry → mocks
```

`make generate` downloads the spec every time, from `OPENAPI_SPEC_URL`
(defaulting to the tenant the Makefile names). It is not pinned, so two runs a
day apart can generate different providers.

## The spec moves under you

Treat generation failures as spec drift first, your change second. The failure
looks like this:

```
error finding resource(s): failed to extract 'served_tenant.create':
  path '/served-tenants' not found in OpenAPI spec
```

That is `generator_config.yml` naming a path the spec no longer has. Diagnose it
by comparing the two — `/spec-drift` does this — and expect renames as well as
removals: `/tags` became `/overlay-tags`, and its `type` discriminator moved
into a nested `config`. Do not "fix" it by deleting the entry until you have
checked whether the endpoint was renamed.

Prep warnings are load-bearing. `tfgen: prep: generator_config claims a variant
of /tags, which the spec does not have` is the same problem one step earlier.

## What the generators guarantee, and what they cannot

Derived from the configuration, so there is no second place to update: a read
path without `{id}` is refreshed by walking its collection; a resource with no
update forces replacement; a path placeholder other than `{id}` becomes a
required attribute that forces replacement; a data source gains `filter` where
the collection is filterable.

Two things the runtime cannot derive, and which have rules:

- **Variants.** One endpoint serving several kinds of object becomes several
  Terraform types via `x_terraform.variant`. The kind is recognised on read
  through `Variant.Discriminator`, which may be a dotted path (`config.type`).
  A variant with no discriminator and no distinctive fields cannot be told apart
  from its siblings — that is a bug, not a shrug.
- **Opaque documents.** An object whose configuration the API declares no shape
  for is named `*_raw` and gated behind `enable_raw_<feature>`. `tfgen registry`
  fails the build if a raw object is ungated or unsuffixed, so the plain name
  stays free for a typed replacement.

## Checks

`make ci` is exactly what CI runs, in order: `generate`, `fmt-check`,
`terraform-check`, `vet`, `lint`, `test`, `provider`, `docs-check`. Run it
before pushing — `/preflight` does.

`docs-check` fails when the committed docs differ from generated ones, which
also catches drift in `examples/`, since the docs embed them. It compares
against git, so regenerated docs must be committed for it to pass.

Tests state the API calls they expect through a mock of `bwanclient.API`; a
request the test did not ask for fails it. That is the property worth keeping
when the engine works out its requests from a schema.

## Releasing

`make release VERSION=v1.2.3` tags and pushes; the tag is what triggers the
release workflow. It is irreversible once the registry ingests the version, so
the target checks everything it can first and makes you type the tag.

Two things learned the hard way:

- A tag push runs the workflow **from the tagged commit**. Fixing a broken
  release workflow on the branch does nothing for a tag that already exists —
  move the tag or cut a new version.
- Prereleases publish as ordinary GitHub releases. Marking one as a GitHub
  "pre-release" stops the Terraform registry ingesting it at all. `.goreleaser.yml`
  says so where someone would otherwise re-add the flag.

A prerelease build refuses to configure until the practitioner acknowledges its
exact version with `enable_pre_release` — see `internal/provider/provider.go`.

## Commits

Explain why, not what; the diff covers what. `git log` is the house style — the
reason a change exists, and the failure it prevents, in prose.

## Standard flows

`/add-object`, `/regen`, `/spec-drift`, `/preflight`, `/cut-release` in
[`.claude/commands`](.claude/commands). They are readable as documentation as
well as runnable.
