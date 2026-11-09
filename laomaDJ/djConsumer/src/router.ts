import {createRouter, createWebHashHistory} from 'vue-router'
import Login from './components/Login.vue'
import SelectAddress from './components/SelectAddress.vue'

// 使用 hash 路由，地址是 /#/ 和 /#/selectAddress。刷新页面不需要后端配回退。
const routes = [
  { path: '/', component: Login },
  { path: '/selectAddress', component: SelectAddress },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router
