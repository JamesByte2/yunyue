<template>
  <view class="page">
    <view class="query-card">
      <input class="input" v-model="phone" type="number" maxlength="11" placeholder="输入预约时填写的手机号" />
      <button class="query-btn" @click="load">查询我的预约</button>
    </view>

    <view v-for="b in list" :key="b.id" class="card">
      <view class="row">
        <text class="service">{{ b.service_name }}</text>
        <text class="status" :class="b.status">{{ STATUS[b.status] }}</text>
      </view>
      <view class="meta">📅 {{ b.book_date }} · {{ fmtSlot(b.start_min) }}</view>
      <button
        v-if="b.status === 'pending'"
        class="cancel"
        size="mini"
        @click="cancel(b)"
      >
        取消预约
      </button>
    </view>

    <view v-if="loaded && !list.length" class="empty">没有查到预约记录</view>
  </view>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { apiGet, apiPost, toast } from "../../api";

const STATUS: Record<string, string> = {
  pending: "待确认",
  confirmed: "已确认",
  done: "已完成",
  canceled: "已取消",
};

const phone = ref(uni.getStorageSync("yy_phone") || "");
const list = ref<any[]>([]);
const loaded = ref(false);

function fmtSlot(min: number) {
  return `${String(Math.floor(min / 60)).padStart(2, "0")}:${String(min % 60).padStart(2, "0")}`;
}

async function load() {
  if (phone.value.length !== 11) {
    toast("请输入 11 位手机号");
    return;
  }
  uni.setStorageSync("yy_phone", phone.value);
  list.value = await apiGet(`/api/public/mybookings?phone=${phone.value}`);
  loaded.value = true;
}

async function cancel(b: any) {
  uni.showModal({
    title: "取消预约",
    content: "确定取消这条预约吗？",
    success: async (res) => {
      if (!res.confirm) return;
      try {
        await apiPost(`/api/public/mybookings/${b.id}/cancel`, { phone: phone.value });
        toast("已取消");
        load();
      } catch (e: any) {
        toast(e.message);
      }
    },
  });
}
</script>

<style>
.page { padding: 12px; background: #f5f6f8; min-height: 100vh; }
.query-card { background: #fff; border-radius: 12px; padding: 14px; margin-bottom: 12px; }
.input { border: 1px solid #e4e7ed; border-radius: 8px; padding: 10px; margin-bottom: 10px; }
.query-btn { background: #2b85e4; color: #fff; border-radius: 8px; }
.card { background: #fff; border-radius: 12px; padding: 14px; margin-bottom: 12px; }
.row { display: flex; justify-content: space-between; align-items: center; }
.service { font-size: 15px; font-weight: 600; }
.status.pending { color: #e6a23c; }
.status.confirmed { color: #2b85e4; }
.status.done { color: #67c23a; }
.status.canceled { color: #909399; }
.meta { color: #606266; font-size: 13px; margin: 8px 0; }
.cancel { background: #fff; color: #f56c6c; border: 1px solid #fbc4c4; }
.empty { text-align: center; color: #909399; padding: 40px 0; }
</style>
