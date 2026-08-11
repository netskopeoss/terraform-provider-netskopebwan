# terraform-provider-bwan

Terraform provider for Netskope Borderless WAN, generated from the BWAN v2
OpenAPI document that the REST gateway itself serves.

Nothing here is written per resource. Adding an object to the provider means
adding an entry to [`generator_config.yml`](generator_config.yml).

## How it is generated

```
//mgmt/go/cmd/rest-gateway/internal/v2/rest:openapi-build   the bundled v2 spec
        │
        ▼  tfgen prep
   openapi_tf_gen.yaml                                      spec the generators can map
        │
        ▼  tfplugingen-openapi + generator_config.yml
   provider_code_spec_gen.json                              provider code specification
        │
        ├──▶ tfplugingen-framework  ──▶  internal/gen/{resources,datasources}
        └──▶ tfgen registry         ──▶  internal/registry/registry_gen.go
```

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
  the collection that object belongs to be filtered.

## Objects the API describes as several shapes

Parts of the API describe one object as a choice between shapes — a tag is a
wanlink tag or an overlay tag; a link monitor's target is an fqdn, an ipv4 or an
ipv6. Terraform has no type for that, so there are two ways out and
`generator_config.yml` picks per object.

**A Terraform type per shape**, where the shapes are really different kinds of
object. An entry claiming a `variant` gets the shape to itself:

```hcl
resource "bwan_tag_wanlink" "probe" {
  name      = "probe"
  frequency = 60          # only a wanlink tag has one
}
```

Four kinds of tag become four resources, and `frequency` is settable on the one
that has it rather than optional on all four and rejected by the API on three.
The API still serves them all from `/tags`, so each type recognises its own
objects: reading one by id checks it really is that kind, and a list drops the
others.

**A block per shape**, where the choice is about how one object is configured.
No claim is made, so the shapes become sibling blocks of which exactly one is set:

```hcl
resource "bwan_link_monitor" "probe" {
  fqdn = { fqdn = "probe.example.com" }
}

resource "bwan_cloud_account" "aws" {
  name           = "production"
  cloud_provider = "aws"
  config         = { aws = { key_id = "...", secret_access_key = "..." } }
}
```

The API declared those fields on the object itself, so the runtime splices the
set block back out when it builds a request and nests the matching one again when
it reads a response. `terraform validate` reports a set of blocks with none or
several set, which is the constraint the schema could not carry.

`bwan_cloud_account` also shows the one rename in the provider: Terraform reserves
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
first hand-written resource lands — a typed `bwan_gateway` replacing
`bwan_gateway_raw`, say — a generated client earns its keep, and it can sit behind
the same interface.

## Building

Enter the project shell first to get Go, Terraform, the HashiCorp generators,
`mockgen`, `golangci-lint`, Python and the local helper tools:

```bash
devenv shell
```

```bash
heph run //mgmt/tf-provider:provider          # the plugin binary
heph query //mgmt/tf-provider/... | grep go_test | heph run -
heph run //mgmt/tf-provider:lint
heph run //mgmt/tf-provider:terraform-fix     # format the examples
heph run //mgmt/tf-provider:docs              # the registry documentation, markdown
heph run //mgmt/tf-provider:docs-html         # the same thing as a browsable site
heph run //mgmt/tf-provider:docs-serve        # read it at localhost:8080
```

`:docs` produces what the registry consumes, which is markdown. `:docs-html`
renders that into a site with a sidebar, a filter box and light and dark themes,
because 92 pages of raw markdown in a browser is barely better than reading the
files. `docs-serve` builds it and serves it, and takes a `PORT`.

None of it needs a terraform binary, a plugin directory or the network:
[`tools/tfdocs`](tools/tfdocs) reports the provider's own schema for
`tfplugindocs` to render, and renders the result to HTML.

Everything generated (`internal/gen`, `internal/bwanclient/mock`,
`internal/registry/registry_gen.go`, `openapi_tf_gen.yaml`,
`generator_config_gen.yml`, `provider_code_spec_gen.json`,
`provider_schema_gen.json`, `docs`, `docs_html`) is a build output and is not
checked in.

## Using it

`endpoint` and `token` default to `$BWAN_ENDPOINT` and `$BWAN_TOKEN`. The
endpoint is the API base URL without the version prefix; the provider appends
`/v2`. See [`examples/`](examples).

A list data source walks every page by default, so `data` holds the whole
collection. Setting `first` or `after` asks for exactly one page instead.

A data source addressing a single object takes either its `id` or a `filter` in
the API's filter syntax — exactly one of the two:

```hcl
data "bwan_segment" "by_id" {
  id = "6501f0c2e4b0a1b2c3d4e5f6"
}

data "bwan_segment" "by_name" {
  filter = "name eq \"corporate\""
}
```

A filter has to match exactly one object; matching none or several is an error
rather than a silent pick, so a configuration never depends on the order the API
returns things in.

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
| `bwan_gateway_raw`, `bwan_gateways_raw` | `enable_raw_gateway` | the whole device configuration is opaque |
| `bwan_gateway_template_raw`, `bwan_gateway_templates_raw` | `enable_raw_gateway_template` | as above |
| `bwan_policy_raw`, `bwan_policies_raw` | `enable_raw_policy` | its policy, destination and score configuration is opaque |
| `bwan_client_template_raw`, `bwan_client_templates_raw` | `enable_raw_client_template` | the probe shorthand cannot be expressed |

**Enabling one is an explicit opt-out of compatibility.** These resources and data
sources are *not* covered by the provider's backward-compatibility guarantees,
their schemas *will* change without a major release, and they *will* be removed
once the objects can be modelled properly. The `_raw` suffix is what keeps the
plain name free for that: `bwan_gateway` has already been taken by the typed form
of a gateway, and `bwan_policy`, `bwan_gateway_template` and `bwan_client_template`
are still free for theirs.

Without the opt-in, planning one of these fails with a diagnostic naming the
argument to set — Terraform never gets as far as applying it.

`tfgen registry` enforces the naming: an object whose generated schema carries a
`*_raw` attribute has to be gated, and a gated object has to be suffixed so the
plain name stays free. An opaque *value* — a credentials blob, an audit diff — does
not make an object raw, because there is nothing better to roll out for it.

## What is not exposed

- **Site commands.** Running a command is an action, not a managed object.
- **A client's `device_config_raw`.** Deliberately dropped. Note the spec marks it
  required on `POST /clients`, so the API rejects creating a client until that
  changes; reading and updating one works.
- **A tenant's `has_active_network`.** Read-only status, deliberately dropped.
- **Gateway sub-actions** (activation tokens, exec, passwords, telemetry) and the
  other read-only report endpoints that have no object behind them.
- **The shorthand form of a `oneOf` that mixes an object with a scalar.** Only the
  object form is exposed; it can always express the shorthand too.

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
