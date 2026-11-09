import './assets/main.css'

import { createApp } from 'vue'
import App from './App.vue'
const app = createApp(App);

// router
import router from "./router"
app.use(router)


// Vant
// 1. 引入你需要的组件
import { Button, NavBar,  Field, CellGroup, Checkbox } from 'vant'
// 2. 引入组件样式
import 'vant/lib/index.css'
// 3. 注册你需要的组件
app
.use(Button)
.use(NavBar)
.use(Field)
.use(CellGroup)
.use(Checkbox)


// AMap
import VueAMap, {initAMapApiLoader} from '@vuemap/vue-amap'
import '@vuemap/vue-amap/dist/style.css'
initAMapApiLoader({
    key: 'f2d9051cac9e32664e1fdb95f64045b2',
    securityJsCode: '21419c7ec0cf97cfcef60c928e4f9c5c', // 新版key需要配合安全密钥使用
    //Loca:{
    //  version: '2.0.0'
    //} // 如果需要使用loca组件库，需要加载Loca
  })
app.use(VueAMap)



app.mount('#app')
