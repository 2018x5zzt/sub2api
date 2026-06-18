import { xlabClient } from '../xlabClient'
import type { PaymentOrder } from '../payment'

export async function getMyOrders(params: Record<string, unknown> = {}) {
  const { data } = await xlabClient.get('/payment/orders/my', { params })
  return data
}

export async function getOrder(id: number | string): Promise<PaymentOrder> {
  const { data } = await xlabClient.get<PaymentOrder>(`/payment/orders/${id}`)
  return data
}

export const xlabPaymentAPI = { getMyOrders, getOrder }
