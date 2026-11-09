<template>
  <van-nav-bar title="老马代驾" left-arrow @click-left="onClickLeft" @click-right="onClickRight" />

  <van-cell-group inset>
    <div>欢迎登录老马代驾</div>
  </van-cell-group>


  <!-- 可以使用 CellGroup 作为容器 -->
  <van-cell-group inset>
    <van-field v-model="phoneNumber" center clearable label="" placeholder="手机号码">
      <template #button>
        <van-button size="small" type="primary" @click="verifyCode">发送验证码</van-button>
      </template>
    </van-field>
    <van-field v-model="code" label="" placeholder="请输入验证码" />
  </van-cell-group>

  <div style="margin: 16px;">
    <van-button block type="primary" @click="login">
      登录
    </van-button>
  </div>

  <div style="margin: 16px;">

    <van-checkbox v-model="agree" style="font-size: small;">我已阅读并同意《软件使用协议》《隐私政策》</van-checkbox>
  </div>
</template>
    
<script lang="ts" setup>
import { ref } from 'vue'

const phoneNumber = ref('')
const code = ref('')
const agree = ref(false)

import { getVerifyCode, checkLogin } from '@/api/login'
async function verifyCode() {
  getVerifyCode(phoneNumber.value)
}

async function login() {
  let resp = checkLogin({
    telephone: phoneNumber.value,
    verify_code: code.value,
  })

  console.log(resp)
}




</script>
    
<style>
.map-page-container {
  height: 500px;
}
</style>
    