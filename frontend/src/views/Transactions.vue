<template>
  <el-card>
    <template #header>交易流水（最近 100 条）</template>
    <el-table :data="list" stripe>
      <el-table-column prop="id" label="#" width="60" />
      <el-table-column label="类型" width="90">
        <template #default="{ row }">
          <el-tag :type="row.type === 'recharge' ? 'success' : 'warning'">{{ row.type === "recharge" ? "充值" : "消费" }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="金额" width="110">
        <template #default="{ row }">{{ row.type === "recharge" ? "+" : "-" }}¥{{ money(row.amount_cents) }}</template>
      </el-table-column>
      <el-table-column label="变动后余额" width="120">
        <template #default="{ row }">¥{{ money(row.balance_after) }}</template>
      </el-table-column>
      <el-table-column prop="member_id" label="会员ID" width="90" />
      <el-table-column prop="note" label="备注" min-width="160" />
      <el-table-column label="时间" width="170">
        <template #default="{ row }">{{ new Date(row.created_at * 1000).toLocaleString() }}</template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { api, money } from "../api";

const list = ref<any[]>([]);
onMounted(async () => {
  list.value = (await api.get("/transactions")).data;
});
</script>
