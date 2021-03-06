<div align=center>
    <img src="https://gfer.me/cover.png" width="150"/>
</div>

`GF(GoFrame)` is a modular, lightweight, loosely coupled, high performance application development framework written in Go. Supporting graceful server, hot updates, multi-domain, multi-port, multi-service, HTTP/HTTPS, dynamic/hook routing and many more features. Providing a series of core components and dozens of practical modules.

# Installation
```
go get -u gitee.com/johng/gf
```
# Limitation
```
golang version >= 1.9.2
```

# Documentation

* [中文文档](https://gfer.me/)

# Architecture
<div align=center>
<img src="https://gfer.me/images/arch.png"/>
</div>

# Quick Start

```go
package main

import (
    "gitee.com/johng/gf/g"
    "gitee.com/johng/gf/g/net/ghttp"
)

func main() {
    s := g.Server()
    s.BindHandler("/", func(r *ghttp.Request) {
        r.Response.Write("Hello World")
    })
    s.Run()
}
```

# License

GF is licensed under the [MIT License](LICENSE), 100% free and open-source.
