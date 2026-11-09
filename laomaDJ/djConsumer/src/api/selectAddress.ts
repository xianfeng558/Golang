import axios from 'axios'
import { apiCustomer } from './base'

// estimate 向乘客服务询价。origin、destination 必须是「经度,纬度」，不是地点名称。
// 价格在响应的 price 里，单位是分，JSON 里可能是字符串。没有令牌时乘客服务返回 401。
export async function estimate(origin: string, destination: string) {
  const token = localStorage.getItem('customerToken') || ''
  const response = await axios.get(
    `${apiCustomer}/customer/estimate-price/${encodeURIComponent(origin)}/${encodeURIComponent(destination)}`,
    { headers: { Authorization: `Bearer ${token}` } },
  )
  return response.data
}
