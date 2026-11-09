<template>
  <van-nav-bar title="老马代驾" />

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

  <div style="margin: 16px; color: #666;">{{ hint }}</div>

  <div style="margin: 16px;">

    <van-checkbox v-model="agree" style="font-size: small;">我已阅读并同意《软件使用协议》《隐私政策》</van-checkbox>
  </div>
</template>
    
<script lang="ts" setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { getVerifyCode, checkLogin } from '@/api/login'

const phoneNumber = ref('13800000000')
const code = ref('')
// 未勾选协议时不允许登录。
const agree = ref(false)
const hint = ref('')
const router = useRouter()

// 向乘客服务要 6 位验证码。业务 code 是字符串，所以用 Number 比较。
// 成功后把码填进输入框。60 秒后 Redis 里的键过期，需要重新获取。
async function verifyCode() {
  hint.value = ''
  try {
    const resp = await getVerifyCode(phoneNumber.value)
    if (Number(resp.code) !== 0) {
      hint.value = resp.message || '验证码获取失败'
      return
    }
    code.value = resp.verifyCode || resp.verify_code || ''
    hint.value = '验证码已填入，60 秒内有效'
  } catch {
    hint.value = '验证码服务没有响应'
  }
}

// 登录成功后把 JWT 存到 customerToken，再进入选址页。令牌大约 60 天，退出接口会让它立刻失效。
async function login() {
  if (!agree.value) {
    hint.value = '请先勾选协议'
    return
  }
  hint.value = ''
  try {
    const resp = await checkLogin(phoneNumber.value, code.value)
    if (Number(resp.code) !== 0) {
      hint.value = resp.message || '登录失败'
      return
    }
    localStorage.setItem('customerToken', resp.token)
    router.push('/selectAddress')
  } catch {
    hint.value = '登录请求失败'
  }
}
</script>
    
<style>
.map-page-container {
  height: 500px;
}
</style>
    