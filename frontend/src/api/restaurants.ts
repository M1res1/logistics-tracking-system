import client from './client'

export interface Restaurant {
  id: number
  owner_id: number
  name: string
  address: string
  lat: number
  lng: number
  cuisine_types: string   // stored as string in backend
  opening_time: string
  closing_time: string
  is_active: boolean
  rating?: number
}

export interface MenuItem {
  id: number
  name: string
  description: string
  category: string
  price: number
  prep_time_minutes: number
}

export interface RestaurantOrder {
  id: number
  status: string
  total?: number
  items?: { menu_item_id: number; quantity: number; unit_price: number }[]
  delivery_address?: string
  created_at?: string
}

export const listRestaurants = (params?: { lat?: number; lng?: number; radius?: number }) =>
  client.get<{ data: { restaurants: Restaurant[]; total: number } }>('/restaurants', { params })

export const getRestaurant = (id: number) =>
  client.get<{ data: Restaurant }>(`/restaurants/${id}`)

export const getMenu = (id: number) =>
  client.get<{ data: MenuItem[] }>(`/restaurants/${id}/menu`)

export const createRestaurant = (data: {
  name: string
  address: string
  lat: number
  lng: number
  cuisine_types: string
  opening_time: string
  closing_time: string
}) => client.post<{ data: Restaurant }>('/restaurants', data)

export const updateRestaurant = (id: number, data: Partial<{
  name: string
  address: string
  lat: number
  lng: number
  cuisine_types: string
  opening_time: string
  closing_time: string
}>) => client.put<{ data: Restaurant }>(`/restaurants/${id}`, data)

export const toggleRestaurant = (id: number) =>
  client.put(`/restaurants/${id}/toggle`)

export const addMenuItem = (restaurantId: number, data: {
  name: string
  description: string
  category: string
  price: number
  prep_time_minutes: number
}) => client.post<{ data: MenuItem }>(`/restaurants/${restaurantId}/menu-items`, data)

export const updateMenuItem = (restaurantId: number, itemId: number, data: Partial<{
  name: string
  description: string
  category: string
  price: number
  prep_time_minutes: number
}>) => client.put<{ data: MenuItem }>(`/restaurants/${restaurantId}/menu-items/${itemId}`, data)

export const deleteMenuItem = (restaurantId: number, itemId: number) =>
  client.delete(`/restaurants/${restaurantId}/menu-items/${itemId}`)

export const listOrders = (restaurantId: number) =>
  client.get<{ data: RestaurantOrder[] }>(`/restaurants/${restaurantId}/orders`)

export const acceptOrder = (restaurantId: number, orderId: number) =>
  client.post(`/restaurants/${restaurantId}/orders/${orderId}/accept`)

export const readyOrder = (restaurantId: number, orderId: number) =>
  client.post(`/restaurants/${restaurantId}/orders/${orderId}/ready`)

export const rejectOrder = (restaurantId: number, orderId: number) =>
  client.post(`/restaurants/${restaurantId}/orders/${orderId}/reject`)
