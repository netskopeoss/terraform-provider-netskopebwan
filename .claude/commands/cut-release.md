---
description: Publish a release or prerelease, and confirm the registry actually took it
argument-hint: [version, e.g. v1.2.3 or v1.2.3-alpha.1]
---

Publish `$ARGUMENTS`. A release is a tag; pushing one matching `v*` triggers the
release workflow. It cannot be undone once the registry has ingested the version,
so confirm before starting and verify after finishing.

## 1. The tree has to be releasable first

```bash
make ci
```

The release workflow regenerates the provider and runs `docs-check` on the tagged
commit, so anything that fails locally fails there — after the tag exists, which
is the expensive time to find out.

## 2. Cut it

```bash
make release VERSION=$ARGUMENTS
```

The target refuses unless the version is a `v`-prefixed semver, the tree is
clean, the tag is absent locally and on the remote, and `HEAD` matches its
upstream. It then prints what it is about to publish and requires the tag typed
back. Do not work around any of that; if it refuses, fix what it names.

A prerelease is the same command with a semver prerelease suffix
(`v1.2.3-alpha.1`). Terraform will not install one without an exact version
constraint, and the provider refuses to configure until the practitioner
acknowledges that version with `enable_pre_release`.

## 3. Watch the run

```bash
gh run watch $(gh run list --workflow=release.yml --limit=1 --json databaseId --jq '.[0].databaseId')
```

If it fails, remember the workflow came from the **tagged commit**. Fixing
`.github/workflows/release.yml` on the branch does nothing for a tag that already
exists: either move the tag (only safe while no release was published — check
`gh release view <tag>`) or cut the next version.

## 4. Confirm the registry took it

The GitHub release existing is not the same as the version being installable:

```bash
gh release view $ARGUMENTS --json tagName,isPrerelease,isDraft,assets --jq \
  '{tag: .tagName, prerelease: .isPrerelease, draft: .isDraft, assets: (.assets | length)}'
curl -fsS https://registry.terraform.io/v1/providers/netskopeoss/netskopebwan/versions \
  | python3 -c "import json,sys; print(sorted(v['version'] for v in json.load(sys.stdin)['versions']))"
```

Expect the version to appear within a few minutes, with protocol `6.0` and every
platform. If it does not, check `isPrerelease` — a release flagged as a GitHub
pre-release is ignored by the registry, which is why `.goreleaser.yml` does not
set that flag. Nothing else about the flag matters: what keeps an alpha from being
installed by accident is its version and the `enable_pre_release` gate.

Note that the registry's "latest" pointer includes prereleases, so publishing an
alpha makes it the version shown on the provider's page even though
`terraform init` will not select it.

## 5. Report

The tag, the run's conclusion, whether the release is flagged as a prerelease,
the asset count, and whether the registry lists the version.
