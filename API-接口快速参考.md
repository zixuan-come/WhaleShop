# WhaleShop 接口快速参考

WhaleShop 是一个 Go 编写的订单服务，提供健康检查、订单 CRUD，以及用于接口测试和压测的慢接口、错误接口。

## 1. 服务地址

根据调用位置选择基础地址：

| 场景 | Base URL |
|---|---|
| 云服务器公网 | `http://1.12.53.100:8081` |
| 云服务器本机 | `http://127.0.0.1:8081` |
| 本地直接运行 | `http://localhost:8080` |
| WhaleTestPro 本地 Docker 容器访问本机 WhaleShop | `http://host.docker.internal:8080` |

WhaleShop 没有 `/api` 或 `/api/v1` 前缀，路径直接从根路径开始。

公网健康检查示例：

```bash
curl -i http://1.12.53.100:8081/health
```

预期响应：

```json
{"status":"ok"}
```

## 2. 通用约定

### 2.1 项目隔离 Header

除 `GET /health` 外，其余接口都支持 `X-Project-Id`：

```http
X-Project-Id: 1
```

- 不传：使用默认项目 `1`。
- 传入正整数：使用对应项目的数据空间。
- 非数字、`0` 或负数：回落到项目 `1`。
- 不同项目的订单数据相互隔离。

### 2.2 JSON 格式

发送 JSON 请求体时使用：

```http
Content-Type: application/json
```

业务处理函数在 HTTP 允许响应体时返回 JSON，常见业务错误格式为：

```json
{"error":"错误说明"}
```

`204`、`304` 不携带响应体，不能对它们执行 JSON 响应体断言。未匹配路由或方法产生的 `404`、`405` 可能由 Go 路由器返回纯文本，不属于上述业务错误格式。

### 2.3 常用状态码

| 状态码 | 含义 |
|---:|---|
| 200 | 查询、更新或删除成功 |
| 201 | 创建成功 |
| 400 | 路径参数或 JSON 格式错误 |
| 404 | 订单不存在 |
| 409 | 业务冲突，例如已取消订单不能修改 |
| 422 | 创建订单的字段校验失败 |
| 5xx | 服务器错误或错误接口主动返回的状态码 |

## 3. 接口总览

| 方法 | 路径 | 用途 |
|---|---|---|
| GET | `/health` | 健康检查 |
| GET | `/orders` | 查询当前项目全部订单 |
| GET | `/orders/{id}` | 查询单个订单 |
| POST | `/orders` | 创建订单 |
| PUT | `/orders/{id}` | 全量更新订单 |
| DELETE | `/orders/{id}` | 删除订单 |
| GET | `/orders/slow` | 延迟后返回，用于压测/超时测试 |
| GET | `/orders/error` | 返回指定状态码，用于错误断言/重试测试 |

## 4. 订单数据结构

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

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | integer | 服务端生成的订单编号 |
| `item` | string | 商品名；创建时不能为空 |
| `quantity` | integer | 数量；创建时必须大于 0 |
| `price` | number | 价格，可为小数 |
| `status` | string | 状态；创建时不传默认为 `pending` |
| `created_at` | string | 服务端生成的 ISO 8601 时间 |

建议使用的状态值：`pending`、`paid`、`shipped`、`cancelled`。

注意：数据保存在内存中，容器重启后会恢复默认种子数据，运行期间创建的数据不会持久化。

## 5. 接口详情

以下示例使用云服务器地址；本机测试时可将 `BASE_URL` 换成 `http://127.0.0.1:8081`。

```bash
BASE_URL=http://1.12.53.100:8081
```

### 5.1 健康检查

```http
GET /health
```

```bash
curl -i "$BASE_URL/health"
```

响应 `200`：

```json
{"status":"ok"}
```

### 5.2 查询订单列表

```http
GET /orders
X-Project-Id: 1
```

```bash
curl -i "$BASE_URL/orders"
curl -i -H "X-Project-Id: 2" "$BASE_URL/orders"
```

成功响应 `200` 为数组；没有订单时返回 `[]`。

### 5.3 查询单个订单

```http
GET /orders/{id}
```

```bash
curl -i "$BASE_URL/orders/1"
```

- `200`：返回订单对象。
- `400`：例如 `/orders/abc`，返回 `{"error":"id 必须是整数"}`。
- `404`：订单不存在，返回 `{"error":"订单不存在"}`。

### 5.4 创建订单

```http
POST /orders
Content-Type: application/json
```

请求体：

```json
{
  "item": "Kindle",
  "quantity": 1,
  "price": 999,
  "status": "pending"
}
```

```bash
curl -i -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -H "X-Project-Id: 1" \
  -d '{"item":"Kindle","quantity":1,"price":999,"status":"pending"}'
```

- `201`：创建成功；`id`、`created_at` 由服务端生成。
- `400`：请求体不是合法 JSON。
- `422`：`item` 为空或 `quantity <= 0`。

### 5.5 全量更新订单

```http
PUT /orders/{id}
Content-Type: application/json
```

```bash
curl -i -X PUT "$BASE_URL/orders/2" \
  -H "Content-Type: application/json" \
  -H "X-Project-Id: 1" \
  -d '{"item":"AirPods Pro 2","quantity":3,"price":1899,"status":"shipped"}'
```

这是全量替换语义：`id` 和 `created_at` 保留，未传的其他字段会变为零值。当前服务未实现 PATCH。

- `200`：更新成功。
- `400`：`id` 不是整数或 JSON 非法。
- `404`：订单不存在。
- `409`：原订单状态为 `cancelled`，返回 `{"error":"已取消的订单不能修改"}`。

### 5.6 删除订单

```http
DELETE /orders/{id}
```

```bash
curl -i -X DELETE "$BASE_URL/orders/3" \
  -H "X-Project-Id: 1"
```

- `200`：`{"message":"ok"}`。
- `400`：`id` 不是整数。
- `404`：订单不存在。

### 5.7 慢接口

```http
GET /orders/slow?ms=500
```

`ms` 是阻塞毫秒数：默认 `500`，有效范围 `0~30000`；超出范围或不是数字时使用默认值。

```bash
curl -i "$BASE_URL/orders/slow"
curl -i "$BASE_URL/orders/slow?ms=2000"
```

响应 `200`：

```json
{"slept_ms":2000}
```

### 5.8 指定错误码接口

```http
GET /orders/error?code=500
```

`code` 有效范围是 `100~599`；缺失、超出范围或不是数字时使用 `500`。

```bash
curl -i "$BASE_URL/orders/error?code=418"
curl -i "$BASE_URL/orders/error?code=503"
```

当指定 `200~599` 中允许响应体的状态码时，响应体为：

```json
{"error":"触发指定错误码"}
```

HTTP 状态码分别为 `418`、`503`。

边界：虽然参数解析接受 `100~199`，这些属于信息性响应，不能保证客户端最终收到的状态码等于指定值；`101` 涉及协议切换，本接口没有实现协议升级。指定 `204`、`304` 时不会返回上述 JSON 响应体。用于错误断言和重试测试时，建议选择 `4xx` 或 `5xx`，例如上述 `418`、`503`。HTTP 响应写入规则见 [Go net/http 官方说明](https://pkg.go.dev/net/http#ResponseWriter)。

## 6. 推荐验证顺序

```bash
# 1. 服务健康
curl -i "$BASE_URL/health"

# 2. 查询种子订单
curl -i "$BASE_URL/orders"

# 3. 创建订单
curl -i -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{"item":"测试商品","quantity":1,"price":10}'

# 4. 验证慢接口
curl -i "$BASE_URL/orders/slow?ms=100"

# 5. 验证错误码
curl -i "$BASE_URL/orders/error?code=418"
```

WhaleTestPro 中配置环境时，云部署地址使用 `http://1.12.53.100:8081`；接口路径只填写 `/health`、`/orders` 等路径，不要重复填写 Base URL。

## 7. 相关文档

- [完整接口文档](API.md)
- [WhaleShop GitHub 仓库](https://github.com/zixuan-come/WhaleShop)
