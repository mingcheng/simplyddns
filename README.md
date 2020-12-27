# Simply DDNS

根据服务器指定的 IP 地址，动态更新对应的 DNS 记录。

## 思路

https://github.com/mingcheng/ddns-go

其实我看到已经有对应的类似的项目，已经是非常的好用了。但是，有几点不符合我个人的期望以及要求：

1. 我不需要个 UI 界面，对应的我其实更想要的是个全功能的配置文件；
2. 需要更多的灵活性同时可以自由的搭配和运行；
3. 作为服务端需要更加详细的日志等信息。

于是就有了这个应用，相对而言可能有比较高的配置门槛，但是有更多的灵活性。

### 概念

[将 DDNS 动态注册 DNS](https://en.wikipedia.org/wiki/Dynamic_DNS) 的行为抽象为两种：一种为如何获取 IP 地址、另外一种为如何配置 DNS，简单的将如下图所示：

因此，主要实现了这个两个接口，并且对应上配置就可以灵活的搭配以及运行，详细参见扩展章节。

## 安装

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

## 扩展

扩展 SimplyDDNS 其实非常的容易，可以分别参考 source 以及 target 中对应的文件即可，比较简单的是 `lo.go` 以及 target 目录中的 `sleep.go` 文件，因为它们并不做任何实际的事情。

### 已实现模块

#### source / lo

#### source / static

#### source / myip

#### target / sleep

#### starget / namedotcom

## FAQ

//@TODO

`- eof -`
