# SimplyDDNS

[简体中文](README_ZH.md)

Keep your domain pointed at your current public IP, so your home server, NAS or other services remain easy to reach when the address changes.

- **Simple to run**: manage everything in a YAML file, without a web interface.
- **Flexible address discovery**: use automatic checks across multiple services, or restrict requests to servers you choose.
- **Multiple domains and providers**: support IPv4 and IPv6 with Cloudflare, Aliyun DNS and Name.com.
- **Updates only when needed**: avoid unnecessary DNS changes, with logs and optional webhook notifications.

## Quick start

### 1. Install

Install with [Homebrew](https://brew.sh/):

```sh
brew install mingcheng/tap/simplyddns
```

With Homebrew, use `simplyddns` instead of `./simplyddns` in the commands below.

Alternatively, download the archive for Linux, macOS or Windows from [Releases](https://github.com/mingcheng/simplyddns/releases) and extract it. The commands below use Linux/macOS syntax; on Windows, run `simplyddns.exe`.

<details>
<summary>Build from source with mise</summary>

Install [mise](https://mise.jdx.dev/), then run:

```sh
git clone https://github.com/mingcheng/simplyddns.git
cd simplyddns
mise trust
mise install go make
mise exec go make -- make build
```

</details>

### 2. Configure

Create `simplyddns.yml`, or replace the demo configuration included in the release archive. This example updates two domains through Cloudflare; replace the credentials and domains with your own:

```yaml
logfile: "/dev/stderr"

ddns:
  - source:
      type: auto
      interval: 300
    target:
      type: cloudflare
      key: "<your-zone-id>"
      token: "<your-api-token>"
      domains:
        - home.example.com
        - nas.example.com
```

The Cloudflare token needs DNS edit permissions for the specified zone. For other providers, see [DNS providers](#dns-providers).

### 3. Run

```sh
./simplyddns --version
./simplyddns --config simplyddns.yml
```

Keep the process running to check and update your records. The first check starts after one configured interval.

To try SimplyDDNS without modifying DNS, use the [offline demo](example/basic.yml) or [automatic address discovery demo](example/auto.yml). The release archive includes the offline demo as its default configuration.

## Configuration

Each entry under `ddns` is a job: `source` chooses how to get an address, and `target` chooses where to update DNS. Add domains to the same job when they share a provider and credentials; use separate jobs for different accounts, providers, or IPv4/IPv6. Complete examples are in [example/](example/).

### DNS providers

| Provider               | `target.type` | `key`           | `token`                             | Example                                    |
| ---------------------- | ------------- | --------------- | ----------------------------------- | ------------------------------------------ |
| Cloudflare             | `cloudflare`  | Zone ID         | API token with DNS edit permissions | [cloudflare.yml](example/cloudflare.yml)   |
| Aliyun DNS             | `alidns`      | AccessKey ID    | AccessKey Secret                    | [alidns.yml](example/alidns.yml)           |
| Name.com               | `namedotcom`  | Username        | API token                           | [namedotcom.yml](example/namedotcom.yml)   |
| Tencent Cloud DNSPod   | `dnspod`      | SecretId        | SecretKey                           | [dnspod.yml](example/dnspod.yml)           |
| AWS Route 53           | `route53`     | Access key ID   | Secret access key                   | [route53.yml](example/route53.yml)         |
| Huawei Cloud DNS       | `huaweicloud` | Access key (AK) | Secret key (SK)                     | [huaweicloud.yml](example/huaweicloud.yml) |
| GoDaddy                | `godaddy`     | API key         | API secret                          | [godaddy.yml](example/godaddy.yml)         |
| PowerDNS Authoritative | `powerdns`    | Not needed      | API key                             | [powerdns.yml](example/powerdns.yml)       |
| Demo, no DNS updates   | `sleep`       | Not needed      | Not needed                          | -                                          |

To access a DNS provider through SOCKS5, set `target.proxy`, for example `127.0.0.1:1080`. This setting applies to DNS updates, not address discovery.

### Address sources

**Automatic mode is the default**: omit `source.type` or set it to `auto` to discover IPv4 addresses using multiple services.

| Use case                | `source.type`                                                                          |
| ----------------------- | -------------------------------------------------------------------------------------- |
| Automatic IPv4          | `auto` or `auto-ipv4`                                                                  |
| Automatic IPv6          | `auto-ipv6`; see [ipv6.yml](example/ipv6.yml)                                          |
| Choose servers yourself | For example, `ipify` or `ipify,ipsb`; see [multi-source.yml](example/multi-source.yml) |
| Use a fixed address     | `static`, with the address in `source.content`; see [static.yml](example/static.yml)   |
| Offline demo            | `lo`                                                                                   |

When you specify servers, **only those servers are queried, even on failure**. Successful responses from a list must agree. Do not mix automatic modes with named servers, or IPv4 and IPv6 sources in one job.

Automatic mode uses 3 concurrent sources by default. Set `SIMPLYDDNS_AUTO_SOURCE_COUNT` to adjust it:

```sh
SIMPLYDDNS_AUTO_SOURCE_COUNT=5 ./simplyddns --config simplyddns.yml
```

| Concurrent sources | Matching responses required |
| ------------------ | --------------------------- |
| 3                  | 2                           |
| 5                  | 3                           |
| 7                  | 4                           |

Only these values are supported. Unset means 3; an empty or invalid value is an error. The setting affects automatic mode only. If too few results agree, DNS is left unchanged and checked again later.

<details>
<summary>All available server names</summary>

These services need no API key:

- **IPv4**: `ipify`, `ipsb`, `opendns`, `icanhazip-ipv4`, `ident-ipv4`, `myexternalip-ipv4`, `tnedi-ipv4`, `wtfismyip-ipv4`.
- **IPv6**: `icanhazip-ipv6`, `ident-ipv6`, `ipify-ipv6`, `ipinfo-ipv6`, `ipsb-ipv6`, `opendns-ipv6`, `tnedi-ipv6`, `wtfismyip-ipv6`.
- **Address family depends on the connection**:
  `cloudflare`, `countryis`, `curlmyip`, `cz88`, `eth0`, `freeipapi`, `httpbin`, `httpbingo`, `icanhazip`, `ident`,
  `ifconfigco`, `ifconfigio`, `ifconfigme`, `ipaddress`, `ipconfigio`, `ipecho`, `ipguide`, `ipify-dual`, `ipinfo`, `ipip`,
  `iplocation`, `ipplus360`, `ipquery`, `ipsb-dual`, `ipwhois`, `myexternalip`, `myip`, `tencent`, `tnedi`, `tyk`, `ustc`, `wtfismyip`, `zxinc`.

Choose entries from the IPv4 or IPv6 list when you need a specific address family.

</details>

> Proxies may cause sources to report a proxy exit IP, even when multiple results agree. Check your direct-routing or proxy bypass rules before using the result for DDNS. IPv6 discovery needs working IPv6 connectivity, either directly or through the proxy.

### Common settings

| Setting           | Usage                                                                                                          |
| ----------------- | -------------------------------------------------------------------------------------------------------------- |
| `source.interval` | Check interval in seconds. Default: 60; values below 30 use 30. For regular use, 300 or longer is recommended. |
| `logfile`         | Defaults to stderr. Use `/dev/stdout` or a writable file path if preferred.                                    |
| `debug`           | Set to `true` for more detailed troubleshooting logs.                                                          |

Use `--config` to select a file. Without it, SimplyDDNS searches `/etc/simplyddns`, `$HOME/.config/simplyddns`, then the current directory for `simplyddns.yml` or `simplyddns.yaml`.

### Notifications

To receive a notification after a successful DNS update, add `webhook` to the job at the same level as `source` and `target`:

```yaml
webhook:
  url: "https://example.com/ddns-hook"
  token: "<optional-bearer-token>"
```

Notifications include the new address, domains and update time. You can also configure `username` and `password` for authentication. Use an HTTPS endpoint.

## Docker

With Docker Engine and Compose available, change the configuration mount in [compose.yml](compose.yml) to your own file, then run from the repository:

```sh
mise exec docker-cli -- docker compose up --build -d
```

The default mount is the offline demo and does not update DNS. Compose also supports `SIMPLYDDNS_AUTO_SOURCE_COUNT`.

## Source and target development

Each online source has its own implementation and matching test file in [source/](source/), including separate IPv4/IPv6 variants. File names replace `-` with `_`: for example, `ipify-ipv6` lives in [ipify_ipv6.go](source/ipify_ipv6.go) and [ipify_ipv6_test.go](source/ipify_ipv6_test.go). Keep the endpoint, headers, parsing options and response fixtures in that source's files; only common HTTP/validation helpers and test assertions are shared.

Each target type follows the same rule in [target/](target/), for example [dnspod.go](target/dnspod.go) and [dnspod_test.go](target/dnspod_test.go). Target tests use mocked HTTP APIs and never modify real DNS.

Run individual source or target tests without contacting the API:

```sh
mise exec go -- go test ./source -run '^TestSourceIPifyIPv6$' -count=1
mise exec go -- go test ./target -run '^TestDNSPod' -count=1
```

For an opt-in connectivity/response-format check, set `SIMPLYDDNS_LIVE_SOURCES=ipify-ipv6` and run `TestOnlineHTTPSources` instead. Live checks do not compare the response with the machine's real public IP.

## License

This project is licensed under the [MIT](LICENSE) license.
