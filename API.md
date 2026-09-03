# WhaleShop API 文档

极简订单服务，作为 WhaleTestPro 的被测系统（SUT）。

- **Base URL（宿主机）**：`http://localhost:8080`
- **Base URL（从 WhaleTestPro 容器）**：`http://host.docker.internal:8080`
- **无 URL 前缀**：所有接口挂根路径下（不是 `/api/v1`）
- **Content-Type**：所有 JSON body 请求响应均为 `application/json; charset=utf-8`

## 通用约定

### 多租户 Header

除 `/health` 外，所有接口读 `X-Project-Id`：

| 情况 | 结果 |
|------|------|
| 不带 header | 默认 `pid=1`（默认租户）|
| `X-Project-Id: 2` | 路由到独立租户 2，数据完全隔离 |
| 非数字/负数 | 回落 `pid=1` |

### 错误响应统一格式

```json
{"error": "订单不存在"}
```

配合 HTTP 状态码使用（400/404/409/422/500 等）。

### 状态码语义

| 码 | 触发场景 |
|----|---------|
| 200 | 查询成功 / 更新成功 / 删除成功 |
| 201 | 创建成功 |
| 400 | 路径参数不合法（如 id 不是整数）/ 请求体不是合法 JSON |
| 404 | 资源不存在 |
| 409 | 业务规则冲突（如已取消订单不能改）|
| 422 | 字段校验失败（如 item 空、quantity ≤ 0）|
| 500 | 服务器内部错（正常情况不会出）|

---

## 1. GET /health

健康检查，不带任何认证，不读 X-Project-Id。

**请求**
```
GET /health
```

**响应 · 200**
```json
{"status": "ok"}
```

**curl**
```bash
curl http://localhost:8080/health
```

---

## 2. GET /orders — 列出订单

按当前租户返回全部订单。

**请求**
```
GET /orders
Headers: X-Project-Id: <int, optional>
```

**响应 · 200**
```json
[
  {
    "id": 1,
    "item": "iPhone 15",
    "quantity": 1,
    "price": 6999,
    "status": "paid",
    "created_at": "2026-07-04T03:02:13Z"
  },
  {
    "id": 2,
    "item": "AirPods Pro",
    "quantity": 2,
    "price": 1899,
    "status": "shipped",
    "created_at": "2026-07-04T03:02:13Z"
  }
]
```

租户空时返回 `[]`（空数组），不是 404。

**curl**
```bash
# 默认租户
curl http://localhost:8080/orders

# 切租户
curl -H "X-Project-Id: 2" http://localhost:8080/orders
```

---

## 3. GET /orders/{id} — 查单个订单

**请求**
```
GET /orders/1
Headers: X-Project-Id: <int, optional>
```

**响应 · 200**
```json
{
  "id": 1,
  "item": "iPhone 15",
  "quantity": 1,
  "price": 6999,
  "status": "paid",
  "created_at": "2026-07-04T03:02:13Z"
}
```

**响应 · 404**
```json
{"error": "订单不存在"}
```

**响应 · 400**（id 不是整数）
```json
{"error": "id 必须是整数"}
```

**curl**
```bash
curl http://localhost:8080/orders/1
curl http://localhost:8080/orders/999    # 404
curl http://localhost:8080/orders/abc    # 400
```

---

## 4. POST /orders — 创建订单

**请求**
```
POST /orders
Headers: X-Project-Id: <int, optional>
Body:
{
  "item":     "Kindle",           // required, string
  "quantity": 1,                  // required, int > 0
  "price":    999.00,             // optional, float
  "status":   "pending"           // optional, 缺省 "pending"
}
```

`id` 和 `created_at` 服务端生成，请求里传了会被覆盖。

**响应 · 201**
```json
{
  "id": 4,
  "item": "Kindle",
  "quantity": 1,
  "price": 999,
  "status": "pending",
  "created_at": "2026-07-04T03:03:08Z"
}
```

**响应 · 422**（item 空或 quantity ≤ 0）
```json
{"error": "item 不能为空"}
```
```json
{"error": "quantity 必须 > 0"}
```

**响应 · 400**（body 不是合法 JSON）
```json
{"error": "请求体不是合法 JSON"}
```

**curl**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"item":"Kindle","quantity":1,"price":999,"status":"pending"}'
```

---

## 5. PUT /orders/{id} — 全量更新订单

**请求**
```
PUT /orders/2
Headers: X-Project-Id: <int, optional>
Body:
{
  "item":     "AirPods Pro 2",
  "quantity": 3,
  "price":    1899,
  "status":   "shipped"
}
```

**语义**：全量替换（除 `id`、`created_at` 保留）。**没传的字段会被清空**——如果只想改一个字段用 PATCH（本服务未实现 PATCH，需要就整体传）。

**响应 · 200**
```json
{
  "id": 2,
  "item": "AirPods Pro 2",
  "quantity": 3,
  "price": 1899,
  "status": "shipped",
  "created_at": "2026-07-04T03:02:13Z"
}
```

**响应 · 404**
```json
{"error": "订单不存在"}
```

**响应 · 409**（已取消订单不能改）
```json
{"error": "已取消的订单不能修改"}
```

**响应 · 400**（id 不是整数 / body 不是 JSON）— 同 GET/POST

**curl**
```bash
curl -X PUT http://localhost:8080/orders/2 \
  -H "Content-Type: application/json" \
  -d '{"item":"AirPods Pro 2","quantity":3,"price":1899,"status":"shipped"}'
```

---

## 6. DELETE /orders/{id} — 删除订单

**请求**
```
DELETE /orders/3
Headers: X-Project-Id: <int, optional>
```

**响应 · 200**
```json
{"message": "ok"}
```

**响应 · 404**
```json
{"error": "订单不存在"}
```

**响应 · 400**（id 不是整数）
```json
{"error": "id 必须是整数"}
```

**curl**
```bash
curl -X DELETE http://localhost:8080/orders/3
```

---

## 7. GET /orders/slow — 慢接口（压测靶子）

阻塞指定毫秒后返回 200。用于 Locust 压测/超时熔断演示。

**请求**
```
GET /orders/slow?ms=500
Headers: X-Project-Id: <int, optional>
```

**Query 参数**

| 参数 | 类型 | 默认 | 范围 | 说明 |
|-----|------|------|-----|------|
| ms | int | 500 | 0 ~ 30000 | 阻塞毫秒数；超出范围/非数字则用默认 |

**响应 · 200**
```json
{"slept_ms": 500}
```

**curl**
```bash
# 默认 500ms
curl http://localhost:8080/orders/slow

# 自定义 2 秒
time curl http://localhost:8080/orders/slow?ms=2000

# 用于压测（Locust host: http://host.docker.internal:8080）
locust -f locustfile.py --host http://host.docker.internal:8080
```

---

## 8. GET /orders/error — 强返错误码（断言靶子）

强制返回指定 HTTP 状态码，用于练平台的断言、失败重试、告警。

**请求**
```
GET /orders/error?code=500
Headers: X-Project-Id: <int, optional>
```

**Query 参数**

| 参数 | 类型 | 默认 | 范围 | 说明 |
|-----|------|------|-----|------|
| code | int | 500 | 100 ~ 599 | 期望返回的 HTTP 码；超范围用默认 |

**响应 · 任意码**
```json
{"error": "触发指定错误码"}
```

**curl**
```bash
curl -w " [HTTP %{http_code}]\n" http://localhost:8080/orders/error?code=418
# {"error":"触发指定错误码"} [HTTP 418]

curl -w " [HTTP %{http_code}]\n" http://localhost:8080/orders/error?code=503
# {"error":"触发指定错误码"} [HTTP 503]
```

---

## 数据模型

### Order

```go
type Order struct {
    ID        int       // 服务端生成，自增（每租户独立）
    Item      string    // 必填，商品名
    Quantity  int       // 必填 > 0
    Price     float64   // 可选，价格
    Status    string    // 可选，缺省 "pending"；建议:pending/paid/shipped/cancelled
    CreatedAt time.Time // 服务端生成，ISO8601
}
```

### 种子数据（pid=1 自带）

启动时默认租户预置 3 条：

| id | item | quantity | price | status |
|----|------|---------|-------|--------|
| 1 | iPhone 15 | 1 | 6999.00 | paid |
| 2 | AirPods Pro | 2 | 1899.00 | shipped |
| 3 | MacBook Air | 1 | 8999.00 | pending |

**注意**：数据在内存里，容器重启后种子重新加载，你后来 POST 的数据丢失。这是 demo 环境的取舍。

---

## 从 WhaleTestPro 接入

WhaleTestPro 的环境已经设置好：

```
environment.base_url = http://host.docker.internal:8080
```

在接口管理页只填**路径**（如 `/orders`），执行时引擎自动拼成 `http://host.docker.internal:8080/orders`。

**推荐练手接口列表：**

| 平台里配置 | 演示的能力 |
|-----------|-----------|
| GET /orders + assert `status == 200` | 最简 GET |
| GET /orders/1 + assert `json_eq item "iPhone 15"` | JSON 字段断言 |
| GET /orders/999 + assert `status == 404` | 4xx 断言 |
| POST /orders body 合法 + assert `status == 201` | POST + Created |
| POST /orders body 缺 item + assert `status == 422` | 校验失败 |
| GET /orders/slow?ms=800 → 用例超时 500ms | 超时/熔断 |
| GET /orders/error?code=500 | 失败重试演示 |
| PUT /orders/1 → 建 chain 引用上一步的 id | 上下文串联 |
