<template>
  <div class="wrap">
    <el-card class="card">
      <h2>云约 · 门店预约管理</h2>
      <el-form label-position="top" @submit.prevent="submit">
        <el-form-item label="邮箱">
          <el-input v-model="email" placeholder="admin@yunyue.cn" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="password" type="password" show-password placeholder="admin12345" />
        </el-form-item>
        <el-button type="primary" style="width: 100%" :loading="loading" native-type="submit">
          登 录
        </el-button>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { api } from "../api";

const router = useRouter();
const email = ref("admin@yunyue.cn");
const password = ref("admin12345");
const loading = ref(false);

async function submit() {
  loading.value = true;
  try {
    const r = await api.post("/auth/login", { email: email.value, password: password.value });
    localStorage.setItem("yy_token", r.data.token);
    localStorage.setItem("yy_user", JSON.stringify({ name: r.data.name, role: r.data.role }));
    router.push("/");
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error ?? "登录失败");
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.wrap { min-height: 100vh; display: grid; place-items: center; background: #f2f3f5; }
.card { width: 380px; }
h2 { text-align: center; margin: 0 0 20px; }
</style>
