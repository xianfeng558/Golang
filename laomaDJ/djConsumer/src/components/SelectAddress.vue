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
const center = ref([116.304413, 39.984376]);

const init = (e) => {
  const marker = new AMap.Marker({
    position: [116.304413, 39.984376]
  });
  e.add(marker);
  map = e;
}

let f1 = '116.304388,39.984405'
let t1 = '116.454406,39.934862'
import { showConfirmDialog, showDialog } from 'vant';

function readyCall() {
  showConfirmDialog({
    title: '费用预估',
    confirmButtonText: '确认呼叫',
    message:
      '11.75元',
  })
    .then(() => {
      showDialog({
        message: '确认呼叫',
      }).then(() => {
        // on close
      });
    })
    .catch(() => {});
}



</script>
    
<style>
.map-page-container {
  height: 450px;
}
</style>
    