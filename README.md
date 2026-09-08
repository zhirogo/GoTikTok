# GoTikTok

仿照 TikTok 的视频分享平台（React + Go 实战项目）。V1 功能：视频上传/播放、点赞、评论；关注为 V2 排除项。

## 当前状态

**代码处于工程准备期，V1 功能尚未开始编码。**

- 设计文档已完成并入库（`docs/`）：功能规格、接口契约、数据库设计、决策台账
- 后端仅 `GET /ping` 健康检查（Go + gin + zap、优雅停机）；尚未引入 GORM、未连接 MySQL
- 前端为 Vite 占位页（仅渲染标题），未调用任何接口
- V1 功能范围与迭代计划见 [docs/v1-spec.md](docs/v1-spec.md)

## 快速开始

前置：Go 1.26+、Node 24（与 CI 一致）。**本阶段不需要 Docker / MySQL。**

后端：

```bash
cd server
go run ./cmd/gotiktok          # 默认监听 :8080
# 验证：curl http://localhost:8080/ping → {"message":"pong"}
# 改端口：GOTIKTOK_HTTP_ADDR=:9090 go run ./cmd/gotiktok
```

前端（建议先启动后端，`/api` 代理依赖后端在线）：

```bash
cd web
npm ci
npm run dev                    # http://localhost:5173，/api 代理到后端 :8080
```

> 5173 被占用时 Vite 会自动递增端口，以终端输出为准。

## 技术栈

Go（gin + zap）后端 + React（Vite + TypeScript）前端。MySQL、阿里云 OSS、双 token 认证、部署等选型与依据见 [docs/decision-log.md](docs/decision-log.md)（决策台账），此处不重复。

## 文档

| 文档 | 内容 |
|---|---|
| [docs/v1-spec.md](docs/v1-spec.md) | V1 规格：功能点、边界与验收标准、迭代计划 |
| [docs/api-contract.md](docs/api-contract.md) | 接口契约：13 个接口、错误码表、分页与认证模型 |
| [docs/database-design.md](docs/database-design.md) | 数据库设计：5 张表结构、索引与查询模式 |
| [docs/decision-log.md](docs/decision-log.md) | 决策台账：全部已确认技术决策总表 |

## 目录结构

```
server/   后端（cmd 入口、interface 接口层；application/domain/infrastructure/migrations 为预留目录）
web/      前端（React + Vite + TS）
docs/     设计文档（规格 / 契约 / 数据库 / 决策台账）
.github/workflows/   CI（lint + build + test）
```

## 配置

当前唯一运行时变量：

| 变量 | 说明 | 默认值 |
|---|---|---|
| `GOTIKTOK_HTTP_ADDR` | 后端监听地址 | `:8080` |

数据库连接等凭据变量尚未建立（代码未接入数据库），暂无 `.env.example`——待迭代 0 落地 compose 与 GORM 接入时补充。

## 质量与 CI

- 本地提交钩子：`.pre-commit-config.yaml`（需 `pip install pre-commit && pre-commit install`）
- CI：GitHub Actions 在 push/PR 自动跑 lint + build + test（`.github/workflows/ci.yml`）

| 检查 | 命令 |
|---|---|
| 后端 lint | `cd server && golangci-lint run` |
| 后端测试 | `cd server && go test ./...`（当前无测试文件） |
| 前端 lint / build / test | `cd web && npm run lint` / `npm run build` / `npm run test:run` |
