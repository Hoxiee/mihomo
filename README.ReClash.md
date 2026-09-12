# ReClash core patches

This is a fork of [MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo), used as the
proxy core inside the ReClash app. It tracks the upstream release tag **`v1.19.31`** and
adds a thin layer of patches on top of it. Nothing here rewrites how mihomo routes or
proxies traffic — the changes are almost entirely *additive*.

If you're wondering "what's actually different from stock mihomo", this file is the honest
short answer.

## How the patches are shaped

Almost every change follows the same pattern: the core exposes a hook (a package-level
function variable, e.g. `var UrlTestHook ...`), and the ReClash app wires a callback into
it at startup. Until the app sets those, the hooks do nothing. So the core still behaves
like plain mihomo unless the host app opts in.

That keeps the fork easy to reason about and easy to rebase onto a newer upstream tag.

## What changed and why

**Observability hooks.** The app needs to see what the core is doing — for the connection
doctor, the traffic UI, and node health. We added hooks for URL-test results, dial
results, provider-loaded events, geo-database update status, and a first-byte progress
signal. These only report; they don't change routing.

**Flow evidence.** A small event stream (`tunnel/evidence.go`) reports each connection
through a few stages (ingress → route resolved → dial started → dial finished, plus the
failure branches) with a coarse error class. This feeds the app's connection doctor. It's
a passive witness — it observes, it doesn't intervene.

**Route & diversity identity.** Each proxy gets two derived tags:
- a *route fingerprint* — a hash of the proxy config that ignores the display name, so
  renaming a node doesn't change its identity;
- a *diversity fingerprint* — a compact label of transport/TLS/obfs characteristics.
  The SNI is hashed to an opaque class, never stored in the clear.
These let the app group and reason about nodes without leaking secrets.

**Traffic split (proxy vs direct).** The statistics manager now tracks proxied traffic
separately from direct traffic, so the UI can show how much actually went through a proxy.

**Battery-friendly health checks.** On Android, when the screen is off, provider health
checks are deferred until the screen wakes, to avoid waking the radio while idle.

**Smarter geo updates.** Before downloading a multi-megabyte GeoIP/GeoSite database, the
updater fetches the remote checksum and skips the download if the local copy already
matches. The geo updater is also cancellable and driven by the host.

**Lifecycle handed to the host.** A few things the core used to do on its own — recreating
listeners/TUN on config apply, waiting for all providers to load — were moved out so the
ReClash app owns start/stop/reload. There's also a `StopListener()` helper and a Linux
gVisor TUN-stack override point (`sing_tun.NewStack`).

**CMFA → Android flag.** The upstream `cmfa` build flag was renamed to `Android` and the
build tags simplified accordingly. Pure rename, no behavior change.

## What we did *not* change

- No changes to proxy protocols, encryption, or how traffic is routed.
- No telemetry, no phone-home. The hooks report to the local app only.
- No edits to upstream's own `README.md` or licensing.

## Relationship to upstream

- Base: upstream tag `v1.19.31`.
- The patches sit as a short linear stack on top of that tag, so bumping to a newer
  upstream release is a manual rebase of this layer.

## License

Same as upstream mihomo. See `LICENSE`.
