import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import * as restaurantsApi from '../api/restaurants'
import type { Restaurant, MenuItem, RestaurantOrder } from '../api/restaurants'

type Tab = 'restaurant' | 'menu' | 'orders'

function statusBadgeClass(status: string) {
  switch (status?.toUpperCase()) {
    case 'PENDING': return 'badge badge-pending'
    case 'CONFIRMED':
    case 'READY': return 'badge badge-active'
    case 'REJECTED':
    case 'CANCELLED': return 'badge badge-danger'
    default: return 'badge badge-neutral'
  }
}

export default function RestaurantPage() {
  const { user, logout } = useAuth()
  const [tab, setTab] = useState<Tab>('restaurant')
  const [restaurant, setRestaurant] = useState<Restaurant | null>(null)
  const [restLoading, setRestLoading] = useState(true)

  // Create / Edit form
  const [editing, setEditing] = useState(false)
  const [form, setForm] = useState({
    name: '',
    address: '',
    lat: '',
    lng: '',
    cuisine_types: '',
    opening_time: '09:00',
    closing_time: '22:00',
  })
  const [formError, setFormError] = useState('')
  const [formLoading, setFormLoading] = useState(false)
  const [formSuccess, setFormSuccess] = useState('')

  // Menu
  const [menuItems, setMenuItems] = useState<MenuItem[]>([])
  const [menuLoading, setMenuLoading] = useState(false)
  const [menuForm, setMenuForm] = useState({ name: '', description: '', category: '', price: '', prep_time_minutes: '' })
  const [menuFormError, setMenuFormError] = useState('')
  const [menuFormLoading, setMenuFormLoading] = useState(false)
  const [deletingItem, setDeletingItem] = useState<number | null>(null)

  // Orders
  const [orders, setOrders] = useState<RestaurantOrder[]>([])
  const [ordersLoading, setOrdersLoading] = useState(false)
  const [actingOrder, setActingOrder] = useState<number | null>(null)
  const [orderMsg, setOrderMsg] = useState('')

  // Load restaurant on mount — filter by owner_id
  useEffect(() => {
    restaurantsApi.listRestaurants()
      .then((res) => {
        const list = res.data.data?.restaurants || []
        const owned = list.find((r) => r.owner_id === user?.id) || null
        setRestaurant(owned)
      })
      .catch(() => {})
      .finally(() => setRestLoading(false))
  }, [user?.id])

  // Load menu when tab changes
  useEffect(() => {
    if (tab === 'menu' && restaurant) {
      loadMenu()
    }
    if (tab === 'orders' && restaurant) {
      loadOrders()
    }
  }, [tab, restaurant])

  const loadMenu = () => {
    if (!restaurant) return
    setMenuLoading(true)
    restaurantsApi.getMenu(restaurant.id)
      .then((res) => setMenuItems(res.data.data || []))
      .catch(() => {})
      .finally(() => setMenuLoading(false))
  }

  const loadOrders = () => {
    if (!restaurant) return
    setOrdersLoading(true)
    restaurantsApi.listOrders(restaurant.id)
      .then((res) => setOrders(res.data.data || []))
      .catch(() => {})
      .finally(() => setOrdersLoading(false))
  }

  const handleFormChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm((prev) => ({ ...prev, [e.target.name]: e.target.value }))
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormError('')
    setFormLoading(true)
    try {
      const res = await restaurantsApi.createRestaurant({
        name: form.name,
        address: form.address,
        lat: parseFloat(form.lat) || 0,
        lng: parseFloat(form.lng) || 0,
        cuisine_types: form.cuisine_types,
        opening_time: form.opening_time,
        closing_time: form.closing_time,
      })
      setRestaurant(res.data.data)
      setFormSuccess('Restaurant created!')
    } catch {
      setFormError('Failed to create restaurant.')
    } finally {
      setFormLoading(false)
    }
  }

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!restaurant) return
    setFormError('')
    setFormLoading(true)
    try {
      const res = await restaurantsApi.updateRestaurant(restaurant.id, {
        name: form.name,
        address: form.address,
        lat: parseFloat(form.lat) || 0,
        lng: parseFloat(form.lng) || 0,
        cuisine_types: form.cuisine_types,
        opening_time: form.opening_time,
        closing_time: form.closing_time,
      })
      setRestaurant(res.data.data)
      setFormSuccess('Restaurant updated!')
      setEditing(false)
    } catch {
      setFormError('Failed to update restaurant.')
    } finally {
      setFormLoading(false)
    }
  }

  const startEdit = () => {
    if (!restaurant) return
    setForm({
      name: restaurant.name,
      address: restaurant.address,
      lat: String(restaurant.lat),
      lng: String(restaurant.lng),
      cuisine_types: restaurant.cuisine_types || '',
      opening_time: restaurant.opening_time,
      closing_time: restaurant.closing_time,
    })
    setEditing(true)
    setFormError('')
    setFormSuccess('')
  }

  const handleToggle = async () => {
    if (!restaurant) return
    try {
      await restaurantsApi.toggleRestaurant(restaurant.id)
      setRestaurant((prev) => prev ? { ...prev, is_active: !prev.is_active } : prev)
    } catch {
      // ignore
    }
  }

  const handleAddMenuItem = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!restaurant) return
    setMenuFormError('')
    setMenuFormLoading(true)
    try {
      await restaurantsApi.addMenuItem(restaurant.id, {
        name: menuForm.name,
        description: menuForm.description,
        category: menuForm.category,
        price: parseFloat(menuForm.price) || 0,
        prep_time_minutes: parseInt(menuForm.prep_time_minutes) || 0,
      })
      setMenuForm({ name: '', description: '', category: '', price: '', prep_time_minutes: '' })
      loadMenu()
    } catch {
      setMenuFormError('Failed to add menu item.')
    } finally {
      setMenuFormLoading(false)
    }
  }

  const handleDeleteItem = async (itemId: number) => {
    if (!restaurant) return
    setDeletingItem(itemId)
    try {
      await restaurantsApi.deleteMenuItem(restaurant.id, itemId)
      loadMenu()
    } catch {
      // ignore
    } finally {
      setDeletingItem(null)
    }
  }

  const handleOrderAction = async (
    orderId: number,
    action: (restId: number, ordId: number) => Promise<unknown>,
    label: string
  ) => {
    if (!restaurant) return
    setActingOrder(orderId)
    setOrderMsg('')
    try {
      await action(restaurant.id, orderId)
      setOrderMsg(`Order #${orderId} ${label} successfully.`)
      loadOrders()
    } catch {
      setOrderMsg(`Failed to ${label.toLowerCase()} order #${orderId}.`)
    } finally {
      setActingOrder(null)
    }
  }

  return (
    <>
      <nav className="navbar">
        <span className="navbar-brand">FoodApp | Restaurant</span>
        <div className="navbar-links">
          <span className="text-sm" style={{ color: '#93c5fd' }}>{user?.email}</span>
          <button onClick={logout}>Logout</button>
        </div>
      </nav>

      <div className="container">
        <div className="page-header">
          <div className="page-title">Restaurant Dashboard</div>
        </div>

        <div className="tabs">
          <button className={`tab ${tab === 'restaurant' ? 'active' : ''}`} onClick={() => setTab('restaurant')}>
            My Restaurant
          </button>
          <button className={`tab ${tab === 'menu' ? 'active' : ''}`} onClick={() => setTab('menu')} disabled={!restaurant}>
            Menu
          </button>
          <button className={`tab ${tab === 'orders' ? 'active' : ''}`} onClick={() => setTab('orders')} disabled={!restaurant}>
            Orders
          </button>
        </div>

        {/* My Restaurant Tab */}
        {tab === 'restaurant' && (
          <>
            {restLoading ? (
              <div className="text-muted">Loading...</div>
            ) : !restaurant && !editing ? (
              <div className="card">
                <div style={{ fontWeight: 600, marginBottom: '1rem' }}>Create Your Restaurant</div>
                {formError && <div className="error-msg">{formError}</div>}
                {formSuccess && <div className="status-bar success">{formSuccess}</div>}
                <form onSubmit={handleCreate}>
                  <RestaurantForm form={form} onChange={handleFormChange} />
                  <button type="submit" className="btn btn-primary" disabled={formLoading}>
                    {formLoading ? 'Creating...' : 'Create Restaurant'}
                  </button>
                </form>
              </div>
            ) : restaurant && !editing ? (
              <div className="card">
                <div className="flex-between mb-2">
                  <h2 style={{ fontSize: '1.25rem', fontWeight: 700 }}>{restaurant.name}</h2>
                  <span className={restaurant.is_active ? 'badge badge-active' : 'badge badge-danger'}>
                    {restaurant.is_active ? 'Open' : 'Closed'}
                  </span>
                </div>
                <div className="text-sm text-muted mb-1">{restaurant.address}</div>
                {restaurant.cuisine_types && (
                  <div className="text-sm text-muted mb-1">Cuisines: {restaurant.cuisine_types}</div>
                )}
                <div className="text-sm text-muted mb-3">
                  Hours: {restaurant.opening_time} – {restaurant.closing_time}
                </div>
                <div className="flex-gap">
                  <button className="btn btn-primary btn-sm" onClick={startEdit}>Edit Details</button>
                  <button
                    className={`btn btn-sm ${restaurant.is_active ? 'btn-danger' : 'btn-success'}`}
                    onClick={handleToggle}
                  >
                    {restaurant.is_active ? 'Close Restaurant' : 'Open Restaurant'}
                  </button>
                </div>
              </div>
            ) : editing ? (
              <div className="card">
                <div style={{ fontWeight: 600, marginBottom: '1rem' }}>Edit Restaurant</div>
                {formError && <div className="error-msg">{formError}</div>}
                {formSuccess && <div className="status-bar success">{formSuccess}</div>}
                <form onSubmit={handleUpdate}>
                  <RestaurantForm form={form} onChange={handleFormChange} />
                  <div className="flex-gap">
                    <button type="button" className="btn btn-secondary" onClick={() => setEditing(false)}>Cancel</button>
                    <button type="submit" className="btn btn-primary" disabled={formLoading}>
                      {formLoading ? 'Saving...' : 'Save Changes'}
                    </button>
                  </div>
                </form>
              </div>
            ) : null}
          </>
        )}

        {/* Menu Tab */}
        {tab === 'menu' && restaurant && (
          <>
            <div className="card">
              <div style={{ fontWeight: 600, marginBottom: '1rem' }}>Add Menu Item</div>
              {menuFormError && <div className="error-msg">{menuFormError}</div>}
              <form onSubmit={handleAddMenuItem}>
                <div className="grid-2">
                  <div className="form-group">
                    <label>Name</label>
                    <input className="input" value={menuForm.name} onChange={(e) => setMenuForm((p) => ({ ...p, name: e.target.value }))} required />
                  </div>
                  <div className="form-group">
                    <label>Category</label>
                    <input className="input" value={menuForm.category} onChange={(e) => setMenuForm((p) => ({ ...p, category: e.target.value }))} placeholder="e.g. Mains" required />
                  </div>
                </div>
                <div className="form-group">
                  <label>Description</label>
                  <input className="input" value={menuForm.description} onChange={(e) => setMenuForm((p) => ({ ...p, description: e.target.value }))} />
                </div>
                <div className="grid-2">
                  <div className="form-group">
                    <label>Price ($)</label>
                    <input className="input" type="number" step="0.01" min="0" value={menuForm.price} onChange={(e) => setMenuForm((p) => ({ ...p, price: e.target.value }))} required />
                  </div>
                  <div className="form-group">
                    <label>Prep Time (min)</label>
                    <input className="input" type="number" min="0" value={menuForm.prep_time_minutes} onChange={(e) => setMenuForm((p) => ({ ...p, prep_time_minutes: e.target.value }))} required />
                  </div>
                </div>
                <button type="submit" className="btn btn-primary btn-sm" disabled={menuFormLoading}>
                  {menuFormLoading ? 'Adding...' : 'Add Item'}
                </button>
              </form>
            </div>

            <div className="card">
              <div style={{ fontWeight: 600, marginBottom: '1rem' }}>Menu Items ({menuItems.length})</div>
              {menuLoading ? (
                <div className="text-muted">Loading...</div>
              ) : menuItems.length === 0 ? (
                <div className="text-muted">No items yet. Add your first menu item above.</div>
              ) : (
                menuItems.map((item) => (
                  <div key={item.id} className="menu-item-row">
                    <div className="menu-item-info">
                      <div className="menu-item-name">{item.name}</div>
                      <div className="menu-item-desc">{item.description}</div>
                      <div className="menu-item-meta">
                        <span className="badge badge-neutral">{item.category}</span>
                        <span>{item.prep_time_minutes} min</span>
                      </div>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                      <span className="menu-item-price">${item.price.toFixed(2)}</span>
                      <button
                        className="btn btn-danger btn-sm"
                        disabled={deletingItem === item.id}
                        onClick={() => handleDeleteItem(item.id)}
                      >
                        {deletingItem === item.id ? '...' : 'Delete'}
                      </button>
                    </div>
                  </div>
                ))
              )}
            </div>
          </>
        )}

        {/* Orders Tab */}
        {tab === 'orders' && restaurant && (
          <>
            <div className="flex-between mb-4">
              <span className="text-muted">{orders.length} order(s)</span>
              <button className="btn btn-secondary btn-sm" onClick={loadOrders}>Refresh</button>
            </div>

            {orderMsg && <div className="status-bar" style={{ marginBottom: '1rem' }}>{orderMsg}</div>}

            {ordersLoading ? (
              <div className="text-muted">Loading orders...</div>
            ) : orders.length === 0 ? (
              <div className="card"><div className="text-muted">No orders yet.</div></div>
            ) : (
              orders.map((order) => (
                <div key={order.id} className="card">
                  <div className="order-row">
                    <div className="order-info">
                      <div className="flex-gap mb-1">
                        <strong>Order #{order.id}</strong>
                        <span className={statusBadgeClass(order.status)}>{order.status}</span>
                      </div>
                      {order.delivery_address && (
                        <div className="text-sm text-muted">{order.delivery_address}</div>
                      )}
                      {order.total != null && (
                        <div className="text-sm mt-1"><strong>${order.total?.toFixed(2)}</strong></div>
                      )}
                    </div>
                    <div className="order-actions">
                      {order.status === 'PENDING' && (
                        <>
                          <button
                            className="btn btn-success btn-sm"
                            disabled={actingOrder === order.id}
                            onClick={() => handleOrderAction(order.id, restaurantsApi.acceptOrder, 'accepted')}
                          >
                            Accept
                          </button>
                          <button
                            className="btn btn-danger btn-sm"
                            disabled={actingOrder === order.id}
                            onClick={() => handleOrderAction(order.id, restaurantsApi.rejectOrder, 'rejected')}
                          >
                            Reject
                          </button>
                        </>
                      )}
                      {order.status === 'CONFIRMED' && (
                        <button
                          className="btn btn-primary btn-sm"
                          disabled={actingOrder === order.id}
                          onClick={() => handleOrderAction(order.id, restaurantsApi.readyOrder, 'marked ready')}
                        >
                          Mark Ready
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              ))
            )}
          </>
        )}
      </div>
    </>
  )
}

interface RestaurantFormProps {
  form: {
    name: string
    address: string
    lat: string
    lng: string
    cuisine_types: string
    opening_time: string
    closing_time: string
  }
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void
}

function RestaurantForm({ form, onChange }: RestaurantFormProps) {
  return (
    <>
      <div className="form-group">
        <label>Restaurant Name</label>
        <input name="name" className="input" value={form.name} onChange={onChange} required />
      </div>
      <div className="form-group">
        <label>Address</label>
        <input name="address" className="input" value={form.address} onChange={onChange} required />
      </div>
      <div className="grid-2">
        <div className="form-group">
          <label>Latitude</label>
          <input name="lat" className="input" value={form.lat} onChange={onChange} placeholder="0.000000" />
        </div>
        <div className="form-group">
          <label>Longitude</label>
          <input name="lng" className="input" value={form.lng} onChange={onChange} placeholder="0.000000" />
        </div>
      </div>
      <div className="form-group">
        <label>Cuisine Types (comma-separated)</label>
        <input name="cuisine_types" className="input" value={form.cuisine_types} onChange={onChange} placeholder="Italian, Pizza, Pasta" />
      </div>
      <div className="grid-2">
        <div className="form-group">
          <label>Opening Time</label>
          <input name="opening_time" className="input" type="time" value={form.opening_time} onChange={onChange} />
        </div>
        <div className="form-group">
          <label>Closing Time</label>
          <input name="closing_time" className="input" type="time" value={form.closing_time} onChange={onChange} />
        </div>
      </div>
    </>
  )
}
