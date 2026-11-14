# Contributing to GoSCIM

Thanks for considering it. This page covers what you need to get a change
merged.

## Getting set up

Requirements: Go 1.26 or later (1.27.4+ recommended), Docker (for the integration suite), Git.

```bash
git clone https://github.com/arturoeanton/goscim.git
cd goscim
make check
```

`make check` builds, vets and runs the unit suite with the race detector. It
needs no database: the unit tests run the real router over an in-memory store.

For the integration suite, which starts Couchbase in a container:

```bash
make integration
```

It takes a few minutes on the first run while the image downloads.

[docs/development.md](docs/development.md) has the code layout, the test
helpers, and how to regenerate the filter parser.

## What to work on

Most gaps recorded in [RELEASE-1.0.md](RELEASE-1.0.md) (`POST /Bulk`, value
filters in search and `PATCH`, `attributes` / `excludedAttributes`,
`POST /.search`, password hashing, `/healthz` and `/readyz`, structured audit
logging, EC/EdDSA JWKS keys, atomic uniqueness, `Group.members` ↔ `User.groups`
synchronisation, and Microsoft Entra ID / Okta / Keycloak conformance) have
since landed in 1.1.0, 1.2.0, and 1.3.0. Useful areas still open:

- Prometheus metrics endpoint (`/metrics`)
- Interoperability testing and field reports against additional identity
  providers (such as PingFederate, OneLogin, and Google Workspace)
- Secondary Couchbase index definitions for large deployments

Issues labelled `good first issue` and `help wanted` are also a reasonable
starting point. If you are planning something substantial, open an issue first
so nobody duplicates the work.

### Translations

Documentation is English-only. Earlier releases carried partial translations
into six languages; every one of them fell behind the code, and documentation
that has fallen behind is worse than none, because a reader follows it.

If you want to maintain a translation, open an issue proposing it. What matters
is a commitment to keep it in step with the English version, not the initial
translation.

## Reporting a bug

Include the version or commit, what you expected, what happened, and the
smallest way to reproduce it. Relevant configuration — the `SCIM_AUTH` scheme,
the resource type, the schema — usually matters more than the stack trace.

**Security issues do not go in the tracker.** See [SECURITY.md](SECURITY.md).

## Making a change

1. Branch from `main`.
2. Write the change, and a test that fails without it.
3. `make check`, and `make integration` if you touched storage, filters, or
   startup.
4. Open a pull request explaining what problem it solves.

### Tests

A change to behaviour needs a test. Two things this project has learned the
hard way and would rather not relearn:

**Drive the router, not the function.** The HTTP surface is the contract
clients depend on, and it catches middleware ordering mistakes a direct call
never will. `newTestServer(t)` gives you the real router over an in-memory
store.

**Check that the test can fail.** Break the code on purpose and confirm the
test goes red. The 1.0 work found two tests that passed for the wrong reason —
one asserted a `200` from `sortBy` without checking the ordering, another
exercised the JWT library's type safety rather than the algorithm allow-list it
was named after. A green suite is not evidence until you have seen it go red.

Anything touching the N1QL filter translation, bucket creation or Couchbase
query behaviour also belongs in the integration suite (`test/integration/`),
even though `MemoryStore` evaluates filters in memory for unit conformance
suites.

### Style

- `gofmt`. CI checks it, generated parser files excepted.
- Comments explain **why**, not what. If the reason is not obvious from the
  code, write it down; if it is, do not.
- Errors get context: `fmt.Errorf("creating bucket %q: %w", name, err)`.
- Storage errors are normalised at the `Store` boundary — handlers never import
  gocb.
- Keep `name` and `endpoint` straight: `name` is the Couchbase bucket,
  `endpoint` is the URL path.

### Commits

One logical change per commit, with a message that says what problem it solves
rather than what lines moved. If the reasoning behind an approach is not
obvious, the commit message is the right place for it.

## Pull requests

CI runs build, `gofmt`, `go vet`, the race-detector suite, coverage,
`govulncheck` and a `go mod tidy` check, plus the integration suite. All of it
has to pass.

In the description, say what problem the change solves, how you tested it, and
anything you decided against and why. If it changes the wire contract, say so
explicitly — that goes in the CHANGELOG.

Review is a conversation. Expect questions; ask your own.

## Documentation

Code changes that alter behaviour need the documentation updated in the same
pull request. Stale documentation is a bug, and it is the kind that survives
for years.

The English documentation under `docs/` is authoritative. `README.md` should
stay short and accurate; the detail belongs in `docs/`.

## Code of conduct

Be decent. Assume good faith, keep criticism about the work, and take
disagreements to the technical merits. Harassment of any kind is not welcome
here, and the [GitHub Community Guidelines](https://docs.github.com/en/site-policy/github-terms/github-community-guidelines)
apply.

## Getting help

- [Discussions](https://github.com/arturoeanton/goscim/discussions) for
  questions and ideas
- [Issues](https://github.com/arturoeanton/goscim/issues) for bugs and features
- [docs/](docs/) for everything else
