module github.com/mingcheng/simplyddns

go 1.15

require (
	github.com/hashicorp/serf v0.8.2
	github.com/jpillora/go-tld v1.1.1
	github.com/namedotcom/go v0.0.0-20180403034216-08470befbe04
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.3.0
	github.com/valyala/fastjson v1.6.3
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
)

replace github.com/namedotcom/go v0.0.0-20180403034216-08470befbe04 => github.com/mingcheng/namedotcom v0.0.0-20201225012315-7aca9c240303
