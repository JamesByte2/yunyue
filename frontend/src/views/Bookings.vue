<template>
  <div>
    <el-card>
      <div class="toolbar">
        <el-date-picker v-model="date" type="date" value-format="YYYY-MM-DD" placeholder="按日期筛选" clearable style="width: 160px" />
        <el-select v-model="status" placeholder="按状态筛选" clearable style="width: 140px">
          <el-option v-for="(v, k) in BOOKING_STATUS" :key="k" :label="v.label" :value="k" />
        </el-select>
        <el-button type="primary" @click="openCreate">新建预约</el-button>
      </div>

      <el-table :data="list" stripe>
        <el-table-column prop="id" label="#" width="60" />
        <el-table-column prop="book_date" label="日期" width="110" />
        <el-table-column label="时间" width="90">
          <template #default="{ row }">{{ minToTime(row.start_min) }}</template>
        </el-table-column>
        <el-table-column prop="customer_name" label="客户" width="110" />
        <el-table-column prop="customer_phone" label="电话" width="130" />
        <el-table-column prop="service_name" label="服务" width="110" />
        <el-table-column prop="staff_name" label="技师" width="90" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="BOOKING_STATUS[row.status]?.type">{{ BOOKING_STATUS[row.status]?.label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="120" />
        <el-table-column label="操作" width="240">
          <template #default="{ row }">
            <el-button v-if="row.status === 'pending'" size="small" type="primary" @click="setStatus(row, 'confirmed')">确认</el-button>
            <el-button v-if="row.status === 'confirmed'" size="small" type="success" @click="openFinish(row)">完成结算</el-button>
            <el-button v-if="row.status === 'pending' || row.status === 'confirmed'" size="small" type="danger" plain @click="setStatus(row, 'canceled')">取消</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="createOpen" title="新建预约" width="480px">
      <el-form label-width="90px">
        <el-form-item label="客户姓名"><el-input v-model="form.customer_name" /></el-form-item>
        <el-form-item label="电话"><el-input v-model="form.customer_phone" /></el-form-item>
        <el-form-item label="服务项目">
          <el-select v-model="form.service_item_id" style="width: 100%">
            <el-option v-for="s in services" :key="s.id" :label="`${s.name}（${s.duration_min}分钟 / ¥${money(s.price_cents)}）`" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="技师">
          <el-select v-model="form.staff_id" style="width: 100%">
            <el-option v-for="s in staffList" :key="s.id" :label="`${s.name} · ${s.title}`" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="日期"><el-date-picker v-model="form.book_date" type="date" value-format="YYYY-MM-DD" /></el-form-item>
        <el-form-item label="开始时间">
          <el-time-select v-model="startHHmm" start="09:00" end="20:30" step="00:30" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createOpen = false">取消</el-button>
        <el-button type="primary" @click="create">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="finishOpen" title="完成结算" width="420px">
      <p>服务金额：¥{{ money(currentServicePrice) }}</p>
      <el-select v-model="finishMemberId" clearable placeholder="选择会员卡支付（留空 = 现金）" style="width: 100%">
        <el-option v-for="m in members" :key="m.id" :label="`${m.name} ${m.phone} · 余额¥${money(m.balance_cents)} · ${m.discount_percent / 10}折`" :value="m.id" />
      </el-select>
      <template #footer>
        <el-button @click="finishOpen = false">取消</el-button>
        <el-button type="success" @click="finish">确认完成</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import { api, money, minToTime, BOOKING_STATUS } from "../api";

const list = ref<any[]>([]);
const services = ref<any[]>([]);
const staffList = ref<any[]>([]);
const members = ref<any[]>([]);
const date = ref("");
const status = ref("");
const createOpen = ref(false);
const finishOpen = ref(false);
const finishMemberId = ref<number | undefined>();
const current = ref<any>(null);
const today = new Date().toISOString().slice(0, 10);

const form = ref<any>({ customer_name: "", customer_phone: "", service_item_id: null, staff_id: null, book_date: today, remark: "" });
const startHHmm = ref("10:00");
const currentServicePrice = ref(0);

async function load() {
  const params: any = {};
  if (date.value) params.date = date.value;
  if (status.value) params.status = status.value;
  list.value = (await api.get("/bookings", { params })).data;
}

async function loadRefs() {
  services.value = (await api.get("/services")).data;
  staffList.value = (await api.get("/staff")).data;
  try {
    members.value = (await api.get("/members")).data;
  } catch {
    members.value = []; // 技师角色无权查看会员列表
  }
}

onMounted(async () => {
  await Promise.all([load(), loadRefs()]);
});

watch([date, status], load);
watch(() => form.value.service_item_id, (id) => {
  const s = services.value.find((x) => x.id === id);
  currentServicePrice.value = s?.price_cents ?? 0;
});

function openCreate() {
  form.value = { customer_name: "", customer_phone: "", service_item_id: null, staff_id: null, book_date: today, remark: "" };
  createOpen.value = true;
}

async function create() {
  const [h, m] = startHHmm.value.split(":").map(Number);
  try {
    await api.post("/bookings", { ...form.value, start_min: h * 60 + m });
    ElMessage.success("预约已创建");
    createOpen.value = false;
    load();
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error ?? "创建失败");
  }
}

async function setStatus(row: any, st: string) {
  try {
    await api.patch(`/bookings/${row.id}/status`, { status: st });
    ElMessage.success("已更新");
    load();
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error ?? "操作失败");
  }
}

function openFinish(row: any) {
  current.value = row;
  finishMemberId.value = undefined;
  const s = services.value.find((x) => x.id === row.service_item_id);
  currentServicePrice.value = s?.price_cents ?? 0;
  finishOpen.value = true;
}

async function finish() {
  try {
    const body: any = {};
    if (finishMemberId.value) body.member_id = finishMemberId.value;
    await api.post(`/bookings/${current.value.id}/finish`, body);
    ElMessage.success("已完成结算");
    finishOpen.value = false;
    load();
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error ?? "结算失败");
  }
}
</script>

<style scoped>
.toolbar { display: flex; gap: 12px; margin-bottom: 14px; }
</style>
