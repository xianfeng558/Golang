import axios from 'axios'
import qs from 'qs'
import { apiCustomer } from './base'

export async function getVerifyCode(phoneNumber?: any) {
    try {
        const response = await axios.get(`${apiCustomer}/customer/get-verify-code/${phoneNumber}`);
        return response.data
    } catch (error) {
        return error
    }
}

export async function checkLogin(params?: any) {
    try {
        axios.defaults.headers.post['Content-Type'] = 'application/x-www-form-urlencoded'
        const response = await axios.post(`${apiCustomer}/customer/login`, qs.stringify(params));
        return response.data
    } catch (error) {
        return error
    }
}