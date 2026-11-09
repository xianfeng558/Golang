<template>
  <van-nav-bar title="呼叫代驾" />


  <div class="map-page-container">
    <el-amap :zoom="zoom" :center="center" @init="init" />
  </div>

  <!-- 可以使用 CellGroup 作为容器 -->
  <van-cell-group inset>
    <van-field v-model="f" label="出发地" placeholder="左岸公社停车场">
      <template #button>
        <van-button size="small" type="primary">更改</van-button>
      </template>
    </van-field>
    <van-field v-model="t" label="目的地" placeholder="三里屯太古里" />
  </van-cell-group>

  <div style="margin: 16px;">
    <van-button block type="primary" @click="readyCall">
      准备呼叫
    </van-button>
  </div>
</template>
    
<script lang="ts" setup>
import { ref } from "vue";
const zoom = ref(18);
// 地图中心是示例起点附近，只用于显示，不参与询价。
const center = ref([116.304413, 39.984376]);
// f、t 是地点名称，当前询价没有把它们解析成坐标。
const f = ref("左岸公社停车场");
const t = ref("三里屯太古里");
let map: any;

const init = (e) => {
  const marker = new AMap.Marker({
    position: [116.304413, 39.984376]
  });
  e.add(marker);
  map = e;
}

import { estimate } from '@/api/selectAddress'
import { showConfirmDialog, showDialog } from 'vant'

// 询价固定使用这两对「经度,纬度」。改输入框里的地点名称不会改变请求。
const f1 = '116.304388,39.984405'
const t1 = '116.454406,39.934862'

// 先询价再弹出确认框。price 单位是分，除以 100 后显示成元。
// 未登录或乘客服务没启动时，请求失败，提示先登录或确认服务已启动。
async function readyCall() {
  try {
    const resp = await estimate(f1, t1)
    if (Number(resp.code) !== 0) {
      showDialog({ message: resp.message || '询价失败' })
      return
    }
    const yuan = (Number(resp.price) / 100).toFixed(2)
    showConfirmDialog({
      title: '费用预估',
      confirmButtonText: '确认呼叫',
      message: `${yuan}元`,
    }).then(() => {
      showDialog({ message: '确认呼叫' })
    }).catch(() => {})
  } catch {
    showDialog({ message: '请先登录，或确认乘客服务已启动' })
  }
}



</script>
    
<style>
.map-page-container {
  height: 450px;
}
</style>
    