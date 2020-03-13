module github.com/guangbochen/device-temp-demo

go 1.13

replace (
	github.com/matryer/moq => github.com/rancher/moq v0.0.0-20190404221404-ee5226d43009
	github.com/rancher/wrangler-api => github.com/guangbochen/wrangler-api v0.4.2-0.20200109041411-75d8ec3a7364
)

require (
	github.com/bettercap/gatt v0.0.0-20191018133023-569d3d9372bb
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/mgutz/logxi v0.0.0-20161027140823-aebf8a7d67ab // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0 // indirect
	github.com/urfave/cli v1.22.2
	github.com/yosssi/gmq v0.0.1
	golang.org/x/sys v0.0.0-20191010194322-b09406accb47 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.2.4 // indirect
	k8s.io/klog v1.0.0
)
