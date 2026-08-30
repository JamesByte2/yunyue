<template>
  <el-card>
    <div class="toolbar">
      <span>技师管理</span>
      <el-button type="primary" @click="openCreate">新增技师</el-button>
    </div>
    <el-table :data="list" stripe>
      <el-table-column prop="id" label="#" width="60" />
      <el-table-column prop="name" label="姓名" width="120" />
      <el-table-column prop="title" label="职位" width="160" />
      <el-table-column prop="phone" label="电话" width="150" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.active ? 'success' : 'info'">{{ row.active ? "在岗" : "停用" }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作">
        <template #default="{ row }">
          <el-button size="small" @click="openEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" plain @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="open" :title="form.id ? '编辑技师' : '新增技师'" width="420px">
      <el-form label-width="80px">
        <el-form-item label="姓名"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="职位"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="电话"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="在岗"><el-switch v-model="form.active" /></el-form-item>
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
import { api } from "../api";

const list = ref<any[]>([]);
const open = ref(false);
const form = ref<any>({});

async function load() {
  list.value = (await api.get("/staff")).data;
}
onMounted(load);

function openCreate() {
  form.value = { name: "", title: "", phone: "", active: true };
  open.value = true;
}
function openEdit(row: any) {
  form.value = { ...row };
  open.value = true;
}
async function save() {
  try {
    if (form.value.id) await api.put(`/staff/${form.value.id}`, form.value);
    else await api.post("/staff", form.value);
    ElMessage.success("已保存");
    open.value = false;
    load();
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error ?? "保存失败");
  }
}
async function remove(row: any) {
  const r = await api.delete(`/staff/${row.id}`);
  if (r.data.deactivated) ElMessage.warning("有历史预约，已改为停用");
  else ElMessage.success("已删除");
  load();
}
</script>

<style scoped>
.toolbar { display: flex; justify-content: space-between; margin-bottom: 14px; }
</style>
