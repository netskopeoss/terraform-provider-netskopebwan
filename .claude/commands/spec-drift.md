---
description: Diagnose a generation failure caused by the upstream OpenAPI spec changing
---

Generation broke, or you want to know what upstream changed. The spec is fetched
live from `OPENAPI_SPEC_URL` and is not pinned, so drift is routine — always
establish what the spec says now before touching `generator_config.yml`.

## 1. Diff the configuration against the spec

`openapi.json` is written by `make generate`; download it directly if generation
fails before that. Then compare every path the configuration names against the
paths the spec has:

```bash
python3 - <<'EOF'
import json, re
live = json.load(open("openapi.json"))["paths"]
cfg = open("generator_config.yml").read()
paths = sorted(set(re.findall(r'^\s*path:\s*(\S+)', cfg, re.M)))
print("missing from the spec:")
for p in paths:
    if p not in live:
        print("  ", p)
print("\nin the spec, unused by the provider:")
for p in sorted(live):
    if p not in paths:
        print("  ", p)
EOF
```

Read both lists. The second one is where a rename hides: `/tags` disappearing at
the same moment `/overlay-tags` appears is one change, not two.

Note that prep renames dashed path parameters, so a spec path
`/address-groups/{group-id}/...` legitimately appears in the configuration as
`{group_id}`. Those are not missing.

## 2. Classify each difference

- **Renamed** — repoint the entry. Then check the schema too: an endpoint that
  moved may also have been restructured, and a renamed path with a reshaped body
  is two changes to make.
- **Removed** — the resource or data source has to go, along with its `docs/`
  pages and `examples/` directory. That is a breaking change for practitioners;
  say so plainly in the summary and the commit message.
- **Restructured** — the path is there but the body moved. Read the component
  schemas; if a discriminated union moved into a nested object, the variant
  machinery needs the nested path, not just a new endpoint.

## 3. Do not fix it in the wrong layer

`tfgen prep` normalises the spec so the HashiCorp generator can map it. If the
spec is merely awkward — composition, an untyped schema, a discriminator naming a
property no branch declares — that belongs in prep, with a test in
`tools/tfgen/prep_test.go`. Only genuine endpoint changes belong in
`generator_config.yml`.

## 4. Verify and report

```bash
make ci
```

Report every difference found, what you did about each, and anything that looks
like an upstream mistake rather than an intentional change — four endpoint
families disappearing at once is worth a question to whoever owns the spec before
the provider drops the resources.
