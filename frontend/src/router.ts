import { createRouter, createWebHashHistory } from "vue-router";
import Login from "./views/Login.vue";
import Layout from "./views/Layout.vue";
import Dashboard from "./views/Dashboard.vue";
import Bookings from "./views/Bookings.vue";
import Services from "./views/Services.vue";
import Staff from "./views/Staff.vue";
import Members from "./views/Members.vue";
import Transactions from "./views/Transactions.vue";

export const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: "/login", component: Login },
    {
      path: "/",
      component: Layout,
      children: [
        { path: "", component: Dashboard },
        { path: "bookings", component: Bookings },
        { path: "services", component: Services },
        { path: "staff", component: Staff },
        { path: "members", component: Members },
        { path: "transactions", component: Transactions },
      ],
    },
  ],
});
