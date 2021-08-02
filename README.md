# Simply DDNS

[![Build Status](https://ci.wooramel.cn/api/badges/mingcheng/simplyddns/status.svg)](https://ci.wooramel.cn/mingcheng/simplyddns)

根据指定的 IP 地址，更新对应的 DNS 记录。

## 思路

[其实我看到已经有对应的类似的项目](https://github.com/mingcheng/ddns-go)，已经是非常的好用。但是，有几点不符合我个人的期望以及要求：

1. 我不需要个 UI 界面，对应的我其实更想要的是个全功能的配置文件；
2. 需要更多的灵活性，而且可以自由的搭配配置；
3. 作为服务端需要更加详细的日志，并能回传需要的信息。

加上这块其实是强需求（大家都知道建站的时候，其实最后一部就是更改 DNS，非常的麻烦而且容易出错）。

于是就有了这个应用，可能相对而言可能有比较高的配置门槛，但是有更多的灵活性。

### 概念

[将 DDNS 动态注册 DNS](https://en.wikipedia.org/wiki/Dynamic_DNS) 的行为抽象为两种：一种为如何获取 IP 地址、另外一种为如何配置 DNS，简单的将如下图所示：

因此，主要实现了这个两个接口，并且对应上配置就可以灵活的搭配以及运行，详细参见扩展章节。

## 安装

出于安全方面的考虑暂时不提供二进制可执行文件，如果您需要自己编译生成二进制文件，可以直接参考 Makefile 文件中的配置。

通常而言 `make build` 是个非常不错的选择，默认情况下运行它以后（当然，需要 Golang 开发环境）就可以在当前目录下生成对应的可执行文件。

### docker-compose

推荐您使用容器的方式运行本应用，在本程序中还提供了 docker-compose.yml 以及 Dockerfile 文件，方便构建和运行镜像。

// @TODO

## 配置

详细的配置文件例子在 example 目录中，具体可以参考详细的 Yaml 文件。下面举个非常简单的例子（来自 basic.yml）文件：

```yaml
logfile: "/dev/stderr"
debug: Yes

ddns:
  - source:
      type: "lo"
      interval: 60 # 1 minute
    target:
      type: "sleep"
```

这个配置文件的其中含义就是从 lo 模块中获取 IP 地址后，传给 sleep 的 target 去执行。

因为 sleep 的 target 除了 sleep 几秒以外，其实什么都不干但由此可以看出 SimplyDDNS 的工作机制。

### Webhook

## 扩展

扩展 SimplyDDNS 其实非常的容易，可以分别参考 source 以及 target 中对应的文件即可，比较简单的是 `lo.go` 以及 target 目录中的 `sleep.go` 文件，因为它们并不做任何实际的事情。

### 已实现模块

#### source

_lo_

这个模块只返回 Lookup 地址，用于测试使用

_static_

指定对应的静态地址，通常用于固定 IP 地址的情况

_myip_

通过访问 ipip.net 得知访问的 IP 地址，推荐的获取 IP 地址方法之一，用于访问主机有公网地址并能直接访问的情况（如 DMZ 主机）。

_biturl_

另外个能够通过访问对方 Web 地址的 IP 获取服务，原有的 ipsb 服务已经废弃不再使用。

#### target

_sleep_

啥都不干，就是 Sleep 暂停几秒，可以作为测试使用。

_alidns_

使用阿里云 DNS 服务的接口，需要后台先申请 API 的 Key 和 Token。详细使用方式在 `target/alidns_test.go` 文件中，配置文件例子在 `example/alidns.yml` 中。

_namedotcom_

使用 Name.com 作为域名服务的后端。因为 Name.com 有白名单验证，所以如果不是本机访问可以考虑使用代理的方式请求（具体请参见 `target/namedotcom_test.go` 这个文件）。

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

WebHook 可以在任何配置更改的情况下通知到用户，可以参考使用以下的配置：

```yaml
webhook:
  method: GET
  url: https://<your-webhook-address>
  timeout: 30
```

默认情况下 Method 为 GET，Timout 如果没有配置的话就是默认值。

`- eof -`
