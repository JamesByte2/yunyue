"""云约后端冒烟测试：对运行中的服务做一次完整业务链路验证。

用法：
    cd frontend && npm install   # 前端单独构建
    cd backend && go build . && ./yunyue-api    # 或设置环境变量后运行
    python scripts/smoke.py [base_url]

覆盖：登录、RBAC、服务/技师管理、预约冲突检测、状态流转、
会员储值折扣结算、流水记账、看板统计。
"""
import sys
import time
import urllib.request
import json

BASE = sys.argv[1] if len(sys.argv) > 1 else "http://127.0.0.1:8010"
TODAY = time.strftime("%Y-%m-%d")
passed = 0


def req(method, path, token=None, body=None, expect=200):
    r = urllib.request.Request(BASE + path, method=method)
    if token:
        r.add_header("Authorization", f"Bearer {token}")
    data = json.dumps(body).encode() if body is not None else None
    if data:
        r.add_header("Content-Type", "application/json")
    try:
        resp = urllib.request.urlopen(r, data, timeout=15)
        code = resp.status
        payload = json.loads(resp.read())
    except urllib.error.HTTPError as e:
        code = e.code
        payload = json.loads(e.read())
    assert code == expect, f"{method} {path}: got {code}, want {expect}: {payload}"
    return payload


def step(name, ok, detail=""):
    global passed
    print(f"{'PASS' if ok else 'FAIL'}  {name}  {detail}")
    assert ok, name
    passed += 1


step("health", "status" in req("GET", "/health"))

admin = req("POST", "/api/auth/login", body={"email": "admin@yunyue.cn", "password": "admin12345"})
step("admin login", admin["role"] == "admin")
staff = req("POST", "/api/auth/login", body={"email": "staff@yunyue.cn", "password": "staff12345"})
step("staff login", staff["role"] == "staff")

try:
    req("POST", "/api/services", token=staff["token"], body={"name": "x", "duration_min": 10, "price_cents": 100}, expect=403)
    step("RBAC staff forbidden on write", True)
except AssertionError:
    raise

svc = req("POST", "/api/services", token=admin["token"], body={"name": "冒烟精剪", "duration_min": 30, "price_cents": 8800})
step("create service", svc["id"] > 0, f"id={svc['id']}")
stf = req("POST", "/api/staff", token=admin["token"], body={"name": "冒烟技师", "title": "发型师", "phone": "13900000009"})
step("create staff", stf["id"] > 0)

b1 = req("POST", "/api/bookings", token=admin["token"], body={
    "customer_name": "冒烟客户", "customer_phone": "13900000000",
    "service_item_id": svc["id"], "staff_id": stf["id"],
    "book_date": TODAY, "start_min": 14 * 60, "remark": "smoke",
})
step("create booking pending", b1["status"] == "pending", f"id={b1['id']}")

conflict = req("POST", "/api/bookings", token=admin["token"], body={
    "customer_name": "冲突客户", "customer_phone": "13900000000",
    "service_item_id": svc["id"], "staff_id": stf["id"],
    "book_date": TODAY, "start_min": 14 * 60 + 15,
}, expect=409)
step("conflict detected 409", "conflictId" in conflict)

req("PATCH", f"/api/bookings/{b1['id']}/status", token=admin["token"], body={"status": "done"}, expect=400)
step("invalid transition locked", True)

req("PATCH", f"/api/bookings/{b1['id']}/status", token=admin["token"], body={"status": "confirmed"})
step("confirm booking", True)

m = req("POST", "/api/members", token=admin["token"], body={"name": "冒烟会员", "phone": "13900000008", "discount_percent": 90})
step("create member 9折", m["discount_percent"] == 90)

r = req("POST", f"/api/members/{m['id']}/recharge", token=admin["token"], body={"amount_cents": 100000})
step("recharge 1000元", r["member"]["balance_cents"] == 100000)

fin = req("POST", f"/api/bookings/{b1['id']}/finish", token=admin["token"], body={"member_id": m["id"]})
expected_pay = 8800 * 90 // 100  # 9折 = 7920
step("finish with member pay", fin["status"] == "done" and fin["payable_cents"] == expected_pay, f"pay={fin['payable_cents']}")

txns = req("GET", f"/api/members/{m['id']}/transactions", token=admin["token"])
step("member ledger 2 entries", len(txns) == 2 and txns[0]["type"] == "consume" and txns[0]["balance_after"] == 100000 - expected_pay)

dash = req("GET", "/api/dashboard", token=admin["token"])
step("dashboard counts", dash["member_count"] >= 2 and dash["today_bookings"] >= 1, f"today={dash['today_bookings']} members={dash['member_count']}")

print(f"\n{passed} steps passed — 云约业务链路完整可用")
