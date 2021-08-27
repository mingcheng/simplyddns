module github.com/mingcheng/simplyddns

go 1.16

require (
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.1147
	github.com/bsm/redislock v0.7.1 // indirect
	github.com/go-redis/redis/v8 v8.10.0 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jpillora/go-tld v1.1.1
	github.com/judwhite/go-svc v1.2.1
	github.com/namedotcom/go v0.0.0-20180403034216-08470befbe04
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.8.0
	github.com/stretchr/testify v1.7.0
	github.com/valyala/fastjson v1.6.3
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
	repo.wooramel.cn/mingcheng/srk-notification v0.0.0-20210827040309-dbc75ee87b87
	repo.wooramel.cn/mission/tools v0.0.0-20210406094928-04a6f4977b8b // indirect
)

replace github.com/namedotcom/go v0.0.0-20180403034216-08470befbe04 => github.com/mingcheng/namedotcom v0.0.0-20201225012315-7aca9c240303

//replace repo.wooramel.cn/mingcheng/srk-notification v0.0.0-20210521005617-7dc574f34041 => ../srk-notification
