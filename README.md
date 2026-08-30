# 云约 YunYue · 门店预约管理系统

面向中小门店（美业/家政/场馆）的预约管理 SaaS 后台：服务项目管理、技师排班预约（时段冲突检测）、会员储值卡（折扣结算 + 流水台账）、经营看板。

**在线演示**：http://8.216.24.166:8080 （演示账号：`admin@yunyue.cn` / `admin12345`；技师角色：`staff@yunyue.cn` / `staff12345`）

> 技术栈：Go (Gin + GORM) + MySQL + Vue 3 + TypeScript + Element Plus
> 与 KnowBase（AI 知识库）构成互补的作品组合：一个展示 AI 工程能力，一个展示业务系统交付能力。

## 核心业务规则

- **预约冲突检测**：同技师同日期，时间段 `[start, start+时长)` 与任何未完结预约重叠即拒绝（HTTP 409），冲突时返回冲突单号。
- **状态机**：`pending → confirmed → done / canceled`，终态锁定，非法流转返回 400。
- **会员结算**：完成预约时可选会员卡支付，按会员折扣（如 90 = 九折）计费，扣款与消费流水在**同一数据库事务**内完成；余额不足整体失败。
- **流水台账**：充值/消费均记录 `balance_after`（变动后余额快照），可审计；充值侧为模拟支付网关，对接微信/支付宝时只需替换记账前的回调入口。
- **RBAC**：admin 全权；staff 只读服务/技师/会员，可操作预约，写操作返回 403。
- **数据可追溯**：有历史预约的服务/技师不物理删除，自动转为停用。

## 快速开始

```bash
# 数据库：准备 MySQL，创建数据库 yunyue（表结构首次启动自动迁移，并写入演示数据）
cd backend
go build -o yunyue-api .
DB_DSN="root:密码@tcp(127.0.0.1:3306)/yunyue?charset=utf8mb4&parseTime=True" ./yunyue-api

cd frontend
npm install && npm run dev   # http://localhost:5174
```

## 冒烟测试（15 步业务链路）

```bash
python backend/scripts/smoke.py http://127.0.0.1:8010
```

覆盖：登录、RBAC 403、服务/技师创建、预约创建、**同段冲突 409**、非法状态流转 400 锁定、确认流转、会员九折创建、充值记账、会员卡折扣结算（8800 分 → 7920 分）、双流水余额快照、看板统计。

## 部署架构（1GB 内存轻量服务器）

```
nginx :8080 ──静态──▶ Vue3 构建产物
   │ /api/ 反代
   ▼
Go 单二进制（systemd，127.0.0.1:8010，常驻内存约 30MB）
   └── MySQL（与 KnowBase 共享实例，独立数据库 yunyue）
```

Go 交叉编译出单文件部署，无运行时依赖，适配小内存机器。
