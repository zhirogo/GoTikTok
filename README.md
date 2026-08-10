# GoTikTok

## 技术栈

| 领域 | 选型 |
|---|---|
| 架构 | 模块化单体（Monorepo 单仓库） |
| 前端 | React + Vite + TypeScript（`web/`） |
| 后端 | Go + Gin + GORM（`server/`） |
| 数据 | MySQL + Redis（Docker Compose 运行） |
| 存储 | 阿里云 OSS，视频原文件直传、不转码 |
| API 契约 | swag 注释生成 Swagger 2.0 + Swagger UI |
| 日志 | zap 结构化 JSON + docker logs + json-file 轮转 |
| 部署 | Docker Compose |

> 详细决策依据见 `docs/决策基线.md`。
