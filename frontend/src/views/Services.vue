<template>
  <el-card>
    <div class="toolbar">
      <span>服务项目（预约的基本单位）</span>
      <el-button type="primary" @click="openCreate">新增服务</el-button>
    </div>
    <el-table :data="list" stripe>
      <el-table-column prop="id" label="#" width="60" />
      <el-table-column prop="name" label="名称" width="160" />
      <el-table-column prop="duration_min" label="时长（分钟）" width="120" />
      <el-table-column label="价格" width="120">
        <template #default="{ row }">¥{{ money(row.price_cents) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.active ? 'success' : 'info'">{{ row.active ? "上架" : "停用" }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作">
        <template #default="{ row }">
          <el-button size="small" @click="openEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" plain @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="open" :title="form.id ? '编辑服务' : '新增服务'" width="420px">
      <el-form label-width="100px">
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="时长（分钟）"><el-input-number v-model="form.duration_min" :min="5" :step="5" /></el-form-item>
        <el-form-item label="价格（元）"><el-input-number v-model="priceYuan" :min="0" :precision="2" :step="10" /></el-form-item>
        <el-form-item label="上架"><el-switch v-model="form.active" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="open = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { ElMessage } from "element-plus";
import { api, money } from "../api";

const list = ref<any[]>([]);
const open = ref(false);
const form = ref<any>({});
const priceYuan = ref(0);

async function load() {
  list.value = (await api.get("/services")).data;
}
onMounted(load);

function openCreate() {
  form.value = { name: "", duration_min: 30, active: true };
  priceYuan.value = 0;
  open.value = true;
}
function openEdit(row: any) {
  form.value = { ...row };
  priceYuan.value = row.price_cents / 100;
  open.value = true;
}
async function save() {
  const body = { ...form.value, price_cents: Math.round(priceYuan.value * 100) };
  try {
    if (body.id) await api.put(`/services/${body.id}`, body);
    else await api.post("/services", body);
    ElMessage.success("已保存");
    open.value = false;
    load();
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error ?? "保存失败");
  }
}
async function remove(row: any) {
  const r = await api.delete(`/services/${row.id}`);
  if (r.data.deactivated) ElMessage.warning("有历史预约，已改为停用");
  else ElMessage.success("已删除");
  load();
}
</script>

<style scoped>
.toolbar { display: flex; justify-content: space-between; margin-bottom: 14px; }
</style>
