# 复杂业务跨领域协作示例

## 一、文档定位

本文用于固定“一个主业务域跨多个领域协作”时的推荐组织方式。后续类似复杂业务默认参考本文，不再各自发挥。

## 二、适用场景

当一个复杂业务动作同时涉及以下任一领域时，应按本文组织代码：

- 商品
- 用户
- 钱包
- 库存
- 优惠券
- 营销活动
- 订单日志
- 通知或异步任务

典型动作包括：

- 创建订单
- 支付订单
- 取消订单
- 超时关闭订单
- 售后退款
- 密码登录
- 短信登录
- 登录风控校验

## 三、通用组织原则

- 先确定一个主业务域，由主业务域承担 `biz` 总编排职责。
- 其他参与领域只暴露本领域 `service` 能力。
- 跨领域事务边界优先收口在主业务域 `biz`。
- 不允许把跨领域流程塞进某个领域 `service` 内部链式互调。

## 四、订单示例

### 4.1 目录组织

推荐目录：

```text
internal/module/
├── order/
│   ├── biz/
│   ├── convert/
│   ├── dto/
│   ├── model/
│   ├── repository/
│   └── service/
├── product/
├── user/
└── wallet/
```

职责说明：

- `order/biz`：订单主业务流程编排层。
- `order/service`：订单领域本身的规则和状态流转。
- `order/repository`：订单数据访问。
- `order/convert`：订单领域对象转换。
- `product/service`：商品领域能力。
- `user/service`：用户领域能力。
- `wallet/service`：钱包领域能力。

### 4.2 调用关系

推荐调用关系：

```text
handler
  → order/biz
      → order/service
      → product/service
      → user/service
      → wallet/service
      → order/repository
```

核心原则：

- 订单是主业务域时，由 `order/biz` 负责总编排。
- 其他领域只提供本领域能力，不反向主导订单流程。
- 跨领域协作不放进 `order/service`。

### 4.3 下单示例

下单动作示例：

```text
CreateOrderHandler
  → order/biz.CreateOrder
      → user/service.CheckUserAvailable
      → product/service.GetSaleSnapshot
      → wallet/service.CheckBalance
      → order/service.BuildOrder
      → order/service.SaveOrder
      → wallet/service.FreezeAmount
      → product/service.LockStock
```

可以进一步拆成以下职责：

- `user/service.CheckUserAvailable`
  校验用户是否存在、是否冻结、是否允许下单。
- `product/service.GetSaleSnapshot`
  获取商品快照、价格、当前售卖状态。
- `wallet/service.CheckBalance`
  校验钱包余额或支付能力。
- `order/service.BuildOrder`
  生成订单快照、金额、状态等领域对象。
- `order/service.SaveOrder`
  保存订单主记录和订单明细。
- `wallet/service.FreezeAmount`
  冻结或扣减待支付金额。
- `product/service.LockStock`
  锁定库存或扣减可售库存。

### 4.4 事务边界

跨领域事务边界优先放在 `order/biz`：

- 订单创建成功但钱包冻结失败，应由 `order/biz` 统一回滚。
- 钱包冻结成功但库存锁定失败，也应由 `order/biz` 统一回滚。
- 不允许多个 Service 各自开事务，再由上层碰运气拼起来。

规则：

- 单领域内部事务可以放在本领域 `service`。
- 只要跨多个领域，就优先上移到发起方领域的 `biz`。

### 4.5 禁止事项

- 不允许 `order/service -> wallet/service -> product/service` 这种链式互调。
- 不允许把商品、用户、钱包规则直接写进 `order/repository`。
- 不允许把跨领域流程散落到多个 Handler 中。
- 不允许为了图省事，让钱包或商品领域反向主导订单流程。

## 五、登录示例

### 5.1 主业务域选择

登录动作建议单独视为 `auth` 主业务域，而不是直接塞进 `user/service`。

推荐目录：

```text
internal/module/
├── auth/
│   ├── biz/
│   ├── convert/
│   ├── dto/
│   ├── model/
│   ├── repository/
│   └── service/
├── user/
└── common/
```

说明：

- `auth/biz`：登录总编排层。
- `auth/service`：登录领域本身的令牌生成、登录态处理等能力。
- `user/service`：用户密码校验、用户状态校验等用户领域能力。
- `system/service`：系统设置、风控参数、登录日志等后台通用能力。
- `sms/service`：短信验证码校验能力。如果短信是外部基础设施，也可以落在 `pkg` 或独立集成模块，但对 `auth/biz` 来说它表现为一个被调用能力。

### 5.2 调用关系

推荐调用关系：

```text
handler
  → auth/biz
      → sms/service
      → user/service
      → system/service
      → auth/service
```

### 5.3 登录动作示例

以“登录接口需要验证短信验证码、校验用户密码、读取系统风控参数、记录登录日志”为例：

```text
LoginHandler
  → auth/biz.Login
      → system/service.GetLoginRiskConfig
      → sms/service.VerifyCode
      → user/service.CheckPassword
      → user/service.CheckUserStatus
      → auth/service.GenerateToken
      → system/service.CreateLoginLog
```

可以进一步拆成以下职责：

- `system/service.GetLoginRiskConfig`
  读取登录相关风控参数，例如验证码开关、失败次数限制、IP 限制、设备限制。
- `sms/service.VerifyCode`
  校验短信验证码是否正确、是否过期、是否匹配当前手机号或场景。
- `user/service.CheckPassword`
  校验用户名密码是否正确。
- `user/service.CheckUserStatus`
  校验用户是否冻结、禁用、注销，是否允许登录。
- `auth/service.GenerateToken`
  生成 JWT 或其他登录态凭证。
- `system/service.CreateLoginLog`
  记录登录日志，包括用户、结果、IP、设备、原因等。

### 5.4 组织约束

- 短信验证码校验、密码校验、风控参数读取、登录日志记录都由 `auth/biz` 编排。
- 不把短信校验、风控参数、登录日志直接揉进 `user/service`。
- 不允许形成 `user/service -> system/service -> sms/service` 这种反向串联。
- 登录成功后的 token 生成、登录态保存等能力可以放在 `auth/service`。
- 如果同一个登录接口同时支持密码登录、短信登录、免密登录，也应由 `auth/biz` 统一分发流程。

## 六、推广规则

这套模式不只适用于订单。凡是“某个领域作为主业务域，协调多个其他领域”的复杂场景，都按同样方式处理：

- 发起方领域负责 `biz`
- 各参与领域只暴露自己的 `service`
- 跨领域事务和流程编排统一收口在主领域 `biz`
