# GoTikTok V1 接口契约（API Contract）

> 本文档为 GoTikTok 的**接口契约**：接口设计的权威依据，swag 注解、后端 handler、前端联调均按此实现。
> 人读版设计以本文档为准；swag 生成的 `server/docs/`（Swagger 2.0，不入库）为实现产物。
> 依据：`docs/v1-spec.md`
> 标记【默认值】的条目为起草时的合理默认，实现前仍可调整。

---

## 一、全局约定

| 项 | 约定 |
|---|---|
| 基础路径 | `/api/v1` |
| 数据格式 | 请求/响应均为 JSON，字段名 snake_case |
| 鉴权 | `Authorization: Bearer <access_token>`；需登录的接口未带/凭证过期 → `401` |
| 成功响应 | 直接返回业务数据，不包 `{data: ...}` 壳 |
| 错误响应 | HTTP 状态码 + `{"code": "<业务码>", "message": "<可读提示>"}` |
| 分页 | 游标分页（见下文），视频列表与评论列表统一采用 |
| CORS | **V1 前后端同源部署，后端不开 CORS**：开发走 Vite 代理（`/api` → `:8080`），生产走 nginx 同域反代；若未来分域名部署再评估开启（2026-08-29 记录） |

### 认证模型（2026-08-29 确认：双 token）

- **access token**：JWT，有效期 **2 小时**【默认值】，无状态不落库，置于 `Authorization: Bearer` 头
- **refresh token**：服务端 CSPRNG 生成的随机串（32 字节，base64url），有效期 **7 天**（沿用已确认值）；**明文不落库，仅存 SHA-256 哈希**
- **轮换**：每次调用刷新接口签发新 refresh token，旧的立即失效；旧 token 再次使用 → `401 REFRESH_TOKEN_INVALID`
- **会话窗口**：过期时间自登录起**固定 7 天，不随刷新滑动延长**【已确认 2026-08-29：固定 / 纯滑动 / 滑动+绝对上限三方案比选后选固定——安全止损线最硬（V1 无登出接口，固定窗口是凭证被盗后唯一的自动止血）、实现最小；代价为活跃用户每 7 天需重新登录一次。比选详情见 `docs/v1-spec.md` 第四节】
- **多设备**【默认值】：每次登录产生独立会话（refresh_tokens 表一行一会话），设备间互不影响；重新登录不吊销其他设备

### 分页约定

- 第一页不传 `cursor`；每页响应返回 `next_cursor`（本页最后一条的 id）
- 下一页请求携带 `next_cursor`；`has_more = false` 时 `next_cursor` 为 `null`
- 排序：主键 id 倒序（自增 id 即创建时间顺序，天然满足"时间倒序"）
- `limit`：默认 20，最大 50，传 0 或省略按默认

### 用户对象（User）

```json
{ "id": 1, "username": "alice", "nickname": "alice" }
```

### 视频摘要（VideoSummary，列表项）

```json
{
  "id": 10,
  "description": "内部年会花絮",
  "like_count": 3,
  "comment_count": 2,
  "author": { "id": 1, "nickname": "alice" },
  "created_at": "2026-08-18T10:00:00+08:00"
}
```

### 视频详情（VideoDetail，播放页）

在摘要基础上增加：

```json
{
  "play_url": "https://oss.example.com/uploads/1/xxx.mp4",
  "liked": false
}
```

`liked`：当前登录用户是否已赞该视频——播放页点赞按钮的实心/空心状态据此渲染；请求携带有效 token 时判断，未登录恒为 `false`（同评论 `is_mine` 语义）（2026-09-07 补充：原契约漏该字段，前端拿到详情后无法画出按钮状态）。登记接口 6 的 201 响应同为 VideoDetail，此时恒为 `false`。

### 评论对象（Comment）

```json
{
  "id": 5,
  "content": "拍得不错",
  "author": { "id": 1, "nickname": "alice" },
  "is_mine": true,
  "created_at": "2026-08-18T10:00:00+08:00"
}
```

`is_mine`：请求携带有效 token 时判断是否本人评论；未登录恒为 `false`。

### 错误码表

| HTTP | code | 场景 |
|---|---|---|
| 400 | `VALIDATION_ERROR` | 参数校验失败（含密码过弱、空评论）；登记视频时 OSS 上文件不存在（2026-09-07） |
| 400 | `UNSUPPORTED_FORMAT` | 视频格式不在白名单 |
| 401 | `UNAUTHORIZED` | 未携带/凭证无效/已过期 |
| 401 | `INVALID_CREDENTIALS` | 登录失败（用户名或密码错误，统一提示） |
| 401 | `REFRESH_TOKEN_INVALID` | refresh token 缺失/无效/已过期/已轮换（2026-08-29 新增） |
| 403 | `FORBIDDEN` | 无权限操作（删他人评论） |
| 404 | `VIDEO_NOT_FOUND` | 视频不存在 |
| 404 | `COMMENT_NOT_FOUND` | 评论不存在 |
| 409 | `USERNAME_EXISTS` | 用户名已存在 |
| 409 | `VIDEO_ALREADY_EXISTS` | 仅并发竞态兜底：同用户重复登记同一 key 幂等返回 200（2026-09-07 修订，见接口 6） |
| 413 | `FILE_TOO_LARGE` | 视频超过 500MB |
| 500 | `INTERNAL` | 服务端内部错误 |

### 与 HTTP 标准的偏离及取舍（2026-08-29 评审记录）

以下为对照 RFC 9110（语义）/ RFC 9457（Problem Details）/ OWASP API Security 后，**有意保留的非标准选择**及其理由，避免日后被误判为疏漏：

1. **错误体不用 RFC 9457 problem+json**：自定义 `{code, message}` 而非 `application/problem+json`（type/title/detail）。理由：前端按 `code` 精确分支即可、Swagger 2.0 生态对 problem+json 支持一般、小范围使用场景从简
2. **游标为裸 id 而非不透明串**：`next_cursor: 8` 直接暴露主键，未做 base64 包装。理由：id 本就存在于响应体中、单键倒序短期不会变、裸 id 便于前端调试；若未来换复合排序键，届时再引入不透明游标（API 形状变更需评估）
3. **201 不带 Location 头**：RFC 9110 建议 201 Created 携带 `Location` 指向新资源，本契约仅返回 body（含 id）。理由：前端从 body 取 id 即可，省一次约定
4. **405/415 不自定义 code**：方法不存在（405）、Content-Type 不支持（415）走 Gin 框架默认响应，错误码表不覆盖。理由：前端不对这两类错误做业务分支
5. **不做 HATEOAS**：响应不含超链接，前端自行拼路由。全行业共识，无需理由

---

## 二、接口清单

### F1/F2 用户认证

#### 1. 注册

`POST /api/v1/auth/register`

```json
// 请求
{ "username": "alice", "password": "secret123" }
// 响应 201
{ "id": 1, "username": "alice", "nickname": "alice" }
```

- `username`：3-32 位，字母/数字/下划线，首字符为字母【默认值】
- `password`：8-72 位（72 为 bcrypt 输入上限，按字节数校验）
- 注册成功返回用户信息（不含 token），前端跳转登录页【默认值：注册不自动登录】
- 冲突：`409 USERNAME_EXISTS`；参数非法：`400 VALIDATION_ERROR`

#### 2. 登录

`POST /api/v1/auth/login`

```json
// 请求
{ "username": "alice", "password": "secret123" }
// 响应 200
{
  "access_token": "eyJhbGciOi...",
  "expires_in": 7200,
  "refresh_token": "dGhpc0lzUmVmcmVzaFRva2Vu...",
  "refresh_expires_in": 604800,
  "user": { "id": 1, "username": "alice", "nickname": "alice" }
}
```

- 用户名或密码任一错误，统一返回 `401 INVALID_CREDENTIALS`（提示不区分具体哪一项）
- 时序等化【2026-08-29 补充】：「用户不存在」与「密码错误」两条路径必须耗时一致，防止通过响应时间枚举有效用户名（timing side channel）。实现方式：用户不存在时也对一个固定的哑 bcrypt 哈希执行一次比对再返回错误（依据：OWASP Authentication Cheat Sheet——防账号枚举要求错误文案与响应时间均不可区分）
- `expires_in`：access token 有效期（秒）；`refresh_expires_in`：refresh token 有效期（秒）
- 每次登录生成独立会话（新 refresh token 一行），不吊销既有会话【默认值：多设备】
- 限流取舍【2026-08-29 记录】：V1 **不在应用层做登录限流**——目标规模小（几百到上千人）、无可枚举面（错误提示已统一）、bcrypt 单次比对 ~100ms 本身即天然减速；若上线后观测到暴力尝试，演进项为 nginx 层 IP 限流（如 5 次/分钟/IP）（依据：OWASP API Security——认证端点应有限流，此为风险接受记录而非遗漏）

#### 3. 刷新凭证（2026-08-29 新增）

`POST /api/v1/auth/refresh`

```json
// 请求
{ "refresh_token": "dGhpc0lzUmVmcmVzaFRva2Vu..." }
// 响应 200
{
  "access_token": "eyJhbGciOi...",
  "expires_in": 7200,
  "refresh_token": "bmV3UmVmcmVzaFRva2Vu...",
  "refresh_expires_in": 432000
}
```

- 轮换：签发新 access token + 新 refresh token，旧 refresh token 立即失效
- 并发刷新同一旧 token（如多标签页同时过期）【2026-09-07 补充】：数据库轮换须用**条件更新**（CAS，见 `database-design.md` 轮换实现），保证一行仅被换新一次；竞争失败方返回 `401 REFRESH_TOKEN_INVALID` 并重新登录。若不加条件，后写者覆盖、先者拿到废票而不自知（下次刷新才被踢）。前端另需刷新"单飞"（见实现备注）
- 响应中的 `refresh_expires_in` 为该会话剩余有效期（固定窗口下随每次刷新递减，如剩 5 天则返回 432000）
- 请求未带 `refresh_token`、token 无效/已过期/已轮换：统一 `401 REFRESH_TOKEN_INVALID`，前端收到此码应清除本地凭证并跳转登录页
- 服务端按 `SHA-256(refresh_token)` 查库比对，明文不落库

#### 4. 当前用户

`GET /api/v1/auth/me`

- 响应 200：`User` 对象；未登录/凭证过期：`401 UNAUTHORIZED`
- 用途：前端刷新页面后校验登录态、拉取昵称；access token 过期时先走接口 3 刷新再调本接口

---

### F3/F4 视频上传与播放

上传流程（三段式，失败重试不产生脏记录）：

```
① POST /api/v1/videos/upload-policy   → 服务端校验（大小/格式）并返回 OSS 签名表单
② 前端 POST 直传阿里云 OSS（multipart/form-data）
③ POST /api/v1/videos                 → 服务端登记记录（含 OSS key），此后视频可见
```

#### 5. 获取上传凭证（登录）

`POST /api/v1/videos/upload-policy`

```json
// 请求
{ "file_name": "team-building.mp4", "size": 104857600, "mime_type": "video/mp4" }
// 响应 200
{
  "upload_url": "https://gotiktok.oss-cn-hangzhou.aliyuncs.com",
  "form_fields": {
    "key": "uploads/1/5f9a8c-abc123.mp4",
    "policy": "eyJleHBpcmF0aW9uIjoi...",
    "signature": "uVwJ8wMq2...",
    "oss_access_key_id": "LTAI5t...",
    "success_action_status": "200"
  },
  "object_key": "uploads/1/5f9a8c-abc123.mp4",
  "expires_at": "2026-08-18T10:05:00+08:00"
}
```

- 服务端校验：`size` > 500MB → `413 FILE_TOO_LARGE`；扩展名/MIME 不在白名单 → `400 UNSUPPORTED_FORMAT`
- 格式白名单【默认值】：`mp4 / mov / webm / m4v / avi / mkv`
- OSS key 规则：`uploads/{user_id}/{uuid}.{ext}`，UUID 由服务端生成，保证唯一
- 签名方式：OSS PostObject policy 签名（有效期 5 分钟）【默认值，无需 STS】
- policy conditions 必须包含以下两条（2026-08-29 修订：原设计仅在第①步服务端校验，凭证本身不限制内容，持有凭证者可绕过校验直传任意大小文件）：
  - `["content-length-range", 0, 524288000]` —— 500MB 上限随签名在 OSS 侧强制生效，超限上传被 OSS 拒绝（依据：OSS 官方文档《PostObject》《POST 请求的策略》，`content-length-range` 条件即用于限制上传大小）
  - `["eq", "$key", "<object_key>"]` —— 锁定本凭证只能写入服务端生成的精确 key，无法改传其他路径（依据：同上，Post Policy conditions 的 `eq` 条件）
- 前端注意：`form_fields` 的键为 snake_case，直传 OSS 时须按 PostObject 表单域规范映射（如 `oss_access_key_id` → `OSSAccessKeyId`）
- 凭证泄漏风险：key 已限定用户与随机名，且 policy 已锁定 key 与大小，即便泄漏他人也只能往该 key 传合规文件

#### 6. 登记视频记录（登录）

`POST /api/v1/videos`

```json
// 请求
{
  "object_key": "uploads/1/5f9a8c-abc123.mp4",
  "description": "内部年会花絮",   // 可选，≤140 字，留空不展示
  "file_name": "team-building.mp4",
  "size": 104857600,
  "mime_type": "video/mp4"
}
// 响应 201：VideoDetail
```

- 服务端校验 `object_key` 必须以 `uploads/{当前用户 id}/` 开头，防止冒用他人 key【默认值】
- 登记前先调用 OSS `HeadObject` 确认文件真实存在【2026-09-07 补充】：对象不存在（上传实际未成功或 key 伪造）→ `400 VALIDATION_ERROR`（提示重新上传），不落库——杜绝"列表出现永远播不了的死链视频"。落库口径【2026-09-08 补充】：`file_size` 以 HeadObject 返回的实测大小为准（客户端自报仅供预检，防列表 size 失真）；`mime_type` 保留请求提交值（OSS 的 Content-Type 随浏览器表单变化、不可靠，不作事实源）
- `description` 留空则存 NULL，前端不展示该行
- 重复提交同一 key【2026-09-07 修订：幂等优先，409 仅作竞态兜底】：该 key 已有记录且属当前用户（登记成功但响应丢失后的重试）→ **幂等返回 `200` + 已有 VideoDetail**，不重复插入、不报错；仅当"查无记录但插入时撞唯一键"的并发竞态才返回 `409 VIDEO_ALREADY_EXISTS`【2026-08-29 定名，对齐 USERNAME_EXISTS 命名模式】
- 上传中断/失败：只产生 OSS 孤儿对象（无 DB 记录），重试即可，无脏数据

#### 7. 视频列表（公开）

`GET /api/v1/videos?cursor=&limit=`

- 响应 200：

```json
{
  "items": [ VideoSummary, ... ],
  "next_cursor": 8,
  "has_more": true
}
```

- 无需鉴权【已确认】；排序：id 倒序【已确认】

#### 8. 视频详情/播放（公开）

`GET /api/v1/videos/:id`

- 响应 200：`VideoDetail`（`liked` 见上文对象定义，未登录恒 `false`）；不存在：`404 VIDEO_NOT_FOUND`
- `play_url` 为 OSS 对象完整 URL【默认值：bucket 公共读，后端直接拼 URL】
- 源文件不可用：OSS 返回 404/403，前端播放器展示"视频加载失败"而非白屏

---

### F5 点赞

#### 9. 点赞（登录，幂等）

`POST /api/v1/videos/:id/like`

- 已赞过重复点赞：不产生新记录，仍返回 200（幂等）
- 响应 200：`{ "liked": true, "like_count": 4 }`
- 视频不存在：`404 VIDEO_NOT_FOUND`；未登录：`401 UNAUTHORIZED`

#### 10. 取消点赞（登录，幂等）

`DELETE /api/v1/videos/:id/like`

- 未赞过也返回 200（幂等）
- 响应 200：`{ "liked": false, "like_count": 3 }`

- 计数直接读库 COUNT，不引入 Redis【已确认】

---

### F6 评论

#### 11. 评论列表（公开）

`GET /api/v1/videos/:id/comments?cursor=&limit=`

- 响应 200：

```json
{
  "items": [ Comment, ... ],
  "next_cursor": 3,
  "has_more": false
}
```

- 无需鉴权；排序：id 倒序（最新在前）

#### 12. 发表评论（登录）

`POST /api/v1/videos/:id/comments`

```json
// 请求
{ "content": "拍得不错" }
// 响应 201：Comment
```

- `content`：非空、≤500 字【默认值：长度上限】，空串 → `400 VALIDATION_ERROR`
- 视频不存在：`404 VIDEO_NOT_FOUND`；未登录：`401 UNAUTHORIZED`

#### 13. 删除评论（登录，仅限本人）

`DELETE /api/v1/comments/:id`

- 响应 204 无 body
- 评论不存在：`404 COMMENT_NOT_FOUND`；非本人：`403 FORBIDDEN`；未登录：`401 UNAUTHORIZED`

---

## 三、接口与功能点对照

| 功能点 | 接口 |
|---|---|
| F1 注册 | 1 |
| F2 登录 | 2、3、4 |
| F3 上传 | 5、6 |
| F4 播放 | 7、8 |
| F5 点赞 | 9、10 |
| F6 评论 | 11、12、13 |

## 四、实现备注

- 路由分组：`/api/v1/auth`、`/api/v1/videos`、`/api/v1/videos/:id/like`、`/api/v1/videos/:id/comments`、`/api/v1/comments`
- 需登录的接口统一挂 JWT 鉴权中间件（只校验 access token）；`/auth/me` 返回 401 而非重定向
- 可选鉴权【2026-09-08 补充】：接口 8（视频详情）与接口 11（评论列表）为公开接口，但响应含当前用户视角字段（`liked` / `is_mine`）——挂**可选鉴权**中间件：带有效 token 则解析出当前用户并填充字段；未带/无效按匿名处理（恒 `false`），**不得返回 401**
- 前端拦截器约定：业务请求遇 `401 UNAUTHORIZED` → 携带 refresh_token 调接口 3 → 成功则重放原请求；刷新又失败（`401 REFRESH_TOKEN_INVALID`）→ 清除凭证跳登录页
- 多标签页竞态【2026-09-07 修订】：双开页签可能同时持同一旧 refresh token 刷新。若 UPDATE 不带条件，两请求都会"成功"、后写者覆盖先者的新哈希——先者实际拿到废票，下次刷新才被 401（体验是莫名登出）。修复：轮换 UPDATE 带 `AND token_hash = 旧值` 条件（CAS/乐观锁），仅一方换新成功、另一方影响行数 0 → 返回 `401 REFRESH_TOKEN_INVALID`；前端刷新另须全局单飞（加锁并发只发一次）
- 上传前校验（大小/格式）与登记时的二次校验一致，双保险
