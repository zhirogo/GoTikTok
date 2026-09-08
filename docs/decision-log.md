# GoTikTok 决策台账（Decision Log）

> 本文档为 GoTikTok 的**决策台账**：全部已确认技术决策的总表（含依据说明）。
> 重大决策的详细比选过程按 ADR 形式记录于 `docs/adr/`（首篇待迭代 1 编写）；本表只记结论。

---

## 一、决策基线（全部经确认）

| 领域 | 决策 |
|---|---|
| 项目定位 | 仿照 TikTok 的视频分享平台（实战项目），**小范围使用（几百到上千人）** |
| 合规 | 匿名昵称即可；小范围试运行形态，无公开商用合规要求（无实名认证/无内容审核/无等保要求）【2026-09-08 对外口径修订】 |
| 团队 | 5-20 人；GitHub 个人账号 + Actions + Projects 协作 |
| 架构 | **模块化单体**（ADR 记录演进路径）；**Monorepo 单仓库** |
| 前端 | React + Vite + TypeScript |
| 后端 | Go + Gin + GORM |
| 数据 | MySQL（Docker Compose 运行）；Redis V1 前期不引入，全部功能完成后视需要评估 |
| 存储 | 阿里云 OSS（视频原文件直传，不转码） |
| 部署 | Docker Compose；**staging = 开发机**；生产服务器 M3 前再定；域名 = IP:端口 |
| V1 功能 | **视频上传/播放、点赞、评论**（关注 = 排除项，V2 再加） |
| 登录 | 用户名 + 密码；**双 token 认证**（JWT access 2h【默认值】+ refresh 7 天逐次轮换、SHA-256 哈希落库）；会话**固定 7 天窗口不滑动**【2026-08-29 确认】 |
| 管理后台 | V1 不需要 |
| 日志 | zap 结构化 JSON + docker logs 查看 + json-file 轮转（10m / 5 个） |
| 监控 | **V1 不上**（风险接受记录见下方「事实说明」；2026-09-08 修订失效指针） |
| 镜像仓库 | M0-M2 不设（staging=开发机，本地构建）；**M3 上线前引入阿里云 ACR** |
| 密钥管理 | GitHub Actions Secrets + 服务器 .env 文件（不入库） |
| API 契约 | swag 注解生成 **Swagger 2.0** 文档（Gin 生态默认）+ Swagger UI；人读版设计文档为 `docs/api-contract.md` |
| 测试 | 后端 `go test` + testcontainers-go；前端 Vitest + Testing Library |
| 分支策略 | GitHub Flow（main + feature 分支 + PR） |
| Commit 规范 | Conventional Commits |
| 代码规范 | golangci-lint + gofmt ｜ ESLint + Prettier ｜ **pre-commit 提交钩子（2026-08-09 接入）** |
| 版本管理 | 语义化版本 + git tag + CHANGELOG 自动生成 |

**事实说明**：
- swag 默认生成 Swagger 2.0 格式（Gin 生态事实标准），小范围使用够用；OpenAPI 3.0 需要额外工具链，未列入 V1
- **监控不上的风险接受**【2026-09-08 记录，原引用"决策基线第八节"在台账合并后失效】：目标规模小（几百到上千人），V1 以 zap 结构化日志 + docker logs（json-file 轮转）作为排障手段；指标（QPS/延迟/错误率）与告警作为 V1 上线后的评估项

---
