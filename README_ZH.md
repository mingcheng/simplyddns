# SimplyDDNS

[English](README.md)

让域名自动跟随公网 IP 变化，方便通过固定域名访问家庭服务器、NAS 或其他自建服务。

- **配置简单**：一个 YAML 文件即可管理，无需 Web 界面。
- **来源灵活**：默认自动从多个服务核对地址，也可限定为自己指定的服务器。
- **适用多种场景**：支持 IPv4、IPv6、多域名，以及 Cloudflare、阿里云 DNS、Name.com。
- **按需更新**：避免不必要的 DNS 修改，提供日志和可选的 Webhook 通知。

## 快速开始

### 1. 安装

可通过 [Homebrew](https://brew.sh/) 直接安装：

```sh
brew install mingcheng/tap/simplyddns
```

Homebrew 用户请将下文命令中的 `./simplyddns` 替换为 `simplyddns`。

也可从 [Releases](https://github.com/mingcheng/simplyddns/releases) 下载适用于 Linux、macOS 或 Windows 的压缩包并解压。以下命令以 Linux/macOS 为例，Windows 请运行 `simplyddns.exe`。

<details>
<summary>使用 mise 从源码构建</summary>

安装 [mise](https://mise.jdx.dev/) 后执行：

```sh
git clone https://github.com/mingcheng/simplyddns.git
cd simplyddns
mise trust
mise install go make
mise exec go make -- make build
```

</details>

### 2. 配置

创建 `simplyddns.yml`，或替换发行包自带的演示配置。下面以 Cloudflare 更新两个域名为例，请先换成自己的凭据和域名：

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

Cloudflare 的 API token 需要拥有对应区域的 DNS 编辑权限。其他服务商的配置见[域名服务商](#域名服务商)。

### 3. 运行

```sh
./simplyddns --version
./simplyddns --config simplyddns.yml
```

保持进程运行即可持续检查和更新记录，首次检查会等待一个配置周期。

仅体验、不修改 DNS 时，可使用[离线示例](example/basic.yml)或[自动获取地址示例](example/auto.yml)。发行包自带的默认配置也是离线示例。

## 配置指南

`ddns` 下的每一项是一项任务：`source` 选择地址来源，`target` 选择要更新的 DNS 服务。同一服务商和凭据下可添加多个域名；不同账号、服务商或 IPv4/IPv6 请分别配置任务。完整示例见 [example/](example/)。

### 域名服务商

| 服务商                 | `target.type` | `key`            | `token`                       | 示例                                       |
| ---------------------- | ------------- | ---------------- | ----------------------------- | ------------------------------------------ |
| Cloudflare             | `cloudflare`  | Zone ID          | 具有 DNS 编辑权限的 API token | [cloudflare.yml](example/cloudflare.yml)   |
| 阿里云 DNS             | `alidns`      | AccessKey ID     | AccessKey Secret              | [alidns.yml](example/alidns.yml)           |
| Name.com               | `namedotcom`  | 用户名           | API token                     | [namedotcom.yml](example/namedotcom.yml)   |
| 腾讯云 DNSPod          | `dnspod`      | SecretId         | SecretKey                     | [dnspod.yml](example/dnspod.yml)           |
| AWS Route 53           | `route53`     | Access key ID    | Secret access key             | [route53.yml](example/route53.yml)         |
| 华为云 DNS             | `huaweicloud` | Access key（AK） | Secret key（SK）              | [huaweicloud.yml](example/huaweicloud.yml) |
| GoDaddy                | `godaddy`     | API key          | API secret                    | [godaddy.yml](example/godaddy.yml)         |
| PowerDNS Authoritative | `powerdns`    | 无需填写         | API key                       | [powerdns.yml](example/powerdns.yml)       |
| 体验模式，不更新 DNS   | `sleep`       | 无需填写         | 无需填写                      | -                                          |

访问 DNS 服务商需要 SOCKS5 代理时，可设置 `target.proxy`，例如 `127.0.0.1:1080`。该设置仅用于 DNS 更新，不控制地址来源请求。

### 地址来源

**默认使用自动模式**：省略 `source.type` 或填写 `auto`，即可通过多个服务核对 IPv4 地址。

| 使用方式       | `source.type`                                                               |
| -------------- | --------------------------------------------------------------------------- |
| 自动获取 IPv4  | `auto` 或 `auto-ipv4`                                                       |
| 自动获取 IPv6  | `auto-ipv6`，见 [ipv6.yml](example/ipv6.yml)                                |
| 自行指定服务器 | 如 `ipify` 或 `ipify,ipsb`，见 [multi-source.yml](example/multi-source.yml) |
| 使用固定地址   | `static`，地址填写在 `source.content`，见 [static.yml](example/static.yml)  |
| 离线体验       | `lo`                                                                        |

指定服务器后，**只请求指定的服务器，失败也不会改用其他来源**。多个来源成功返回的地址必须一致。自动模式不能与指定服务器混用，也不要在同一任务中混合 IPv4 和 IPv6 来源。

自动模式默认并发请求 3 个来源，可通过环境变量 `SIMPLYDDNS_AUTO_SOURCE_COUNT` 调整：

```sh
SIMPLYDDNS_AUTO_SOURCE_COUNT=5 ./simplyddns --config simplyddns.yml
```

| 并发来源数 | 至少需要一致的结果数 |
| ---------- | -------------------- |
| 3          | 2                    |
| 5          | 3                    |
| 7          | 4                    |

仅支持表中的取值；未设置时使用 3，空值或无效值会报错。此配置只影响自动模式。相同结果不足时不修改 DNS，稍后再次检查。

<details>
<summary>查看全部可指定的服务器名称</summary>

这些来源均无需 API key：

- **IPv4**：`ipify`、`ipsb`、`opendns`、`icanhazip-ipv4`、`ident-ipv4`、`myexternalip-ipv4`、`tnedi-ipv4`、`wtfismyip-ipv4`。
- **IPv6**：`icanhazip-ipv6`、`ident-ipv6`、`ipify-ipv6`、`ipinfo-ipv6`、`ipsb-ipv6`、`opendns-ipv6`、`tnedi-ipv6`、`wtfismyip-ipv6`。
- **地址类型由实际网络连接决定**：
  `cloudflare`、`countryis`、`curlmyip`、`cz88`、`eth0`、`freeipapi`、`httpbin`、`httpbingo`、`icanhazip`、`ident`、
  `ifconfigco`、`ifconfigio`、`ifconfigme`、`ipaddress`、`ipconfigio`、`ipecho`、`ipguide`、`ipify-dual`、`ipinfo`、`ipip`、
  `iplocation`、`ipplus360`、`ipquery`、`ipsb-dual`、`ipwhois`、`myexternalip`、`myip`、`tencent`、`tnedi`、`tyk`、`ustc`、`wtfismyip`、`zxinc`。

需要固定地址类型时，请从对应的 IPv4 或 IPv6 列表中选择。

</details>

> 代理可能让来源返回代理出口 IP，即使多个结果一致也不能保证是本机真实公网 IP。用于 DDNS 前请检查直连或代理绕过规则。IPv6 来源需要可用的 IPv6 网络，或具备 IPv6 出口的代理。

### 常用选项

| 配置项            | 用法                                                                               |
| ----------------- | ---------------------------------------------------------------------------------- |
| `source.interval` | 检查间隔，单位为秒。默认 60，小于 30 时按 30 处理；日常使用建议设置为 300 或更长。 |
| `logfile`         | 默认输出到 stderr，也可填写 `/dev/stdout` 或可写的日志文件路径。                   |
| `debug`           | 设置为 `true` 可查看更详细的排查日志。                                             |

通过 `--config` 指定配置文件；未指定时，依次在 `/etc/simplyddns`、`$HOME/.config/simplyddns` 和当前目录查找 `simplyddns.yml` 或 `simplyddns.yaml`。

### 更新通知

需要在 DNS 更新成功后接收通知时，在对应任务中添加 `webhook`，与 `source`、`target` 同级：

```yaml
webhook:
  url: "https://example.com/ddns-hook"
  token: "<optional-bearer-token>"
```

通知包含新地址、域名和更新时间。也可使用 `username`、`password` 配置认证，建议使用 HTTPS 地址。

## Docker 部署

准备好 Docker 引擎及 Compose 后，将 [compose.yml](compose.yml) 中的配置挂载路径改为自己的文件，再在仓库目录执行：

```sh
mise exec docker-cli -- docker compose up --build -d
```

默认挂载的是离线示例，不会更新 DNS。Compose 同样支持 `SIMPLYDDNS_AUTO_SOURCE_COUNT` 环境变量。

## 来源与目标开发

[source/](source/) 中每个线上来源独立一个实现文件，并配有同名测试文件，IPv4/IPv6 变体也分别存放。文件名将 `-` 替换为 `_`，例如 `ipify-ipv6` 对应 [ipify_ipv6.go](source/ipify_ipv6.go) 和 [ipify_ipv6_test.go](source/ipify_ipv6_test.go)。接口地址、请求头、解析参数和响应样本在各自文件中维护，仅复用通用 HTTP/校验函数和测试断言。

[target/](target/) 中也按一种类型一份实现和测试组织，例如 [dnspod.go](target/dnspod.go) 与 [dnspod_test.go](target/dnspod_test.go)。target 测试使用模拟 HTTP API，不会修改真实 DNS。

单独运行来源或目标类型的离线测试，不会访问线上 API：

```sh
mise exec go -- go test ./source -run '^TestSourceIPifyIPv6$' -count=1
mise exec go -- go test ./target -run '^TestDNSPod' -count=1
```

需要显式验证连通性及响应格式时，可设置 `SIMPLYDDNS_LIVE_SOURCES=ipify-ipv6`，改为运行 `TestOnlineHTTPSources`；线上检查不会与本机真实公网 IP 比较。

## 许可证

本项目采用 [MIT](LICENSE) 许可证。
