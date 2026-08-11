---
description: Expose a new resource or data source by adding it to generator_config.yml
argument-hint: [object or endpoint, e.g. "radius servers" or /radius-servers]
---

Expose `$ARGUMENTS` through the provider. It is a configuration change, not a
new Go file — if it seems to need hand-written code, say what the generator
cannot do rather than writing around it.

## 1. Find what the API offers

The spec is `openapi.json` after `make generate` (or `openapi_tf_gen.yaml` for
the normalised form the generators see). Establish, for the object:

- which of POST / GET / PATCH / DELETE exist, and on which paths;
- whether a single object can be fetched at all — some collections have no
  `GET /things/{id}`, and the runtime then refreshes by walking the collection;
- whether the create body branches (`oneOf`), and whether the response does;
- whether any field is declared without a type, which forces the `_raw`
  treatment described in CLAUDE.md.

## 2. Write the entry

Add it to `generator_config.yml` in alphabetical position, under `resources:` for
something manageable and `data_sources:` for something readable. Follow the
neighbours: paths use `_` in placeholders (prep renames the API's `-`), and a
comment above the entry earns its place when the object is odd — no update, no
single-object GET, a variant, a gate.

Conventions that are enforced elsewhere, so getting them wrong fails the build
rather than shipping:

- an object with an opaque configuration document is named `<name>_raw` and
  carries `x_terraform.raw_feature: <name>`;
- one endpoint serving several kinds of object gets one entry per kind, each
  with `x_terraform.variant: <kind>`;
- a collection data source is the plural name; the singular addresses one object.

## 3. Generate and look at what came out

```bash
make generate
```

Read the generated schema for the new object rather than assuming: `make docs`
then the page in `docs/`. Check that required arguments are the ones the API
actually requires, that nothing landed as a bare `_raw` string unexpectedly, and
that a variant's kind is recognised (`Variant` in
`internal/registry/registry_gen.go` should carry a discriminator and value, or
distinctive fields).

## 4. Example and docs

`make examples` writes a minimal example and an import command for anything that
has none. Read it. If the minimal version is useless — the interesting arguments
are all optional, as with a cloud account's credentials — write the example by
hand in `examples/{resources,data-sources}/netskopebwan_<name>/`; hand-written
files are never overwritten.

Then `make docs` and commit `docs/` and `examples/` with the change.

## 5. Verify

```bash
make ci
```

Report what the object exposes, anything the API forced (no update, no single-object
GET, a gate) and anything you could not model.
