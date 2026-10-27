# Simply DDNS - Yet Another Universal DDNS Tools

- [概念](#概念)
- [安装](#安装)
  - [使用 mise 构建](#使用-mise-构建)
  - [docker-compose](#docker-compose)
- [配置](#配置)
- [扩展](#扩展)
  - [source](#source)
  - [target](#target)
- [开发与发布](#开发与发布)
- [FAQ](#faq)
  - [WebHook 的配置](#webhook-的配置)

这个小工具简单的说，就是根据获得的 IP 地址，监控和更新对应的域名的记录。[其实我看到已经有对应的类似的项目](https://github.com/mingcheng/ddns-go)，已经是非常的好用。但是，有几点不符合我个人的期望以及要求：

1. 我不需要个 UI 界面，对应的我其实更想要的是个全功能的配置文件；
2. 需要更多的灵活性，而且可以自由的搭配配置；
3. 作为服务端需要更加详细的日志，并能回传需要的信息。

加上这块其实是强需求（大家都知道建站的时候，其实最后一部就是更改 DNS，非常的麻烦而且容易出错）。
于是就有了这个应用，可能相对而言可能有比较高的配置门槛，但是有更多的灵活性。

## 概念

[将 DDNS 动态注册 DNS](https://en.wikipedia.org/wiki/Dynamic_DNS) 的行为抽象为两种：一种为如何获取 IP 地址、另外一种为如何配置 DNS，简单的将如下图所示：

因此，主要实现了这个两个接口，并且对应上配置就可以灵活的搭配以及运行，详细参见扩展章节。

## 安装

发布流程会生成 Linux amd64/arm64、macOS amd64/arm64 和 Windows amd64 的压缩包及 `checksums.txt`。每个包都包含可执行文件、许可证、文档和安全的示例配置。下载后可用 `shasum -a 256 -c checksums.txt` 校验完整性。

### 使用 mise 构建

项目工具链版本统一固定在 [mise.toml](mise.toml)，缺失或升级工具均使用 [mise](https://mise.jdx.dev/)。配置中的 `GOTOOLCHAIN=local` 禁止 Go 自动下载不受 mise 管理的替代工具链。

```sh
mise trust
mise install
mise exec -- make build
./simplyddns --version
./simplyddns --config example/basic.yml
```

最后一条命令运行无需公网访问的 loopback/sleep 示例。安装到 Go 的可执行文件目录可使用 `mise exec -- make install`。

### docker-compose

运行前需要可用的 Docker 引擎及 Compose 插件：

```sh
mise exec -- make build_docker_image
mise exec docker-cli -- docker run --rm simplyddns:2.0.0 --version
mise exec docker-cli -- docker compose up --build -d
```

[compose.yml](compose.yml) 默认只读挂载安全的示例配置，实际更新 DNS 时请改为自己的配置文件。容器以非特权用户运行，建议输出日志到 stderr；使用日志文件时需保证目标路径可写。[Dockerfile](Dockerfile) 内也通过 mise 安装与本地一致的 Go 版本。

## 配置

完整示例见 [example/](example/)。可通过 `--config` 指定配置文件；否则依次在 `/etc/simplyddns`、`$HOME/.config/simplyddns` 和当前目录中查找 `simplyddns.yml` 或 `simplyddns.yaml`。

以下是 [basic.yml](example/basic.yml) 中的最小配置：

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

这个配置文件的其中含义就是从 lo 模块中获取 IP 地址后，传给 sleep 的 target 去执行。 因为 sleep 的 target 除了 sleep
几秒以外，其实什么都不干但由此可以看出 SimplyDDNS 的工作机制。

`interval` 的单位为秒，默认值为 60，小于 30 时按 30 处理，首次检查仍等待一个周期。每轮任务有 30 秒超时限制，内置操作支持停止时取消请求。无效的任务定义、域名及溢出的时间间隔会明确报错。

多个 source 可用逗号分隔，如 `ipify,ipsb`。成功返回的地址必须一致，较大的 source 组还会拒绝过多失败的结果。IPv4 和 IPv6 请分别配置任务；target 根据地址自动选择 A 或 AAAA 记录。成功更新后会缓存地址，地址不变时不重复更新；失败则在后续周期重试。

日志默认输出到 stderr。`/dev/stdout` 和 `/dev/stderr` 直接使用现有标准流，可用于只读容器；普通日志文件会在停止服务时关闭。

## 扩展

扩展通过注册 `SourceFunc` 或 `TargetFunc` 实现，参见 [func.go](func.go)、[source/lo.go](source/lo.go) 和 [target/sleep.go](target/sleep.go)。扩展必须响应传入的 context，并对无效输入和执行失败返回错误。

### source

| 名称         | 地址来源                                    |
| ------------ | ------------------------------------------- |
| `lo`         | 本地回环地址 `127.0.0.1`，用于测试          |
| `static`     | `source.content` 中指定的 IPv4 或 IPv6 地址 |
| `cloudflare` | Cloudflare trace 接口                       |
| `cz88`       | CZ88 JSON 接口                              |
| `ipify`      | ipify IPv4 接口                             |
| `ipinfo`     | IPinfo 纯文本接口                           |
| `ipip`       | IPIP JSON 接口                              |
| `ipplus360`  | IPPlus360 JSON 接口                         |
| `ipsb`       | IP.SB IPv4 接口                             |
| `myip`       | `api.my-ip.io` 纯文本接口                   |
| `opendns`    | OpenDNS IPv4 接口                           |

部分服务只支持一种地址类型，不要在同一任务中混用返回 IPv4 和 IPv6 的 source。

### target

_sleep_

啥都不干，就是 Sleep 暂停几秒，可以作为测试使用。

_alidns_

使用阿里云 DNS，`key` 为 AccessKey ID，`token` 为 AccessKey Secret。配置参见 [example/alidns.yml](example/alidns.yml)。

_namedotcom_

使用 Name.com，`key` 为用户名，`token` 为 API token。配置参见 [example/namedotcom.yml](example/namedotcom.yml)。

_cloudflare_

使用 Cloudflare，`key` 为 Zone ID，`token` 需具有该区域的 DNS 编辑权限。更新采用原地修改，保留记录 ID、TTL、代理开关及其他设置。配置参见 [example/cloudflare.yml](example/cloudflare.yml)。

以上三个 DNS target 均可通过 `proxy` 配置 SOCKS5 代理，格式为 `host:port`。

## 开发与发布

```sh
mise exec -- make check       # 格式、依赖一致性、vet、单元测试、race/覆盖率、CI 与 Dockerfile 检查
mise exec -- make vulncheck   # 可达依赖漏洞检查
mise exec -- make verify-dist
mise exec -- make release     # 完整发布检查；不会创建 tag、推送或发布
```

测试采用本地 HTTP 模拟和 DNS 桩，不读取服务商凭据、不修改真实 DNS。覆盖率写入 `coverage.out`。发布验证会检查包内文件、SHA-256、目标系统/架构、Go 版本、CGO 关闭状态，以及当前平台可执行文件的 `--version` 输出。

打包需要主机提供常规 shell 和归档工具（`tar`、`zip`、`shasum`）；可选的 macOS `darwin_universal` 目标还需要平台自带的 `lipo`。`make clean` 只清理生成的文件，不递归删除整个分发目录。

GitHub CI 覆盖 Linux、macOS、Windows，执行 race 检查并上传候选发布包；安全检查使用 CodeQL 和 mise 固定版本的 govulncheck。容器构建也会先运行测试。

发布 2.0.0 前：

1. 审核 [CHANGELOG.md](CHANGELOG.md)，填写发布日期，并确认 `make version` 输出为 `2.0.0`。
2. 执行 `mise exec -- make release`，并等待所有 CI 任务通过。
3. 提交审核后的变更，再创建和推送 `v2.0.0` 标签。

发布工作流只检出已存在的版本标签，检查标签与 Makefile 版本一致，重新通过发布检查后才上传压缩包及校验和。手动触发同样要求标签已存在，不会把所选分支冒充成另一个版本发布。预发布版本会正确标记，也不会提升为容器的稳定 `latest` 标签。

## FAQ

_如果我只有一种请求 IP 地址的方式但是有多个域名，这样子一来重复配置二来看起来配置非常的臃肿，有没有更好的方案？_

这是个比较常见的场景，你可以考虑使用 Yaml 的引用特性减少重复的配置，例如：

```yaml
defaults: &defaults
  type: lo
  interval: 60
```

首先定义每分钟返回个 lo 的回传本地地址，然后配置对应的域名更新：

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

这样子配置后，只需要关心 target 也就是域名的配置即可。

### WebHook 的配置

Webhook 在地址成功更新后触发，需要配置在对应任务内：

```yaml
ddns:
  - source:
      type: lo
      interval: 60
    target:
      type: sleep
    webhook:
      url: https://example.com/webhook
      token: "<可选的 Bearer token>"
```

类似这样子配置后，当地址发生改变时，会将结果和更改的域名结果 POST 到对应的地址。结果的格式如下：

```json
{
    "address": "127.0.0.1",
    "domains": "a.com,b.com",
    "now": "2022-07-13T15:00:10.931809+08:00"
}
```

如果有多个域名的结果，会用逗号分隔。

请求还会发送 `Address` 和 `Domains` 头；也可使用 `username`、`password` 配置 Basic 认证，携带凭据时应使用 HTTPS。任意 2xx 响应都视为成功；Webhook 失败会记录日志，但不会回滚或重复已成功的 DNS 更新。

`- eof -`
