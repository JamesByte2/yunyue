<template>
  <el-container style="min-height: 100vh">
    <el-aside width="200px" class="aside">
      <div class="brand">云约 · 门店预约</div>
      <el-menu :default-active="route.path" router>
        <el-menu-item index="/">工作台</el-menu-item>
        <el-menu-item index="/bookings">预约管理</el-menu-item>
        <el-menu-item index="/services">服务项目</el-menu-item>
        <el-menu-item index="/staff">技师管理</el-menu-item>
        <el-menu-item v-if="isAdmin" index="/members">会员管理</el-menu-item>
        <el-menu-item v-if="isAdmin" index="/transactions">交易流水</el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <span>{{ user?.name }}（{{ user?.role === "admin" ? "管理员" : "技师" }}）</span>
        <el-button text @click="logout">退出</el-button>
      </el-header>
      <el-main><router-view /></el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { api } from "../api";

const route = useRoute();
const router = useRouter();
const user = JSON.parse(localStorage.getItem("yy_user") ?? "null");
const isAdmin = computed(() => user?.role === "admin");

function logout() {
  api.post; // keep import used
  localStorage.removeItem("yy_token");
  localStorage.removeItem("yy_user");
  router.push("/login");
}
</script>

<style scoped>
.aside { border-right: 1px solid #e4e7ed; }
.brand { font-weight: 700; padding: 18px 20px; font-size: 15px; }
.header { display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #e4e7ed; }
</style>
