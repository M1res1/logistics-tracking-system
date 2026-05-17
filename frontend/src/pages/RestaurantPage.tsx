import { useState, useEffect, lazy, Suspense } from 'react'
import { useAuth } from '../contexts/AuthContext'
import * as restaurantsApi from '../api/restaurants'
import type { Restaurant, MenuItem, RestaurantOrder } from '../api/restaurants'

const LocationPicker = lazy(() => import('../components/LocationPicker'))

type Tab = 'restaurant' | 'menu' | 'orders'

function statusColor(s: string) {
  switch (s?.toUpperCase()) {
    case 'PENDING':  return 'badge-pending'
    case 'CONFIRMED': case 'PREPARING': return 'badge-active'
    case 'READY':    return 'badge-info'
    case 'REJECTED': case 'CANCELLED': return 'badge-danger'
    default: return 'badge-neutral'
  }
}

export default function RestaurantPage() {
  const { user, logout } = useAuth()
  const [tab, setTab] = useState<Tab>('restaurant')
  const [restaurant, setRestaurant] = useState<Restaurant | null>(null)
  const [restLoading, setRestLoading] = useState(true)
  const [editing, setEditing] = useState(false)
  const [form, setForm] = useState({ name: '', address: '', lat: 0, lng: 0, cuisine_types: '', opening_time: '09:00', closing_time: '22:00' })
  const [formErr, setFormErr] = useState('')
  const [formOk, setFormOk] = useState('')
  const [formBusy, setFormBusy] = useState(false)
  const [menuItems, setMenuItems] = useState<MenuItem[]>([])
  const [menuLoading, setMenuLoading] = useState(false)
  const [menuForm, setMenuForm] = useState({ name: '', description: '', category: '', price: '', prep_time_minutes: '' })
  const [menuErr, setMenuErr] = useState('')
  const [menuBusy, setMenuBusy] = useState(false)
  const [deletingItem, setDeletingItem] = useState<number | null>(null)
  const [orders, setOrders] = useState<RestaurantOrder[]>([])
  const [ordersLoading, setOrdersLoading] = useState(false)
  const [actingOrder, setActingOrder] = useState<number | null>(null)
  const [orderMsg, setOrderMsg] = useState('')

  useEffect(() => {
    restaurantsApi.listRestaurants()
      .then((r) => {
        const list = r.data.data?.restaurants || []
        setRestaurant(list.find((x) => x.owner_id === user?.id) || null)
      })
      .catch(() => {})
      .finally(() => setRestLoading(false))
  }, [user?.id])

  useEffect(() => {
    if (tab === 'menu' && restaurant) loadMenu()
    if (tab === 'orders' && restaurant) loadOrders()
  }, [tab, restaurant])

  const loadMenu = () => {
    if (!restaurant) return
    setMenuLoading(true)
    restaurantsApi.getMenu(restaurant.id)
      .then((r) => setMenuItems(r.data.data || []))
      .catch(() => {})
      .finally(() => setMenuLoading(false))
  }

  const loadOrders = () => {
    if (!restaurant) return
    setOrdersLoading(true)
    restaurantsApi.listOrders(restaurant.id)
      .then((r) => setOrders(Array.isArray(r.data.data) ? r.data.data : []))
      .catch(() => {})
      .finally(() => setOrdersLoading(false))
  }

  const startEdit = () => {
    if (!restaurant) return
    setForm({ name: restaurant.name, address: restaurant.address, lat: restaurant.lat, lng: restaurant.lng, cuisine_types: restaurant.cuisine_types || '', opening_time: restaurant.opening_time, closing_time: restaurant.closing_time })
    setEditing(true); setFormErr(''); setFormOk('')
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault(); setFormErr(''); setFormBusy(true)
    try {
      const r = await restaurantsApi.createRestaurant(form)
      setRestaurant(r.data.data); setFormOk('Restaurant created!')
    } catch { setFormErr('Failed to create.') }
    finally { setFormBusy(false) }
  }

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault(); setFormErr(''); setFormBusy(true)
    try {
      const r = await restaurantsApi.updateRestaurant(restaurant!.id, form)
      setRestaurant(r.data.data); setFormOk('Saved!'); setEditing(false)
    } catch { setFormErr('Failed to save.') }
    finally { setFormBusy(false) }
  }

  const handleToggle = async () => {
    if (!restaurant) return
    try {
      await restaurantsApi.toggleRestaurant(restaurant.id)
      setRestaurant((p) => p ? { ...p, is_active: !p.is_active } : p)
    } catch {}
  }

  const handleAddItem = async (e: React.FormEvent) => {
    e.preventDefault(); if (!restaurant) return
    setMenuErr(''); setMenuBusy(true)
    try {
      await restaurantsApi.addMenuItem(restaurant.id, { ...menuForm, price: parseFloat(menuForm.price), prep_time_minutes: parseInt(menuForm.prep_time_minutes) })
      setMenuForm({ name: '', description: '', category: '', price: '', prep_time_minutes: '' })
      loadMenu()
    } catch { setMenuErr('Failed to add item.') }
    finally { setMenuBusy(false) }
  }

  const handleDeleteItem = async (id: number) => {
    if (!restaurant) return
    setDeletingItem(id)
    try { await restaurantsApi.deleteMenuItem(restaurant.id, id); loadMenu() }
    catch {} finally { setDeletingItem(null) }
  }

  const handleOrderAction = async (orderId: number, fn: (rId: number, oId: number) => Promise<unknown>, label: string) => {
    if (!restaurant) return
    setActingOrder(orderId); setOrderMsg('')
    try { await fn(restaurant.id, orderId); setOrderMsg(`Order #${orderId} ${label}.`); loadOrders() }
    catch { setOrderMsg(`Failed: ${label}`) }
    finally { setActingOrder(null) }
  }

  return (
    <>
      <nav className="navbar">
        <span className="navbar-brand">FoodApp</span>
        <div className="navbar-links">
          <span className="navbar-user">🍽️ {user?.email}</span>
          <button onClick={logout}>Sign out</button>
        </div>
      </nav>

      <div className="container">
        <div className="page-header">
          <div className="page-title">Restaurant Dashboard</div>
          <div className="page-subtitle">Manage your restaurant, menu, and orders</div>
        </div>

        <div className="tabs">
          <button className={`tab ${tab === 'restaurant' ? 'active' : ''}`} onClick={() => setTab('restaurant')}>🏪 My Restaurant</button>
          <button className={`tab ${tab === 'menu' ? 'active' : ''}`} onClick={() => setTab('menu')} disabled={!restaurant}>🍴 Menu</button>
          <button className={`tab ${tab === 'orders' ? 'active' : ''}`} onClick={() => setTab('orders')} disabled={!restaurant}>📋 Orders</button>
        </div>

        {/* ── Restaurant tab ── */}
        {tab === 'restaurant' && (
          restLoading ? <div className="loading-state"><span className="spinner" /></div>
          : !restaurant && !editing ? (
            <div className="card">
              <div className="card-title">✨ Create Your Restaurant</div>
              {formErr && <div className="alert alert-error">{formErr}</div>}
              {formOk && <div className="alert alert-success">{formOk}</div>}
              <form onSubmit={handleCreate}><RestaurantForm form={form} setForm={setForm} /><button type="submit" className="btn btn-primary" disabled={formBusy}>{formBusy ? 'Creating...' : 'Create Restaurant'}</button></form>
            </div>
          ) : restaurant && !editing ? (
            <div className="card">
              <div className="flex-between mb-4">
                <div>
                  <div style={{ fontSize: '1.5rem', fontWeight: 800 }}>{restaurant.name}</div>
                  <div className="text-muted">📍 {restaurant.address}</div>
                </div>
                <span className={`badge ${restaurant.is_active ? 'badge-active' : 'badge-danger'}`} style={{ fontSize: '0.85rem', padding: '0.4rem 0.8rem' }}>
                  {restaurant.is_active ? '● Open' : '● Closed'}
                </span>
              </div>
              <div className="stat-grid">
                {restaurant.cuisine_types && <div className="stat-box"><div className="stat-value">🍴</div><div className="stat-label">{restaurant.cuisine_types}</div></div>}
                <div className="stat-box"><div className="stat-value">🕐</div><div className="stat-label">{restaurant.opening_time}–{restaurant.closing_time}</div></div>
                {restaurant.rating ? <div className="stat-box"><div className="stat-value">★ {restaurant.rating}</div><div className="stat-label">Rating</div></div> : null}
              </div>
              {formOk && <div className="alert alert-success">{formOk}</div>}
              <div className="flex-gap">
                <button className="btn btn-secondary" onClick={startEdit}>✏️ Edit</button>
                <button className={`btn ${restaurant.is_active ? 'btn-danger' : 'btn-success'}`} onClick={handleToggle}>
                  {restaurant.is_active ? '🔴 Close' : '🟢 Open'}
                </button>
              </div>
            </div>
          ) : editing ? (
            <div className="card">
              <div className="card-title">✏️ Edit Restaurant</div>
              {formErr && <div className="alert alert-error">{formErr}</div>}
              <form onSubmit={handleUpdate}>
                <RestaurantForm form={form} setForm={setForm} />
                <div className="flex-gap">
                  <button type="button" className="btn btn-secondary" onClick={() => setEditing(false)}>Cancel</button>
                  <button type="submit" className="btn btn-primary" disabled={formBusy}>{formBusy ? 'Saving...' : 'Save Changes'}</button>
                </div>
              </form>
            </div>
          ) : null
        )}

        {/* ── Menu tab ── */}
        {tab === 'menu' && restaurant && (
          <>
            <div className="card">
              <div className="card-title">➕ Add Menu Item</div>
              {menuErr && <div className="alert alert-error">{menuErr}</div>}
              <form onSubmit={handleAddItem}>
                <div className="grid-2">
                  <div className="form-group"><label>Name</label><input className="input" value={menuForm.name} onChange={(e) => setMenuForm((p) => ({ ...p, name: e.target.value }))} required /></div>
                  <div className="form-group"><label>Category</label><input className="input" value={menuForm.category} onChange={(e) => setMenuForm((p) => ({ ...p, category: e.target.value }))} placeholder="Mains, Drinks..." required /></div>
                </div>
                <div className="form-group"><label>Description</label><input className="input" value={menuForm.description} onChange={(e) => setMenuForm((p) => ({ ...p, description: e.target.value }))} /></div>
                <div className="grid-2">
                  <div className="form-group"><label>Price ($)</label><input className="input" type="number" step="0.01" min="0" value={menuForm.price} onChange={(e) => setMenuForm((p) => ({ ...p, price: e.target.value }))} required /></div>
                  <div className="form-group"><label>Prep Time (min)</label><input className="input" type="number" min="0" value={menuForm.prep_time_minutes} onChange={(e) => setMenuForm((p) => ({ ...p, prep_time_minutes: e.target.value }))} required /></div>
                </div>
                <button type="submit" className="btn btn-primary btn-sm" disabled={menuBusy}>{menuBusy ? 'Adding...' : '+ Add Item'}</button>
              </form>
            </div>

            <div className="card">
              <div className="card-title">Menu ({menuItems.length} items)</div>
              {menuLoading ? <div className="loading-state"><span className="spinner" /></div>
                : menuItems.length === 0
                  ? <div className="empty-state"><div className="empty-state-icon">🍽️</div><div className="empty-state-text">No items yet</div></div>
                  : menuItems.map((item) => (
                    <div key={item.id} className="menu-item-row">
                      <div className="menu-item-info">
                        <div className="menu-item-name">{item.name}</div>
                        {item.description && <div className="menu-item-desc">{item.description}</div>}
                        <div className="menu-item-meta"><span className="badge badge-neutral">{item.category}</span><span>⏱ {item.prep_time_minutes} min</span></div>
                      </div>
                      <div className="flex-gap">
                        <span className="menu-item-price">${item.price.toFixed(2)}</span>
                        <button className="btn btn-danger btn-sm" disabled={deletingItem === item.id} onClick={() => handleDeleteItem(item.id)}>
                          {deletingItem === item.id ? '...' : 'Delete'}
                        </button>
                      </div>
                    </div>
                  ))
              }
            </div>
          </>
        )}

        {/* ── Orders tab ── */}
        {tab === 'orders' && restaurant && (
          <>
            <div className="flex-between mb-4">
              <span className="text-muted">{orders.length} order(s)</span>
              <button className="btn btn-secondary btn-sm" onClick={loadOrders}>↻ Refresh</button>
            </div>
            {orderMsg && <div className="alert alert-info">{orderMsg}</div>}
            {ordersLoading ? <div className="loading-state"><span className="spinner" /></div>
              : orders.length === 0
                ? <div className="empty-state"><div className="empty-state-icon">📋</div><div className="empty-state-text">No orders yet</div></div>
                : orders.map((order) => (
                  <div key={order.id} className="card order-card">
                    <div className="order-row">
                      <div className="order-info">
                        <div className="flex-gap mb-1">
                          <span className="order-id">Order #{order.id}</span>
                          <span className={`badge ${statusColor(order.status)}`}>{order.status}</span>
                        </div>
                        {order.delivery_address && <div className="text-sm text-muted">📍 {order.delivery_address}</div>}
                        {order.total != null && <div className="order-total">${order.total.toFixed(2)}</div>}
                      </div>
                      <div className="order-actions">
                        {order.status === 'PENDING' && <>
                          <button className="btn btn-success btn-sm" disabled={actingOrder === order.id} onClick={() => handleOrderAction(order.id, restaurantsApi.acceptOrder, 'accepted')}>Accept</button>
                          <button className="btn btn-danger btn-sm" disabled={actingOrder === order.id} onClick={() => handleOrderAction(order.id, restaurantsApi.rejectOrder, 'rejected')}>Reject</button>
                        </>}
                        {order.status === 'CONFIRMED' && (
                          <button className="btn btn-primary btn-sm" disabled={actingOrder === order.id} onClick={() => handleOrderAction(order.id, restaurantsApi.readyOrder, 'marked ready')}>Mark Ready</button>
                        )}
                      </div>
                    </div>
                  </div>
                ))
            }
          </>
        )}
      </div>
    </>
  )
}

function RestaurantForm({ form, setForm }: {
  form: { name: string; address: string; lat: number; lng: number; cuisine_types: string; opening_time: string; closing_time: string }
  setForm: React.Dispatch<React.SetStateAction<typeof form>>
}) {
  const set = (k: string, v: string | number) => setForm((p) => ({ ...p, [k]: v }))
  return (
    <>
      <div className="form-group"><label>Restaurant Name</label><input name="name" className="input" value={form.name} onChange={(e) => set('name', e.target.value)} required /></div>
      <div className="form-group"><label>Address</label><input name="address" className="input" value={form.address} onChange={(e) => set('address', e.target.value)} required /></div>
      <div className="form-group">
        <label>📍 Location on map</label>
        <Suspense fallback={<div className="loading-state"><span className="spinner" /></div>}>
          <LocationPicker lat={form.lat} lng={form.lng} onChange={(la, ln, addr) => {
            set('lat', la); set('lng', ln)
            if (addr && !form.address) set('address', addr.split(',').slice(0, 3).join(','))
          }} />
        </Suspense>
      </div>
      <div className="form-group"><label>Cuisine Types</label><input name="cuisine_types" className="input" value={form.cuisine_types} onChange={(e) => set('cuisine_types', e.target.value)} placeholder="e.g. Pizza, Italian" /></div>
      <div className="grid-2">
        <div className="form-group"><label>Opens</label><input name="opening_time" className="input" type="time" value={form.opening_time} onChange={(e) => set('opening_time', e.target.value)} /></div>
        <div className="form-group"><label>Closes</label><input name="closing_time" className="input" type="time" value={form.closing_time} onChange={(e) => set('closing_time', e.target.value)} /></div>
      </div>
    </>
  )
}
