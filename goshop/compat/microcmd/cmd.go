// Package cmd 补上 go-micro v4.8 删除的 go-micro.dev/v4/cmd。
// Consul 注册插件在 init 里会执行 cmd.DefaultRegistries["consul"] = NewRegistry。
// 没有这个包，插件无法编译。业务代码不要往这里加逻辑。
package cmd

import "go-micro.dev/v4/registry"

// DefaultRegistries 的键是注册中心名字，值是创建函数。插件启动时往里面写入。
var DefaultRegistries = map[string]func(...registry.Option) registry.Registry{}
