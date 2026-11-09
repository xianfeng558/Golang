import axios from 'axios'
import { apiCustomer } from './base'

export async function estimate(f?: any, t?: any) {
    try {
        const response = await axios.get(`${apiCustomer}/customer/get-verify-code/${f}/${t}`);
        return response.data
    } catch (error) {
        return error
    }
}