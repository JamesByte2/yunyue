<template>
  <view class="page">
    <view class="card">
      <view class="step-title">1 · 选择服务</view>
      <view
        v-for="s in services"
        :key="s.id"
        class="service-item"
        :class="{ active: form.service_item_id === s.id }"
        @click="chooseService(s)"
      >
        <view class="service-name">{{ s.name }}</view>
        <view class="service-meta">{{ s.duration_min }} 分钟 · ¥{{ (s.price_cents / 100).toFixed(2) }}</view>
      </view>
    </view>

    <view class="card" v-if="form.service_item_id">
      <view class="step-title">2 · 选择技师</view>
      <view class="staff-row">
        <view
          v-for="t in staff"
          :key="t.id"
          class="staff-chip"
          :class="{ active: form.staff_id === t.id }"
          @click="chooseStaff(t)"
        >
          {{ t.name }} · {{ t.title }}
        </view>
      </view>

      <view class="step-title">3 · 选择日期与时间</view>
      <picker mode="date" :value="form.book_date" :start="today" @change="onDate">
        <view class="date-picker">📅 {{ form.book_date }}（点击选择）</view>
      </picker>
      <view class="slot-grid" v-if="slots.length">
        <view
          v-for="slot in slots"
          :key="slot.start_min"
          class="slot"
          :class="{ active: form.start_min === slot.start_min, disabled: !slot.available }"
          @click="slot.available && (form.start_min = slot.start_min)"
        >
          {{ fmtSlot(slot.start_min) }}
        </view>
      </view>
      <view v-if="slotsLoading" class="hint">正在查询可约时段……</view>
    </view>

    <view class="card" v-if="form.start_min !== null">
      <view class="step-title">4 · 填写联系信息</view>
      <input class="input" v-model="form.customer_name" placeholder="您的姓名" />
      <input class="input" v-model="form.customer_phone" type="number" maxlength="11" placeholder="手机号（用于查单/取消）" />
      <button class="submit" :loading="submitting" @click="submit">提交预约</button>
    </view>

    <view class="footer-link" @click="goOrders">查看我的预约 →</view>
  </view>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { apiGet, apiPost, toast } from "../../api";

const services = ref<any[]>([]);
const staff = ref<any[]>([]);
const slots = ref<any[]>([]);
const slotsLoading = ref(false);
const submitting = ref(false);
const today = new Date().toISOString().slice(0, 10);
const form = ref<any>({
  service_item_id: null,
  staff_id: null,
  book_date: today,
  start_min: null,
  customer_name: "",
  customer_phone: "",
});

onMounted(async () => {
  const data = await apiGet("/api/public/catalog");
  services.value = data.services;
  staff.value = data.staff;
});

function chooseService(s: any) {
  form.value.service_item_id = s.id;
  form.value.start_min = null;
}
function chooseStaff(t: any) {
  form.value.staff_id = t.id;
  form.value.start_min = null;
}
function onDate(e: any) {
  form.value.book_date = e.detail.value;
  form.value.start_min = null;
}

async function loadSlots() {
  if (!form.value.service_item_id || !form.value.staff_id) return;
  slotsLoading.value = true;
  try {
    const data = await apiGet(
      `/api/public/availability?date=${form.value.book_date}&staff_id=${form.value.staff_id}&service_id=${form.value.service_item_id}`,
    );
    slots.value = data.slots;
  } finally {
    slotsLoading.value = false;
  }
}
watch(() => [form.value.book_date, form.value.staff_id, form.value.service_item_id], loadSlots);

function fmtSlot(min: number) {
  return `${String(Math.floor(min / 60)).padStart(2, "0")}:${String(min % 60).padStart(2, "0")}`;
}

async function submit() {
  if (!form.value.customer_name || form.value.customer_phone.length !== 11) {
    toast("请填写姓名和 11 位手机号");
    return;
  }
  submitting.value = true;
  try {
    await apiPost("/api/public/bookings", form.value);
    toast("预约成功！可在「我的预约」中查看");
    uni.switchTab({ url: "/pages/orders/orders" });
  } catch (e: any) {
    toast(e.message ?? "预约失败");
  } finally {
    submitting.value = false;
  }
}

function goOrders() {
  uni.switchTab({ url: "/pages/orders/orders" });
}
</script>

<style>
.page { padding: 12px; background: #f5f6f8; min-height: 100vh; }
.card { background: #fff; border-radius: 12px; padding: 14px; margin-bottom: 12px; }
.step-title { font-weight: 600; margin-bottom: 10px; }
.service-item { padding: 10px; border: 1px solid #e4e7ed; border-radius: 8px; margin-bottom: 8px; }
.service-item.active { border-color: #2b85e4; background: #ecf5ff; }
.service-name { font-size: 15px; }
.service-meta { font-size: 12px; color: #909399; margin-top: 2px; }
.staff-row { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; }
.staff-chip { padding: 6px 12px; border: 1px solid #e4e7ed; border-radius: 20px; font-size: 13px; }
.staff-chip.active { border-color: #2b85e4; background: #ecf5ff; color: #2b85e4; }
.date-picker { padding: 10px; background: #f5f6f8; border-radius: 8px; margin-bottom: 10px; }
.slot-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.slot { width: 64px; text-align: center; padding: 6px 0; border: 1px solid #e4e7ed; border-radius: 6px; font-size: 13px; }
.slot.active { border-color: #2b85e4; background: #2b85e4; color: #fff; }
.slot.disabled { color: #c0c4cc; background: #f5f6f8; text-decoration: line-through; }
.input { border: 1px solid #e4e7ed; border-radius: 8px; padding: 10px; margin-bottom: 10px; }
.submit { background: #2b85e4; color: #fff; border-radius: 8px; }
.footer-link { text-align: center; color: #2b85e4; padding: 10px; font-size: 14px; }
.hint { color: #909399; font-size: 12px; }
</style>
