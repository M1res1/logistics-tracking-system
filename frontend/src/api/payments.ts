import client from './client'

export interface Wallet {
  id: number
  user_id: number
  balance: number
}

export interface Payment {
  id: number
  order_id: number
  user_id: number
  amount: number
  method: string
  status: string
  created_at: string
}

export const getWallet = (userId: number) =>
  client.get<{ data: Wallet }>(`/wallet/${userId}`)

export const topupWallet = (userId: number, amount: number) =>
  client.post(`/wallet/${userId}/topup`, { amount })

export const processPayment = (data: {
  order_id: number
  user_id: number
  amount: number
  method: string
  idempotency_key: string
}) => client.post<{ data: Payment }>('/payments/process', data)

export const getPayment = (id: number) =>
  client.get<{ data: Payment }>(`/payments/${id}`)
