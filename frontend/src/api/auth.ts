import client from './client'

export interface LoginResp { access_token: string; refresh_token: string }
export interface User { id: number; email: string; phone: string; user_type: string }

export const login = (email: string, password: string) =>
  client.post<LoginResp>('/auth/login', { email, password })

export const register = (data: { email: string; password: string; phone: string; user_type: string }) =>
  client.post<LoginResp>('/auth/register', data)

export const logout = () => client.post('/auth/logout')

export const me = () => client.get<{ data: User }>('/auth/me')
