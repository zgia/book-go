# Library project by golang

## 前端源码
[book-frontend](https://github.com/zgia/book-frontend.git)

## 基础框架

1. [gin](https://github.com/gin-gonic/gin/) 
2. [xorm](https://gitea.com/xorm/xorm)，支持MySQL和SQLite

## go run
编译时，不会复制conf目录，因此找不到app.ini，可以使用env方式解决

> \> BOOK_WORK_DIR=dir/to/work go run book.go

或者在 .bashrc 中加一条：`export BOOK_WORK_DIR=dir/to/work`，命令行时，直接 `go run book.go` 即可

或者使用命令行参数 <code>-c</code> 指定 app.ini 位置，如果是相对路径，则在book所在目录下查找
> \> go run book.go -c custom/conf/app.ini

> \> go run book.go -c /path/to/custom/conf/app.ini

使用 
- <code>CTRL+C</code>

或者
- <code>ps -ef | grep book | grep -v grep | awk '{print $2}' | xargs kill -15</code>

的方式结束进程

## 性能分析
```go
// go get -u github.com/pkg/profile
import "github.com/pkg/profile"

func main() {
	defer profile.Start(profile.MemProfile, profile.MemProfileRate(1), profile.ProfilePath("./log")).Stop()
	// ...
}
```
或者
```go
// go get -u github.com/gin-contrib/pprof
import "github.com/gin-contrib/pprof"

// g = gin.New()
pprof.Register(g)
```
采集协程数据，文件会保存到 $HOME/pprof/ 下
> \> go tool pprof --seconds 20 http://localhost:6767/debug/pprof/goroutine

### Web方式查看结果
使用web分析查看时，需要安装graphviz
> \> brew install graphviz

> \> go tool pprof -http=:9966 log/cpu.pprof


## 组件
使用 go mod 管理组件
> \> go mod init ModuleName

> \> go mod tidy

添加新组件： 
> \> go get -u example.com/component

## 热加载

> \> go install github.com/cosmtrek/air@latest

> \> air

github: [air](https://github.com/cosmtrek/air) 热加载

## 感谢 gogs

学习golang练手用，因此直接照抄了 [gogs](https://github.com/gogs/gogs) 的核心源码，感谢大佬

![gogs-brand](https://user-images.githubusercontent.com/2946214/146899259-6a8b58ad-8d6e-40d2-ab02-79dc6aadabbf.png)