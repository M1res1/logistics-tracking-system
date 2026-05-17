import client from './client'

export interface LocationData {
  lat: number
  lng: number
  updated_at?: string
}

export interface Driver {
  id: number
  email: string
  phone: string
}

export const assignDelivery = (orderId: number, lat: number, lng: number) =>
  client.post('/deliveries/assign', { order_id: orderId, lat, lng })

export const acceptDelivery = (id: number) =>
  client.post(`/deliveries/${id}/accept`)

export const pickupDelivery = (id: number) =>
  client.post(`/deliveries/${id}/pickup`)

export const completeDelivery = (id: number) =>
  client.post(`/deliveries/${id}/complete`)

export const updateLocation = (id: number, lat: number, lng: number) =>
  client.put(`/deliveries/${id}/location`, { lat, lng })

export const getLocation = (id: number) =>
  client.get<{ data: LocationData }>(`/deliveries/${id}/location`)

export const availableDrivers = () =>
  client.get<{ data: Driver[] }>('/drivers/available')
