import axios from 'axios'
import { apiCustomer } from './base'

// getVerifyCode 请求乘客服务发码。响应里的 code 为 0 才表示成功。
// 验证码在 verifyCode 或 verify_code 字段，页面会把它填进输入框。没有短信。
export async function getVerifyCode(phoneNumber: string) {
  const response = await axios.get(`${apiCustomer}/customer/get-verify-code/${phoneNumber}`)
  return response.data
}

// checkLogin 用手机号和验证码换 JWT。
// 请求体字段是 verify_code。成功时 token 要存进 localStorage 的 customerToken，后面询价放在 Authorization 里。
export async function checkLogin(telephone: string, verifyCode: string) {
  const response = await axios.post(`${apiCustomer}/customer/login`, {
    telephone,
    verify_code: verifyCode,
  })
  return response.data
}
