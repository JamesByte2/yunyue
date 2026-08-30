<template>
  <el-card>
    <div class="toolbar">
      <span>会员管理（储值卡）</span>
      <el-button type="primary" @click="openCreate">新增会员</el-button>
    </div>
    <el-table :data="list" stripe>
      <el-table-column prop="id" label="#" width="60" />
      <el-table-column prop="name" label="姓名" width="120" />
      <el-table-column prop="phone" label="电话" width="150" />
      <el-table-column label="余额" width="120">
        <template #default="{ row }">¥{{ money(row.balance_cents) }}</template>
      </el-table-column>
      <el-table-column label="折扣" width="100">
        <template #default="{ row }">{{ row.discount_percent / 10 }} 折</template>
      </el-table-column>
      <el-table-column label="操作">
        <template #default="{ row }">
          <el-button size="small" type="success" @click="openRecharge(row)">充值</el-button>
          <el-button size="small" @click="showTxns(row)">流水</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="createOpen" title="新增会员" width="420px">
      <el-form label-width="90px">
        <el-form-item label="姓名"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="电话"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="折扣（%）">
          <el-input-number v-model="form.discount_percent" :min="50" :max="100" :step="5" />
          <span class="hint">{{ form.discount_percent / 10 }} 折</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createOpen = false">取消</el-button>
        <el-button type="primary" @click="create">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="rechargeOpen" title="会员充值（模拟支付网关）" width="420px">
      <p>{{ current?.name }} 当前余额 ¥{{ money(current?.balance_cents ?? 0) }}</p>
      <el-form label-width="90px">
        <el-form-item label="充值（元）"><el-input-number v-model="amountYuan" :min="1" :step="100" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rechargeOpen = false">取消</el-button>
        <el-button type="success" @click="recharge">确认收款</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="txnOpen" :title="`${current?.name} 的流水`" width="560px">
      <el-table :data="txns" size="small" max-height="360">
        <el-table-column label="类型" width="80">
          <template #default="{ row }">
            <el-tag :type="row.type === 'recharge' ? 'success' : 'warning'">{{ row.type === "recharge" ? "充值" : "消费" }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="金额" width="100">
          <template #default="{ row }">{{ row.type === "recharge" ? "+" : "-" }}¥{{ money(row.amount_cents) }}</template>
        </el-table-column>
        <el-table-column label="余额" width="100">
          <template #default="{ row }">¥{{ money(row.balance_after) }}</template>
        </el-table-column>
        <el-table-column prop="note" label="备注" />
        <el-table-column label="时间" width="160">
          <template #default="{ row }">{{ new Date(row.created_at * 1000).toLocaleString() }}</template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { ElMessage } from "element-plus";
import { api, money } from "../api";

const list = ref<any[]>([]);
const txns = ref<any[]>([]);
const createOpen = ref(false);
const rechargeOpen = ref(false);
const txnOpen = ref(false);
const current = ref<any>(null);
const form = ref<any>({});
const amountYuan = ref(100);

async function load() {
  list.value = (await api.get("/members")).data;
}
onMounted(load);

function openCreate() {
  form.value = { name: "", phone: "", discount_percent: 100 };
  createOpen.value = true;
}
async function create() {
  try {
    await api.post("/members", form.value);
    ElMessage.success("已添加");
    createOpen.value = false;
    load();
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error ?? "保存失败");
  }
}
function openRecharge(row: any) {
  current.value = row;
  amountYuan.value = 100;
  rechargeOpen.value = true;
}
async function recharge() {
  try {
    await api.post(`/members/${current.value.id}/recharge`, { amount_cents: Math.round(amountYuan.value * 100) });
    ElMessage.success("充值成功（模拟网关已收款）");
    rechargeOpen.value = false;
    load();
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error ?? "充值失败");
  }
}
async function showTxns(row: any) {
  current.value = row;
  txns.value = (await api.get(`/members/${row.id}/transactions`)).data;
  txnOpen.value = true;
}
</script>

<style scoped>
.toolbar { display: flex; justify-content: space-between; margin-bottom: 14px; }
.hint { margin-left: 10px; color: #909399; }
</style>
