# Releasing tavrel

This is the maintainer guide for cutting a release.

## Versioning

tavrel uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html): `MAJOR.MINOR.PATCH`.

- **MAJOR** — incompatible changes. Existing users must change configuration, scripts, the Docker setup, or accept a manual database migration.
- **MINOR** — new functionality, backward compatible.
- **PATCH** — backward-compatible bug fixes, performance work, dependency updates, docs.

While tavrel is on `0.x`, the public contract is not stable. Breaking changes bump the **minor** version (`0.1.0 → 0.2.0`) and must be called out in the release notes. Fixes bump the patch version (`0.2.0 → 0.2.1`). `1.0.0` is a deliberate promise of backward compatibility, not a feature milestone.

For an application, the public contract is:

- the HTTP API (`/api/v1` response shapes, status codes, auth behavior);
- OTLP ingest (endpoints, ports, supported protocols);
- environment variables, their defaults and semantics;
- the Docker interface (image name, ports, `/data` volume, UID, entrypoint);
- the database schema and upgrade path;
- the CLI (`version`, `healthcheck`).

Internal Go packages are not part of the contract.

### Deciding the bump

Ask: **can an existing user pull the new image, restart with their existing compose, `.env`, and volume, and keep working without changes or data loss?**

- Yes, and it only fixes something → patch.
- Yes, and it adds something → minor.
- No → major (minor while on `0.x`).

| Change | Bump |
| --- | --- |
| Bug fix, performance, docs, internal refactor, dependency bump | patch |
| New feature, endpoint, filter, env var, platform, additive migration | minor |
| Remove/rename an env var, CLI flag, or API field; change a response shape; destructive or manual migration; change the volume path, ports, or entrypoint; drop a platform | major (minor on 0.x) |

Changing a default is not automatically breaking, but it changes behavior. Document it; use major if it can cause data loss or requires user action.

## Release checklist

1. Ensure `master` is green and contains everything for the release.
2. Review the commits since the last tag: `git log v0.1.0..HEAD --oneline`.
3. Decide the version bump using the rules above.
4. Update `CHANGELOG.md`: move `Unreleased` entries under the new version and date.
5. Commit the release metadata: `chore(release): vX.Y.Z`.
6. Tag and push:
   ```sh
   git tag -a vX.Y.Z -m "vX.Y.Z"
   git push origin master
   git push origin vX.Y.Z
   ```
7. The release workflow runs tests, builds and pushes the multi-arch image, and creates the GitHub Release.
8. Verify on a clean machine:
   ```sh
   TAVREL_VERSION=X.Y.Z docker compose up -d
   ```
   Check the UI, OTLP ingest, the healthcheck, and that the volume survives a restart.
9. Test an upgrade from the previous release using a copy of its database.
10. Make the GHCR package public if this is the first push, and announce the release.

## Docker tags

For a stable `vX.Y.Z`:

- `X.Y.Z` — exact version; recommended for production.
- `X.Y` — newest patch in that minor.
- `latest` — newest stable release.

For a pre-release such as `vX.Y.Z-rc.1`, publish only `X.Y.Z-rc.1`. Do not move `X.Y` or `latest`.

Tags are immutable. If a release is broken, publish a new patch; do not move or delete the old tag.

## Release candidates

Use a release candidate for a breaking change, a database migration, or a change to the Docker interface:

```sh
git tag -a v0.2.0-rc.1 -m "v0.2.0-rc.1"
git push origin v0.2.0-rc.1
```

Mark it as a pre-release, test the upgrade path from the previous stable version, then tag `v0.2.0`.

## Hotfixes

If the fix does not depend on unreleased work, branch from the release tag:

```sh
git checkout -b hotfix/0.2.1 v0.2.0
# fix, test
git tag -a v0.2.1 -m "v0.2.1"
git push origin hotfix/0.2.1 v0.2.1
```

Merge or cherry-pick the fix back to `master`. If `master` is already releasable, fixing there and tagging a patch is also fine.

## Rollback

Publish a patch release. Users can pin the previous image tag:

```sh
TAVREL_VERSION=0.1.0 docker compose up -d
```

Database migrations are forward-only. If an upgrade migrated the database, rolling back the binary is not enough; restore the backup taken before the upgrade. This is why the release checklist requires a backup.

## Version metadata

The binary version is injected at build time:

- `internal/version.Version`
- `internal/version.Commit`
- `internal/version.Date`

The release workflow passes `VERSION`, `COMMIT`, and `DATE` as Docker build args, and the image carries `org.opencontainers.image.*` labels. `tavrel version` prints the embedded values; use it when triaging an issue.
