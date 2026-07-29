# Security policy

## Supported versions

| Version | Supported |
|---|---|
| 1.0.x | ✅ |
| < 1.0 | ❌ |

Anything before 1.0 was development code. It served every endpoint without
authentication and had several injection paths reachable from an ordinary
request; the [1.0 audit](RELEASE-1.0.md) lists them. There are no fixes for
those versions — upgrade.

## Reporting a vulnerability

Report privately, not as a public issue.

- **Preferred**: [open a private advisory](https://github.com/arturoeanton/goscim/security/advisories/new)
  through GitHub.
- **Alternatively**: email <arturoeanton@gmail.com> with `GoSCIM security` in
  the subject.

Useful to include: the version or commit, what an attacker gains, the smallest
reproduction you have, and your own assessment of the impact.

This is a volunteer-maintained project, so treat these as intentions rather
than guarantees: acknowledgement within a week, an assessment within two, and a
fix for anything confirmed as soon as it can reasonably be made. You will be
credited in the advisory unless you prefer otherwise.

Please allow a reasonable window before disclosing publicly.

## Scope

In scope: authentication or authorization bypass, injection through filters,
`sortBy` or any other client-controlled input, information disclosure across
callers, and remote crashes.

Out of scope, because they are documented behaviour rather than accidents —
see the [security guide](docs/security.md):

- **Passwords are stored as sent.** There is no hashing.
- **`SCIM_AUTH=none` serves every request.** Opt-in, and logs a warning.
- **`SCIM_COUCHBASE_TLS_SKIP_VERIFY=true` accepts any certificate.** Opt-in,
  and logs a warning.
- **There is no rate limiting.** Put it in the proxy.
- **The uniqueness check is read-then-write**, so two concurrent creates of the
  same value can both pass.

If you think one of those is worse than the documentation claims, report it
anyway — the classification is a judgement, not a rule.

## Hardening

Before exposing GoSCIM to anything real, work through the
[checklist](docs/security.md#checklist).
