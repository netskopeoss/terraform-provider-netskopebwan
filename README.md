# terraform-provider-netskopebwan

Terraform provider for Netskope Borderless WAN, generated from the BWAN v2
OpenAPI document that the REST gateway itself serves.

> [!WARNING]
> **The 1.x line is an alpha; use the released 0.x line for now.** 1.x resources,
> data sources and attributes are generated from the BWAN v2 OpenAPI
> specification, which is itself still changing, so any of them may be renamed,
> reshaped or removed from one prerelease to the next, and a configuration or
> state written against one may need reworking for the next. Until 1.0.0 is
> released, ask for `version = "~> 0.0"`. Terraform installs a prerelease only for
> an exact version, and a prerelease then refuses to configure until that same
> version is acknowledged with the `enable_pre_release` argument, so nobody runs
> the 1.x line by accident. Setting it accepts that instability: a prerelease is
> NOT covered by the provider's backward-compatibility guarantees.

Nothing here is written per resource. Adding an object to the provider means
adding an entry to [`generator_config.yml`](generator_config.yml).

## How it is generated

```
   openapi.json                                the BWAN v2 spec, as the API serves it
        │
        ▼  tfgen prep
   openapi_tf_gen.yaml                         spec the generators can map
        │
        ▼  tfplugingen-openapi + generator_config.yml
   provider_code_spec_gen.json                 provider code specification
        │
        ├──▶ tfplugingen-framework  ──▶  internal/gen/{resources,datasources}
        └──▶ tfgen registry         ──▶  internal/registry/registry_gen.go
```

`make generate` runs all of it. The spec is downloaded from `OPENAPI_SPEC_URL`,
which defaults to the tenant the Makefile names and can be pointed elsewhere.

The two middle steps are HashiCorp's
[OpenAPI provider spec generator](https://developer.hashicorp.com/terraform/plugin/code-generation/openapi-generator)
and [framework code generator](https://developer.hashicorp.com/terraform/plugin/code-generation/framework-generator).
The steps either side of them are [`tools/tfgen`](tools/tfgen):

- **`tfgen prep`** normalises the spec. The provider spec generator cannot map
  schema composition or a schema with no type, and the BWAN spec has both, so
  `oneOf` becomes the union of its object variants (required only where every
  variant agrees, enums unioned so a discriminator keeps working) and an untyped
  schema becomes a string holding a JSON document. Path parameters with a `-` are
  renamed to `_` so they survive as attribute names. It reports anything it could
  not normalise.
- **`tfgen registry`** turns `generator_config.yml` into the registry the provider
  reads at startup, and fails the build when the generators skipped something the
  configuration asked for — a resource silently missing from the provider is much
  harder to notice than a broken build.

Everything the runtime needs is derived from the configuration, so there is no
second place to keep in step:

- a resource whose read path has no `{id}` is refreshed by walking its collection
  and matching on `id` — the API has no single-object GET for every object;
- a resource with no update operation forces replacement on any change;
- a path placeholder other than `{id}` becomes a required attribute that forces
  replacement, because it identifies the parent the object lives under;
- a configurable attribute that is not part of the path is a request-body field
  for a resource and a query parameter for a data source;
- a data source addressing one object gains a `filter` argument when the API lets
  the collection that object belongs to be filtered;
- a data source reading a collection takes `filter` and `sort` and nothing else,
  and answers with `data` and `total_count`;
- a data source reading a collection and declaring `x_terraform.element: true`
  stands for one object out of it rather than the list: its schema comes from the
  collection's elements and the object is found by walking the collection, which
  is the only way an object the API never serves on its own gets a data source of
  its own.

## Objects the API describes as several shapes

Parts of the API describe one object as a choice between shapes — a tag is a
wanlink tag or an overlay tag; a cloud account's credentials are an AWS pair or
an Azure one. Terraform has no type for that, so there are two ways out and
`generator_config.yml` picks per object.

**A Terraform type per shape**, where the shapes are really different kinds of
object. An entry claiming a `variant` gets the shape to itself:

```hcl
resource "netskopebwan_tag_wanlink" "probe" {
  name = "probe"
  config = {
    type               = "wanlink"
    wan_link_frequency = 60       # only a wanlink tag has one
  }
}
```

Four kinds of tag become four resources, and `wan_link_frequency` is settable on
the one that has it rather than optional on all four and rejected by the API on
three. The API still serves them all from `/overlay-tags`, telling them apart by
`config.type`, so each type recognises its own objects: reading one by id checks
it really is that kind, and a list drops the others.

**A block per shape**, where the choice is about how one object is configured.
No claim is made, so the shapes become sibling blocks of which exactly one is set:

```hcl
resource "netskopebwan_cloud_account" "aws" {
  name           = "production"
  cloud_provider = "aws"
  config         = { aws = { key_id = "...", secret_access_key = "..." } }
}
```

The API declared those fields on the object itself, so the runtime splices the
set block back out when it builds a request and nests the matching one again when
it reads a response. `terraform validate` reports a set of blocks with none or
several set, which is the constraint the schema could not carry.

`netskopebwan_cloud_account` also shows the one rename in the provider: Terraform reserves
`provider` as an attribute name and rejects the whole provider over it, so the
field is exposed as `cloud_provider` and translated on the wire.

## Talking to the API

The engine reaches the API through one interface, [`bwanclient.API`](internal/bwanclient/api.go):
a request is a method, a path and a document, whichever object it belongs to.

That is also the test seam. `mockgen` generates a mock of it into
`internal/bwanclient/mock`, so a test states the requests it expects rather than
checking afterwards which ones arrived, and any request it did not ask for fails
it — which is the property worth having when the engine works out what to call
from a schema. `internal/bwanclient` keeps its own tests on a real HTTP server,
because there the HTTP layer *is* what is under test.

There is deliberately no generated client. `ogen` maps this spec fine on its own
terms, but it maps a `oneOf` to a Go sum type and the provider needs the flattened
union `tfgen prep` produces, so the two disagree about the shape of the same
schema. Bridging them would mean generated glue for all 299 operations whose only
job is to convert a document into a typed struct and straight back again. When the
first hand-written resource lands — a typed `netskopebwan_gateway` replacing
`netskopebwan_gateway_raw`, say — a generated client earns its keep, and it can sit behind
the same interface.

## Building

Enter the project shell first to get Go, Terraform, the HashiCorp generators,
`mockgen`, `golangci-lint`, `goreleaser`, Python and the local helper tools:

```bash
devenv shell
```

```bash
make generate         # every generator: spec download through mocks
make provider         # the plugin binary
make test             # tests, with the race detector
make lint
make terraform        # format the examples
make examples         # fill in any page's missing usage example
make docs             # the registry documentation, markdown
make docs-check       # fail if the committed docs are out of date
make docs-html        # the same thing as a browsable site
make docs-serve       # read it at localhost:8080
make ci               # everything above, in the order CI runs it
```

`make generate` has to come first in a fresh checkout: the packages `main.go`
imports do not exist until it has run, so nothing compiles before it.

`make docs` produces what the registry consumes, which is markdown. The usage
example and import command on each page come from `examples/`, and `make
examples` writes one for every object that has none — from the provider's own
schema, so an example is made of arguments that exist and is rewritten whenever
the schema moves. Every example is generated; removing the notice at the top of
one takes it out of the generator's hands and leaves it to be maintained
by hand. `docs-html` renders that into a site with a sidebar, a filter box and light and dark themes,
because 92 pages of raw markdown in a browser is barely better than reading the
files. `docs-serve` builds it and serves it, and takes a `PORT`.

Neither needs a terraform binary, a plugin directory or the network:
[`tools/tfdocs`](tools/tfdocs) reports the provider's own schema for
`tfplugindocs` to render, and renders the result to HTML.

Most generated files (`internal/gen`, `internal/bwanclient/mock`,
`internal/registry/registry_gen.go`, `openapi_tf_gen.yaml`,
`generator_config_gen.yml`, `provider_code_spec_gen.json`,
`provider_schema_gen.json`, `docs_html`) are build outputs and are not checked
in.

**`docs/` is the exception.** It is generated too, but the Terraform registry
serves it straight from the tagged tree, so it has to be committed. `make
docs-check` regenerates it and fails if the result differs from what is in git;
CI runs it on every push, so docs cannot drift from the schema.

## Using it

The provider is published as
[`netskopeoss/netskopebwan`](https://registry.terraform.io/providers/netskopeoss/netskopebwan):

```hcl
terraform {
  required_providers {
    netskopebwan = {
      source  = "netskopeoss/netskopebwan"
      version = "~> 0.0" # the released line; 1.x is still a prerelease
    }
  }
}
```

Running a 1.x prerelease takes two steps, and both name the same version — a
prerelease is not installed by `~>` or `>=`, and refuses to configure until the
version it is running is acknowledged. Acknowledging it accepts what a prerelease
is: nothing in that version's schema is covered by the provider's
backward-compatibility guarantees, and the next prerelease may rename, reshape or
remove any part of it. Substitute whichever prerelease you mean; `terraform init`
records what it installed in `.terraform.lock.hcl`, and an unacknowledged
prerelease says which version it wants in the error it raises:

```hcl
terraform {
  required_providers {
    netskopebwan = {
      source  = "netskopeoss/netskopebwan"
      version = "= 1.0.0-alpha.1"
    }
  }
}

provider "netskopebwan" {
  enable_pre_release = "1.0.0-alpha.1" # the same version, verbatim
}
```

`endpoint` and `token` default to `$BWAN_ENDPOINT` and `$BWAN_TOKEN`. The
endpoint is the API base URL without the version prefix; the provider appends
`/v2`. See [`examples/`](examples).

A list data source reads the whole collection: every page is walked, so `data`
holds the list and `total_count` is what the API reports for it. `filter` narrows
the list and `sort` orders it, and those are the only arguments there are — a page
of a list that has already been read in full is not worth asking for.

```hcl
data "netskopebwan_segments" "corporate" {
  filter = "name eq \"corporate\""
}
```

A data source addressing a single object takes either its `id` or a `filter` in
the API's filter syntax — exactly one of the two:

```hcl
data "netskopebwan_segment" "by_id" {
  id = "6501f0c2e4b0a1b2c3d4e5f6"
}

data "netskopebwan_segment" "by_name" {
  filter = "name eq \"corporate\""
}
```

A filter has to match exactly one object; matching none or several is an error
rather than a silent pick, so a configuration never depends on the order the API
returns things in.

Not every object has an endpoint that serves one of it — an app category, an audit
event and the software catalogue are only ever listed. Those data sources exist
anyway and are used the same way; the provider finds the object by walking the
collection, so an id that is not in it is an error rather than empty state.

## Objects the provider does not model in full

A few objects cannot be expressed properly yet. Most carry their whole
configuration in a field the API declares without a type — `device_config_raw`,
`policy_config_raw` and friends — which Terraform can only hold as a JSON string,
so it cannot validate or diff a single field of one and a change to any part of the
document is a change to the whole attribute. One is held up by something else: a
client template's probes accept either an object or the string it is shorthand for,
and only the object form is exposed.

Those objects carry a `_raw` suffix and are off until asked for:

| Object | Opt-in | Why |
| --- | --- | --- |
| `netskopebwan_gateway_raw`, `netskopebwan_gateways_raw` | `enable_raw_gateway` | the whole device configuration is opaque |
| `netskopebwan_gateway_template_raw`, `netskopebwan_gateway_templates_raw` | `enable_raw_gateway_template` | as above |
| `netskopebwan_policy_raw`, `netskopebwan_policies_raw` | `enable_raw_policy` | its policy, destination and score configuration is opaque |
| `netskopebwan_client_template_raw`, `netskopebwan_client_templates_raw` | `enable_raw_client_template` | the probe shorthand cannot be expressed |

**Enabling one is an explicit opt-out of compatibility.** These resources and data
sources are *not* covered by the provider's backward-compatibility guarantees,
their schemas *will* change without a major release, and they *will* be removed
once the objects can be modelled properly. The `_raw` suffix is what keeps the
plain name free for that: `netskopebwan_gateway` has already been taken by the typed form
of a gateway, and `netskopebwan_policy`, `netskopebwan_gateway_template` and `netskopebwan_client_template`
are still free for theirs.

Without the opt-in, planning one of these fails with a diagnostic naming the
argument to set — Terraform never gets as far as applying it.

`tfgen registry` enforces the naming: an object whose generated schema carries a
`*_raw` attribute has to be gated, and a gated object has to be suffixed so the
plain name stays free. An opaque *value* — a credentials blob, an audit diff — does
not make an object raw, because there is nothing better to roll out for it.

## Releasing

A release is a tag. Pushing one matching `v*` is what triggers
[the release workflow](.github/workflows/release.yml), which regenerates the
provider, checks the committed docs are current, then builds and signs the
release with [goreleaser](.goreleaser.yml).

```bash
make release VERSION=v1.2.3           # a release
make release VERSION=v1.2.3-alpha.1   # a prerelease
```

A prerelease is published as an ordinary GitHub release, not a GitHub
"pre-release". The flag would be nothing but cosmetic — Terraform installs a
prerelease only for an exact version, and the provider refuses to configure until
that version is acknowledged — and it stops the registry ingesting the version at
all, which is how `v1.0.0-alpha.1` came to exist on GitHub and nowhere else.

Nothing about that is undoable once the registry has picked the version up, so
the target refuses to run on anything it is not sure about. It requires the
version to be a `v`-prefixed semver, the working tree to be clean, the tag to be
absent both locally and on the remote, and `HEAD` to match its upstream — then
it prints what it is about to publish and makes you type the tag out before it
creates or pushes anything. If the push fails it removes the local tag so a
retry is not blocked by the tag it just made.

The registry only accepts a version whose assets are named for the repository,
so three files exist purely to satisfy it and are worth knowing about:

| Path | |
| --- | --- |
| `.goreleaser.yml` | builds the per-platform zips, the checksums and the detached GPG signature. `project_name` is pinned because the module path is not the repository name |
| `terraform-registry-manifest.json` | declares protocol `6.0`, which is what terraform-plugin-framework speaks. Published as `..._manifest.json` |
| `LICENSE` | the registry will not publish a provider without one |

The provider's name is not ours to choose: the registry derives it from the
repository, so `netskopebwan` is what prefixes every resource type, what
`tfplugindocs` writes into `docs/`, and what every release asset is named. It
lives in `PROVIDER_NAME` in the [`Makefile`](Makefile), `TypeName` in
[`internal/provider`](internal/provider/provider.go) and `address` in
[`main.go`](main.go).

## Layout

| Path | |
| --- | --- |
| `generator_config.yml` | what the provider exposes, and the API calls behind it |
| `tools/tfgen` | the spec normaliser and the registry generator |
| `tools/tfdocs` | reports the provider's own schema, and renders the docs to a browsable site |
| `internal/bwanclient` | the `API` interface, its JSON-over-HTTP implementation and the error sentinels |
| `internal/tfjson` | Terraform values ↔ JSON documents |
| `internal/tfschema` | a generated schema as a walkable model: request bodies, merging a response into state, and the shape a choice of forms takes |
| `internal/genresource` | the resource and data source built from a schema plus its API calls |
| `internal/provider` | provider arguments and registration |
| `CLAUDE.md`, `.claude/commands` | what to know before changing any of the above, and the flows for adding an object, regenerating, chasing spec drift, checking a change and cutting a release |
