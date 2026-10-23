# ReClash core patches

**English** · [Русский](PATCHES.ru.md) · [Upstream README](README.md)

This is a fork of [MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo), used as the
proxy core inside the [ReClash app](https://github.com/Hoxiee/ReClash). The current upstream
base is **`v1.19.33`**, with FlClash host-integration patches and ReClash-specific additions
on top. This page describes those patches; the upstream project overview stays in `README.md`.

## How the patches are shaped

Most integration additions expose a hook: a package-level function variable, such as
`UrlTestHook`, which the ReClash app wires up at startup. Reporting hooks do nothing until
the host installs a callback. This keeps observation separate from the core's networking work.

Not every patch is passive. The host owns listener and TUN lifecycle, can take over selected
provider health checks, and can opt into the direct-route override described below. This fork
is therefore not a drop-in standalone replacement for upstream mihomo.

## What changed and why

**Observability hooks.** The app needs to see what the core is doing — for connection
diagnostics, the traffic UI, and node health. Hooks report URL-test results, dial results,
provider-loaded events, geo-database update status, and first-byte progress. These hooks
report events; they do not choose routes.

**Flow evidence.** The event stream in [`tunnel/evidence.go`](tunnel/evidence.go) reports
connection stages: ingress, route resolution, dial start, and dial completion, including
failure paths, timing, and coarse error classes. It feeds the app's connection diagnostics
without intervening in the connection.

**Route, endpoint, and diversity identity.** Each proxy exposes three derived fingerprints:

- A *route fingerprint* hashes the proxy configuration while excluding the display name,
  so renaming a node does not change its identity.
- An *endpoint fingerprint* uses the type, server, port, SNI, and transport host/path without
  credentials. Rotating a UUID, short ID, or password does not by itself change this identity.
- A *diversity fingerprint* describes transport, TLS, Reality, and obfuscation characteristics.
  Its SNI component is represented by a hash.

These let the app recognise and compare nodes without relying on their display names.
See [`adapter/route_identity.go`](adapter/route_identity.go).

**Traffic split (proxy vs direct).** The statistics manager tracks proxied traffic separately
from direct traffic, so the UI can show how much actually went through a proxy.

**Battery-friendly health checks.** Provider health checks can be deferred while the Android
screen is off and resumed when it wakes. The app can also take over automatic checks for
selected providers with `SetAutoHealthCheckSuppressed`, avoiding duplicate probes while
leaving other providers' checks in place. See [`adapter/provider/patch.go`](adapter/provider/patch.go).

**Direct split.** `tunnel.DirectSplitHook` lets the host redirect a connection matched by an
explicit `DIRECT` rule to a proxy group. The final `MATCH` rule is never split, and UDP support
is checked before accepting the replacement. The host decides when to enable the override;
without the callback, it is inactive. See [`tunnel/patch.go`](tunnel/patch.go).

**Smarter geo updates.** When a remote SHA-256 digest is available, the updater compares it
with the local GeoIP/GeoSite database and skips downloading an unchanged file. The host can
start and stop the geo updater and receive update-status callbacks. See
[`component/updater/patch.go`](component/updater/patch.go).

**Lifecycle handed to the host.** Recreating regular listeners and TUN on configuration apply,
and coordinating provider readiness, are no longer owned by the upstream executor path.
The ReClash host coordinates start, stop, and reload. `listener.StopListener()` provides
listener cleanup, and `sing_tun.NewStack` exposes a TUN-stack override point for platform
integration. Building this repository's root executable alone does not produce the runtime
shipped with ReClash; use the [application's build tooling](https://github.com/Hoxiee/ReClash#build-from-source).

## What we did *not* change

- Proxy protocol and encryption implementations remain based on upstream mihomo.
- Diagnostic hooks report to the host application, rather than an external reporting service.
- The upstream README content and `LICENSE` are preserved; only links to these patch notes
  are added above the upstream README.

Routing is not claimed to be identical in every case: the optional direct-split hook can
change the selected outbound, and lifecycle integration deliberately differs from upstream.

## Relationship to upstream

- Base: upstream tag [`v1.19.33`](https://github.com/MetaCubeX/mihomo/releases/tag/v1.19.33).
- Fork branch: [`ReClash`](https://github.com/Hoxiee/mihomo/tree/ReClash).
- The patch stack contains inherited FlClash integration and ReClash additions. Updating the
  upstream base requires rebasing and checking these integration contracts.
- The Go module path remains `github.com/metacubex/mihomo` for import compatibility.
- Application-side bindings and routing policy live in
  [ReClash's core integration](https://github.com/Hoxiee/ReClash/tree/dev/core).

## License

Same as upstream mihomo: GPL-3.0. See [`LICENSE`](LICENSE) and the upstream naming notice
in [`README.md`](README.md#license).
