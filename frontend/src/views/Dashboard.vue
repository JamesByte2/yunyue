<template>
  <div>
    <el-row :gutter="16">
      <el-col :span="6"><el-card><el-statistic title="今日预约" :value="s.today_bookings" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="待确认" :value="s.pending_count" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="本月营收（元）" :value="s.month_revenue / 100" :precision="2" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="会员数" :value="s.member_count" /></el-card></el-col>
    </el-row>

    <el-card style="margin-top: 16px">
      <template #header>今日预约（{{ today }}）</template>
      <el-table :data="s.today_list" stripe>
        <el-table-column prop="start_min" label="时间" width="100">
          <template #default="{ row }">{{ minToTime(row.start_min) }}</template>
        </el-table-column>
        <el-table-column prop="customer_name" label="客户" width="120" />
        <el-table-column prop="customer_phone" label="电话" width="140" />
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="BOOKING_STATUS[row.status]?.type">{{ BOOKING_STATUS[row.status]?.label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { api, minToTime, BOOKING_STATUS } from "../api";

const s = ref<any>({ today_list: [] });
const today = new Date().toISOString().slice(0, 10);

onMounted(async () => {
  s.value = (await api.get("/dashboard")).data;
});
</script>
