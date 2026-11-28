# zGia! Book Library

Go 编写的图书/小说管理 RESTful API 服务：管理图书、作者、分类、卷、章节与正文，支持搜索、导出和 JWT 鉴权，配合 [book-frontend](https://github.com/zgia/book-frontend.git) 使用。

## 特性

- 图书、作者、分类、卷、章节、正文的增删改查
- 关键词搜索（书名、作者等条件），分页返回
- 正文 2-gram 全文索引（SQLite FTS5）：未建索引时自动回退 LIKE，检索不中断
- JWT 登录鉴权（token 有效期 365 天，支持刷新）
- 导出书籍为纯文本文件
- 中英文国际化（i18n）
- CORS、GZIP 压缩、pprof 性能分析（可配置开关）
- 基于 Zap 的结构化日志，日志文件自动轮转
- 数据库支持 MySQL 与 SQLite
- 支持 `--daemon` 直接后台运行，无需 `nohup`

## 技术栈

| 组件 | 说明 |
| --- | --- |
| [Gin](https://github.com/gin-gonic/gin/) | HTTP 框架 |
| [XORM](https://gitea.com/xorm/xorm) | ORM，支持 MySQL / SQLite |
| [Zap](https://github.com/uber-go/zap) | 结构化日志，配合 lumberjack 轮转 |
| [gin-jwt](https://github.com/appleboy/gin-jwt) | JWT 鉴权中间件 |
| [urfave/cli/v2](https://github.com/urfave/cli) | 命令行框架 |
| [golang.org/x/text](https://pkg.go.dev/golang.org/x/text) | 国际化（内置 translation 模块） |

## 目录结构

```
book.go                 入口，注册 CLI 命令与全局参数
conf/                   默认配置（app.ini，编译时内嵌）与 i18n 语言文件
book.sql                MySQL 建库脚本
internal/
  cmd/                  CLI 子命令：api、export、extra、fts
  conf/                 配置加载与解析（内嵌默认配置 + custom 覆盖）
  daemon/               进程自后台化（--daemon 实现）
  db/                   数据库访问层（XORM），fts.go 为 SQLite 全文索引
  initial/              应用初始化
  logger/               Zap 日志
  middleware/           JWT 中间件
  models/               业务模型与数据访问
  modules/              附加模块（translation/i18n）
  route/                HTTP 路由与 Gin 配置
  util/                 工具函数
custom/conf/app.ini     自定义配置（可选，覆盖默认配置）
```

## 快速开始

### 依赖

- Go 1.20+
- 数据库：MySQL 5.7+/8.0，或 SQLite（无需额外服务）

### 初始化数据库

MySQL：

```bash
mysql -u <user> -p <database> < book.sql
```

`book.sql` 会创建全部表与索引，可重复执行。SQLite 模式同样受代码支持（WAL、FTS5 全文索引），仓库不附带单独的建库脚本，可参照 `book.sql` 改写为 SQLite 语法后执行。

### 编译

```bash
go build -tags sqlite_fts5 -o book .
```

`sqlite_fts5` 只在 SQLite 模式下需要：`mattn/go-sqlite3` 默认不编译 FTS5，缺了它正文索引不可用（搜索自动回退 LIKE，功能不受影响）。`.air.toml` 与 CI 配置已带上该标签。

### 运行

```bash
./book api
```

默认监听 `0.0.0.0:6767`，看到 `available_server => http://localhost:6767/` 即启动成功。

开发时可直接 `go run`（注意配置查找问题，见下文）：

```bash
go run book.go api
```

或使用 [air](https://github.com/cosmtrek/air) 热加载（已提供 `.air.toml`，自动以 `api` 参数构建运行）：

```bash
go install github.com/cosmtrek/air@latest
air
```

### 配置查找规则

程序内嵌了一份默认配置 `conf/app.ini`（通过 `go:embed` 编译进二进制），启动时会再叠加一份自定义配置（存在才加载）：

1. `-c` / `--config` 指定的文件，例如 `./book -c /path/to/custom/conf/app.ini api`；
2. 否则为 `<工作目录>/custom/conf/app.ini`。

**工作目录**按以下顺序确定：

1. 环境变量 `BOOK_WORK_DIR`；
2. 未设置时为二进制所在目录。

因此 `go run` 时找不到配置属于正常现象，可任选一种方式：

```bash
# 环境变量
BOOK_WORK_DIR=$(pwd) go run book.go api

# 或 -c 指定配置
go run book.go -c custom/conf/app.ini api
```

## 后台运行（--daemon）

不需要 `nohup`，启动后进程会自动进入后台：

```bash
./book api --daemon     # 或 ./book api -d
```

父进程立即打印 PID 后退出，终端直接回到提示符：

```
book 已在后台启动，PID: 12345，输出写入 nohup.out
```

实现要点：

- 子进程通过 `setsid` 成为新会话的首进程，脱离控制终端，关闭终端不会再收到 `SIGHUP`；
- 标准输入重定向到 `/dev/null`，标准输出与标准错误追加写入**当前目录**的 `nohup.out`；
- 子进程通过环境变量 `BOOK_DAEMONIZED=1` 识别，避免重复后台化；
- 后台化发生在 CLI 初始化之前，父进程不会连接数据库、不会占用端口。

查看与停止：

```bash
pgrep -f 'book api'            # 查看后台进程
kill -15 <PID>                 # 优雅退出（程序收到 SIGTERM 会关闭 HTTP 服务）
```

注意：

- `--daemon` 只支持 `api` 子命令，且要写在 `api` 之后；`book --daemon api` 不支持；
- 输出文件固定为当前工作目录的 `nohup.out`，长期运行请自行清理或改用日志轮转（程序自身日志在 `<工作目录>/log/`）。

### 开机自启（可选，macOS）

如需开机自启和崩溃重启，可交给 launchd 管理，无需 `--daemon`：

```bash
mkdir -p ~/Library/LaunchAgents
cat > ~/Library/LaunchAgents/com.zgia.book.plist <<'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key><string>com.zgia.book</string>
    <key>ProgramArguments</key>
    <array>
        <string>/path/to/book</string>
        <string>api</string>
    </array>
    <key>WorkingDirectory</key><string>/path/to</string>
    <key>RunAtLoad</key><true/>
    <key>KeepAlive</key><dict><key>SuccessfulExit</key><false/></dict>
    <key>StandardOutPath</key><string>/path/to/log/book.out.log</string>
    <key>StandardErrorPath</key><string>/path/to/log/book.err.log</string>
</dict>
</plist>
EOF

launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.zgia.book.plist
launchctl print gui/$(id -u)/com.zgia.book | head -20     # 查看状态
launchctl bootout gui/$(id -u) ~/Library/LaunchAgents/com.zgia.book.plist  # 停止
```

## 命令行

```
book [全局选项] command [命令选项]
```

全局选项：

| 选项 | 说明 |
| --- | --- |
| `-c, --config <path>` | 自定义配置文件路径 |

子命令：

### api

启动 HTTP API 服务。

```bash
./book api              # 前台运行
./book api --daemon     # 后台运行
```

### export

把指定 ID 之后的图书导出为文本文件。

```bash
./book -c /path/to/app.ini export -b 0 -p /tmp
```

| 选项 | 说明 |
| --- | --- |
| `-b, --book_id <id>` | 必填，导出该 ID 之后的图书 |
| `-p, --path <dir>` | 导出目录，默认 `/tmp`，需可写 |

### extra

打印应用信息（版本、路径、配置位置、日志目录、工作目录等），用于排查环境问题。

```bash
./book extra
```

### fts

维护 SQLite 正文全文索引（仅 SQLite 模式）。

```bash
./book fts rebuild      # 全量重建 2-gram 索引
./book fts stat         # 查看索引覆盖情况
```

## 配置项

默认值见 `conf/app.ini`，自定义配置只需覆盖需要修改的项。

| 段落 | 配置项 | 说明 |
| --- | --- | --- |
| 顶层 | `BRAND_NAME` | 应用名称 |
| 顶层 | `RUN_MODE` | `dev` / `prod` / `test`，`prod` 时 Gin 进入 release 模式 |
| `[server]` | `EXTERNAL_URL` | 对外访问地址 |
| `[server]` | `DOMAIN` / `PROTOCOL` | 域名与协议（http/https/fcgi/unix） |
| `[server]` | `HTTP_ADDR` / `HTTP_PORT` | 监听地址与端口，默认 `0.0.0.0:6767` |
| `[server]` | `ENABLE_GZIP` | 是否开启 GZIP |
| `[server]` | `ENABLE_PPROF` | 是否开启 pprof 端点 |
| `[database]` | `TYPE` | `mysql` / `sqlite` / `sqlite3` |
| `[database]` | `HOST` | MySQL 为 `host:port`；SQLite 为数据库文件路径 |
| `[database]` | `NAME` | MySQL 数据库名；SQLite 下为 HOST 为空时的备用路径 |
| `[database]` | `USER` / `PASSWORD` | MySQL 账号密码 |
| `[database]` | `MAX_OPEN_CONNS` / `MAX_IDLE_CONNS` | 连接池大小（SQLite 会被夹到 1~8，默认 4，空闲数跟随） |
| `[http]` | `ACCESS_CONTROL_ALLOW_ORIGIN` | CORS 允许的来源，逗号分隔 |
| `[http]` | `JWT_SECRET_KEY` | JWT 签名密钥，生产环境务必修改 |
| `[api]` | `RESPONSE_ITEMS` / `MAX_RESPONSE_ITEMS` | 分页默认条数与上限 |
| `[log]` | `ROOT_PATH` | 日志根目录，默认 `<工作目录>/log` |
| `[log]` | `MODE` / `LEVEL` | 输出方式（console/file）与级别 |
| `[log.file]` | `LOG_NAME` / `MAX_SIZE` / `MAX_AGE` / `MAX_BACKUPS` | 日志文件名、单文件大小与保留策略 |
| `[i18n]` | `LANGS` / `NAMES` | 启用的语言 |

> 注：日志目录在代码中读取的键名是 `[log] ROOT_PATH`，而默认配置文件里的注释示例写作 `ROOT_DIR`；留空时使用 `<工作目录>/log`。

MySQL 示例：

```ini
[database]
TYPE = mysql
HOST = 127.0.0.1:3306
NAME = library
USER = library
PASSWORD = ******
```

SQLite 示例：

```ini
[database]
TYPE = sqlite
HOST = /path/to/novel.sqlite
```

## 数据库

- 建库脚本：`book.sql`（MySQL，手动执行，代码不会自动同步表结构）；
- 主要表：

| 表 | 说明 |
| --- | --- |
| `author` | 作者，`former_name` 存逗号分隔的曾用名 |
| `book` | 图书，含分类、作者、评分、字数、完本标记 |
| `category` | 分类，支持父子层级与隐藏 |
| `chapter` | 章节，软删除 |
| `content` | 章节正文，`chapter_id` 与 `chapter.id` 一对一 |
| `user` | 用户，密码为 md5 加盐（`salt` 字段） |
| `volume` | 卷，一书可有多卷 |

- 时间字段统一为 `DATETIME`，按 `[time] ZONE`（默认 `Asia/Shanghai`）读写；`updated_at` 由行更新自动刷新，`deleted_at` 为 NULL 表示未删除；
- 列名统一 snake_case：外键 `<表>_id`，布尔 `is_*`，计数 `*_count`。

SQLite 连接固定以 WAL 模式打开（`_journal_mode=wal`、`_busy_timeout=5000`、`_synchronous=normal`、`_foreign_keys=1`），启动时会回读 PRAGMA 并写入日志，便于确认 WAL 是否因文件系统不支持而被降级。首次连接后目录下会出现 `*.sqlite-wal` 与 `*.sqlite-shm`，属正常伴生文件；备份请用 `VACUUM INTO 'backup.sqlite'`，不要在服务运行时直接拷贝主文件。

### 正文全文索引（仅 SQLite）

中文没有词边界，FTS5 内置的 unicode61 会把整句当成一个 token（搜“张三”命中不了“张三丰”）。这里的做法是把正文切成相邻两字的 2-gram 存入 FTS5 虚表，查询词同样切分后用 AND 组合，语义等价于 `LIKE '%关键词%'`，但走倒排索引。

索引表 `content_fts` 使用 `content=''`（contentless，不重复存正文）+ `detail=none`（不存位置），950MB 正文的索引实测约 250–380MB。

```bash
# 全量重建（3 亿字级别的库约 2–3 分钟）
./book fts rebuild

# 查看覆盖情况
./book fts stat
```

说明：

- 重建会先 `DROP` 旧索引表，中途中断不会影响正文，重跑即可
- 章节新增/更新会自动同步索引；软删除的章节保留索引，查询时由 `deleted_at IS NULL` 过滤
- 单字查询、含标点或空格的查询无法用 2-gram 表达，自动回退 LIKE
- 索引表不存在或二进制未编译 FTS5 时，搜索整体回退 LIKE，服务不受影响
- 索引属于加速层，写入索引失败只记 warning，不会导致章节保存失败

## API 概览

### 认证

```bash
# 登录，返回 token
curl -X POST http://localhost:6767/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"******"}'

# 之后请求带上
Authorization: Bearer <token>
```

### 路由分组

| 分组 | 是否需要 Token | 主要接口 |
| --- | --- | --- |
| `GET /` | 否 | 健康检查（返回 code=0） |
| `GET /search` | 否 | 搜索图书，参数 `q` 为关键词 |
| `GET /options` | 否 | 全局选项（分页大小与分类列表） |
| `GET /panic` | 否 | 触发 panic，验证 Recovery 中间件 |
| `/auth` | 部分 | `POST /auth/login`、`GET /auth/refresh_token` 匿名；`GET /auth/user`、`POST /auth/changepassword` 需要 Token |
| `/book` | 是 | 图书列表/详情/新增/修改/删除、下载、完本、章节、卷 |
| `/author` | 是 | `GET /author/index`、`GET /author/:author_id`、`POST /author/:author_id`、`DELETE /author/:author_id` |
| `/cnf` | 是 | `GET /cnf/categories`、`POST /cnf/category/:category_id`、`DELETE /cnf/category/:category_id` |

响应统一为 JSON：

```json
{ "code": 0, "msg": "route.GetBook", "data": { } }
```

`code` 为 0 表示成功；失败时 `code` 为 HTTP 状态码（如 401、404、500）。

## 开发与测试

```bash
gofmt -l .            # 格式检查
go vet ./...          # 静态检查
go test ./...         # 全量测试（连库用例在数据库不可用时自动跳过）
go test -v ./internal/daemon/ -run TestRequested   # 单个测试
```

### 性能分析

先开启 `[server] ENABLE_PPROF = true`，然后：

```bash
go tool pprof --seconds 20 http://localhost:6767/debug/pprof/profile   # CPU
go tool pprof --seconds 20 http://localhost:6767/debug/pprof/heap      # 内存
go tool pprof --seconds 20 http://localhost:6767/debug/pprof/goroutine # 协程

go tool pprof -http=:9966 log/cpu.pprof   # Web 查看（需 brew install graphviz）
```

## 常见问题

**启动报找不到配置**

`go run` 不会复制 `conf/` 目录。设置 `BOOK_WORK_DIR` 或用 `-c` 指定配置文件（见“配置查找规则”）。

**端口被占用**

修改 `[server] HTTP_PORT`，或先停掉旧进程：`pgrep -f 'book api'` 找到 PID 后 `kill -15 <PID>`。

**如何停止服务**

- 前台运行：`CTRL+C`；
- 后台（`--daemon` / nohup）：`kill -15 <PID>`，程序会优雅关闭 HTTP 服务；
- 也兼容 Linux 写法：`ps -ef | grep book | grep -v grep | awk '{print $2}' | xargs kill -15`。

**nohup.out 不断变大**

`--daemon` 只负责把标准输出落到 `nohup.out`，不会轮转。程序自身的结构化日志在 `<工作目录>/log/` 并按大小轮转，可定期清理 `nohup.out`。

**MySQL 报时区错误**

服务端会按 `[time] ZONE` 自动设置会话时区（如 `SET time_zone='+08:00'`）。若数据库账号不允许改会话时区，请在 MySQL 侧统一时区，或把 `[time] ZONE` 调整为与服务器一致。

## 感谢

项目用于 Golang 练手，核心结构参考了 [gogs](https://github.com/gogs/gogs)，感谢原作者。

![gogs-brand](https://user-images.githubusercontent.com/2946214/146899259-6a8b58ad-8d6e-40d2-ab02-79dc6aadabbf.png)
