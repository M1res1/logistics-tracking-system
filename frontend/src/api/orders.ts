import client from './client'

export interface OrderItem {
  menu_item_id: number
  quantity: number
  unit_price: number
}

export interface Order {
  id: number
  restaurant_id: number
  status: string
  total: number
  delivery_address: string
  lat: number
  lng: number
  items: OrderItem[]
  created_at: string
}

export interface CreateOrderData {
  restaurant_id: number
  items: OrderItem[]
  delivery_address: string
  lat: number
  lng: number
}

export const createOrder = (data: CreateOrderData) =>
  client.post<{ data: Order }>('/orders', data)

export const myOrders = () =>
  client.get<{ data: { orders: Order[]; total: number; page: number; limit: number } }>('/orders/my')

export const getOrder = (id: number) =>
  client.get<{ data: Order }>(`/orders/${id}`)

export const cancelOrder = (id: number) =>
  client.post(`/orders/${id}/cancel`)
