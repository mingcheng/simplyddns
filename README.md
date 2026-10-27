# Simply DDNS - Yet Another Universal DDNS Tools

- [Concepts](#concepts)
- [Installation](#installation)
  - [Build with mise](#build-with-mise)
  - [docker-compose](#docker-compose)
- [Configuration](#configuration)
- [Extensions](#extensions)
  - [source](#source)
  - [target](#target)
- [FAQ](#faq)
  - [WebHook Configuration](#webhook-configuration)

SimplyDDNS is a universal DDNS tool that can be used to update domain name records based on obtained IP addresses. It is designed to be flexible and easy to extend. A Chinese（简体中文） version is available in [README_ZH.md](README_ZH.md).

This tool simply monitors and updates domain name records based on obtained IP addresses. [While there are similar existing projects](https://github.com/mingcheng/ddns-go) that are already very useful, there were several points that didn't meet my personal expectations and requirements:

1. I don't need a UI interface; instead, I prefer a fully functional configuration file
2. Need more flexibility and freedom in configuration combinations
3. As a server-side application, more detailed logging and ability to transmit required information

Since this is actually a strong requirement (everyone knows that when setting up a website, the last step is changing DNS, which is very troublesome and error-prone),
this application was created. While it may have a relatively higher configuration threshold, it offers more flexibility.

## Concepts

[Dynamic DNS registration (DDNS)](https://en.wikipedia.org/wiki/Dynamic_DNS) behavior is abstracted into two types: how to obtain IP addresses and how to configure DNS.

Therefore, by implementing these two interfaces and matching them with configurations, you can flexibly arrange and run them. See the Extensions section for details.

## Installation

The release workflow produces archives for Linux amd64/arm64, macOS amd64/arm64 and Windows amd64, together with `checksums.txt`. Each archive contains the executable, license, documentation and a safe `simplyddns.yml` example. Verify downloaded archives with `shasum -a 256 -c checksums.txt` before extracting them.

### Build with mise

Project tool versions are pinned in [mise.toml](mise.toml). Use [mise](https://mise.jdx.dev/) to install or update tools; Go must not download an unmanaged replacement toolchain (`GOTOOLCHAIN=local`).

```sh
mise trust
mise install
mise exec -- make build
./simplyddns --version
./simplyddns --config example/basic.yml
```

The last command runs an offline loopback/sleep demonstration. To install the executable into Go's binary directory, use `mise exec -- make install`.

### docker-compose

With a running Docker engine and the Compose plugin:

```sh
mise exec -- make build_docker_image
mise exec docker-cli -- docker run --rm simplyddns:2.0.0 --version
mise exec docker-cli -- docker compose up --build -d
```

[compose.yml](compose.yml) mounts the safe example configuration read-only. Replace that mount with your own configuration for real DNS updates. The container runs as an unprivileged user; log to stderr or provide a writable log destination. Go is installed through mise inside the [Dockerfile](Dockerfile), using the same project version as local builds.

## Configuration

See [example/](example/) for complete configurations. Select a file with `--config`; otherwise SimplyDDNS searches `/etc/simplyddns`, `$HOME/.config/simplyddns`, then the working directory for `simplyddns.yml` or `simplyddns.yaml`.

A minimal configuration, from [basic.yml](example/basic.yml):

```yaml
logfile: "/dev/stderr"
debug: true

ddns:
  - source:
      type: "lo"
      interval: 60 # 1 minute
    target:
      type: "sleep"
```

This configuration file means getting the IP address from the lo module and passing it to the sleep target for execution. The sleep target does nothing except sleep for a few seconds, but this demonstrates SimplyDDNS's working mechanism.

Intervals are in seconds: the default is 60 and values below 30 are clamped to 30. The first check waits for the interval. Each cycle has a 30-second deadline, and built-in operations honor shutdown cancellation. Invalid job definitions and domain names are rejected instead of being retried indefinitely.

Comma-separated source names are supported, for example `ipify,ipsb`. Successful sources must agree; larger groups also reject excessive failures. Configure IPv4 and IPv6 in separate jobs: targets select A or AAAA records from the supplied address. Successful updates are cached until the source address changes; failed updates are retried on a later tick.

Logs go to stderr by default. `/dev/stdout` and `/dev/stderr` use the existing standard streams, including in read-only containers. Other log files are closed when the service stops.

## Extensions

Extensions register a `SourceFunc` or `TargetFunc`; see [func.go](func.go), [source/lo.go](source/lo.go) and [target/sleep.go](target/sleep.go). Implementations must honor the supplied context and return errors for invalid input or failed operations.

### source

| Name         | Address source                                |
| ------------ | --------------------------------------------- |
| `lo`         | Local loopback `127.0.0.1`, for testing       |
| `static`     | An IPv4 or IPv6 address from `source.content` |
| `cloudflare` | Cloudflare trace endpoint                     |
| `cz88`       | CZ88 JSON endpoint                            |
| `ipify`      | ipify IPv4 endpoint                           |
| `ipinfo`     | IPinfo plain-text endpoint                    |
| `ipip`       | IPIP JSON endpoint                            |
| `ipplus360`  | IPPlus360 JSON endpoint                       |
| `ipsb`       | IP.SB IPv4 endpoint                           |
| `myip`       | `api.my-ip.io` plain-text endpoint            |
| `opendns`    | OpenDNS IPv4 endpoint                         |

An endpoint may only support one address family; do not combine sources that return different families in the same job.

### target

_sleep_

Does nothing but sleep for a few seconds, can be used for testing

_alidns_

Uses Aliyun DNS. `key` is the access key ID and `token` is the access key secret. See [example/alidns.yml](example/alidns.yml).

_namedotcom_

Uses Name.com. `key` is the username and `token` is the API token. See [example/namedotcom.yml](example/namedotcom.yml).

_cloudflare_

Uses Cloudflare. `key` is the zone ID and `token` is an API token with DNS edit permissions for that zone. Existing records are updated in place, preserving their settings. See [example/cloudflare.yml](example/cloudflare.yml).

All three DNS targets support an optional SOCKS5 `proxy` in `host:port` form.

## FAQ

_If I only have one way to request IP addresses but multiple domains, repeating configurations makes it look bloated. Is there a better solution?_

This is a common scenario. You can consider using Yaml's reference feature to reduce repeated configurations, for example:

```yaml
defaults: &defaults
  type: lo
  interval: 60
```

First define returning a local address from lo every minute, then configure the corresponding domain updates:

```yaml
ddns:
  - source:
      <<: *defaults
    target:
      type: "sleep"
      domains:
        - a.com
        - b.com
  - source:
      <<: *defaults
    target:
      type: "another"
      domains:
        - c.com
        - d.com
```

After configuring this way, you only need to focus on the target (domain) configuration.

### WebHook Configuration

A webhook runs after a successful address update. Configure it inside the corresponding job:

```yaml
ddns:
  - source:
      type: lo
      interval: 60
    target:
      type: sleep
    webhook:
      url: https://example.com/webhook
      token: "<optional-bearer-token>"
```

After configuring like this, when the address changes, the results and changed domain results will be POSTed to the corresponding address. The result format is as follows:

```json
{
    "address": "127.0.0.1",
    "domains": "a.com,b.com",
    "now": "2022-07-13T15:00:10.931809+08:00"
}
```

If there are multiple domain results, they will be separated by commas.

The same address and domain list are sent in the `Address` and `Domains` headers. Basic authentication is also available through `username` and `password`. Use HTTPS for credentials. Any 2xx response is successful; failures are logged without rolling back or repeating the DNS update.

`- eof -`
