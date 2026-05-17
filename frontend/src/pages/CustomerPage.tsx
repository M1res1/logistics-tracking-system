import { useState, useEffect, useCallback } from 'react'
import { useAuth } from '../contexts/AuthContext'
import * as restaurantsApi from '../api/restaurants'
import * as ordersApi from '../api/orders'
import * as deliveryApi from '../api/delivery'
import type { Restaurant, MenuItem } from '../api/restaurants'
import type { Order } from '../api/orders'
import type { LocationData } from '../api/delivery'

type Tab = 'restaurants' | 'orders' | 'track'

function statusBadgeClass(status: string) {
  switch (status?.toUpperCase()) {
    case 'PENDING': return 'badge badge-pending'
    case 'CONFIRMED':
    case 'READY':
    case 'PICKED_UP': return 'badge badge-active'
    case 'DELIVERED': return 'badge badge-info'
    case 'CANCELLED':
    case 'REJECTED': return 'badge badge-danger'
    default: return 'badge badge-neutral'
  }
}

interface CartItem {
  menu_item_id: number
  name: string
  unit_price: number
  quantity: number
}

interface MenuModalProps {
  restaurant: Restaurant
  onClose: () => void
}

function MenuModal({ restaurant, onClose }: MenuModalProps) {
  const [menu, setMenu] = useState<MenuItem[]>([])
  const [cart, setCart] = useState<CartItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [step, setStep] = useState<'menu' | 'order'>('menu')
  const [deliveryAddress, setDeliveryAddress] = useState('')
  const [lat, setLat] = useState('')
  const [lng, setLng] = useState('')
  const [ordering, setOrdering] = useState(false)
  const [orderSuccess, setOrderSuccess] = useState('')

  useEffect(() => {
    restaurantsApi.getMenu(restaurant.id)
      .then((res) => setMenu(res.data.data || []))
      .catch(() => setError('Failed to load menu'))
      .finally(() => setLoading(false))
  }, [restaurant.id])

  const setQty = (item: MenuItem, qty: number) => {
    if (qty <= 0) {
      setCart((prev) => prev.filter((c) => c.menu_item_id !== item.id))
    } else {
      setCart((prev) => {
        const existing = prev.find((c) => c.menu_item_id === item.id)
        if (existing) {
          return prev.map((c) => c.menu_item_id === item.id ? { ...c, quantity: qty } : c)
        }
        return [...prev, { menu_item_id: item.id, name: item.name, unit_price: item.price, quantity: qty }]
      })
    }
  }

  const getQty = (itemId: number) => cart.find((c) => c.menu_item_id === itemId)?.quantity ?? 0

  const total = cart.reduce((sum, c) => sum + c.unit_price * c.quantity, 0)

  const handleOrder = async (e: React.FormEvent) => {
    e.preventDefault()
    if (cart.length === 0) { setError('Add at least one item to your cart'); return }
    setOrdering(true)
    setError('')
    try {
      await ordersApi.createOrder({
        restaurant_id: restaurant.id,
        items: cart.map(({ menu_item_id, quantity, unit_price }) => ({ menu_item_id, quantity, unit_price })),
        delivery_address: deliveryAddress,
        lat: parseFloat(lat) || 0,
        lng: parseFloat(lng) || 0,
      })
      setOrderSuccess('Order placed successfully!')
      setCart([])
    } catch {
      setError('Failed to place order. Please try again.')
    } finally {
      setOrdering(false)
    }
  }

  // Group menu items by category
  const categories = [...new Set(menu.map((m) => m.category))]

  return (
    <div className="modal-overlay" onClick={(e) => e.target === e.currentTarget && onClose()}>
      <div className="modal">
        <div className="modal-header">
          <div className="modal-title">{restaurant.name}</div>
          <button className="modal-close" onClick={onClose}>✕</button>
        </div>

        {orderSuccess && <div className="status-bar success">{orderSuccess}</div>}
        {error && <div className="error-msg">{error}</div>}

        {step === 'menu' && (
          <>
            {loading ? (
              <div className="text-muted">Loading menu...</div>
            ) : menu.length === 0 ? (
              <div className="text-muted">No menu items available.</div>
            ) : (
              <>
                {categories.map((cat) => (
                  <div key={cat} style={{ marginBottom: '1rem' }}>
                    <div style={{ fontWeight: 700, color: '#6b7280', fontSize: '0.75rem', textTransform: 'uppercase', marginBottom: '0.5rem' }}>{cat}</div>
                    {menu.filter((m) => m.category === cat).map((item) => (
                      <div key={item.id} className="menu-item-row">
                        <div className="menu-item-info">
                          <div className="menu-item-name">{item.name}</div>
                          {item.description && <div className="menu-item-desc">{item.description}</div>}
                          <div className="menu-item-meta">
                            <span>{item.prep_time_minutes} min</span>
                          </div>
                        </div>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                          <span className="menu-item-price">${item.price.toFixed(2)}</span>
                          <input
                            type="number"
                            min={0}
                            className="qty-input"
                            value={getQty(item.id)}
                            onChange={(e) => setQty(item, parseInt(e.target.value) || 0)}
                          />
                        </div>
                      </div>
                    ))}
                  </div>
                ))}

                {cart.length > 0 && (
                  <div style={{ marginTop: '1rem', paddingTop: '1rem', borderTop: '1px solid #e5e7eb' }}>
                    <div className="flex-between">
                      <strong>Total: ${total.toFixed(2)}</strong>
                      <button className="btn btn-primary" onClick={() => setStep('order')}>
                        Place Order ({cart.length} items)
                      </button>
                    </div>
                  </div>
                )}
              </>
            )}
          </>
        )}

        {step === 'order' && !orderSuccess && (
          <form onSubmit={handleOrder}>
            <div style={{ marginBottom: '1rem' }}>
              <strong>Order Summary</strong>
              {cart.map((c) => (
                <div key={c.menu_item_id} className="flex-between text-sm mt-2">
                  <span>{c.name} x{c.quantity}</span>
                  <span>${(c.unit_price * c.quantity).toFixed(2)}</span>
                </div>
              ))}
              <div className="flex-between mt-3" style={{ fontWeight: 700 }}>
                <span>Total</span><span>${total.toFixed(2)}</span>
              </div>
            </div>
            <hr className="divider" />
            <div className="form-group">
              <label>Delivery Address</label>
              <input
                className="input"
                value={deliveryAddress}
                onChange={(e) => setDeliveryAddress(e.target.value)}
                placeholder="123 Main St, City"
                required
              />
            </div>
            <div className="grid-2">
              <div className="form-group">
                <label>Latitude</label>
                <input className="input" value={lat} onChange={(e) => setLat(e.target.value)} placeholder="0.000000" />
              </div>
              <div className="form-group">
                <label>Longitude</label>
                <input className="input" value={lng} onChange={(e) => setLng(e.target.value)} placeholder="0.000000" />
              </div>
            </div>
            <div className="flex-gap">
              <button type="button" className="btn btn-secondary" onClick={() => setStep('menu')}>Back</button>
              <button type="submit" className="btn btn-primary" disabled={ordering}>
                {ordering ? 'Placing order...' : `Confirm Order — $${total.toFixed(2)}`}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  )
}

export default function CustomerPage() {
  const { user, logout } = useAuth()
  const [tab, setTab] = useState<Tab>('restaurants')

  // Restaurants
  const [restaurants, setRestaurants] = useState<Restaurant[]>([])
  const [restLoading, setRestLoading] = useState(false)
  const [selectedRestaurant, setSelectedRestaurant] = useState<Restaurant | null>(null)

  // Orders
  const [orders, setOrders] = useState<Order[]>([])
  const [ordersLoading, setOrdersLoading] = useState(false)
  const [cancellingId, setCancellingId] = useState<number | null>(null)

  // Track
  const [trackId, setTrackId] = useState('')
  const [location, setLocation] = useState<LocationData | null>(null)
  const [trackError, setTrackError] = useState('')
  const [tracking, setTracking] = useState(false)

  useEffect(() => {
    if (tab === 'restaurants') {
      setRestLoading(true)
      restaurantsApi.listRestaurants()
        .then((res) => setRestaurants(res.data.data || []))
        .catch(() => {})
        .finally(() => setRestLoading(false))
    }
    if (tab === 'orders') {
      loadOrders()
    }
  }, [tab])

  const loadOrders = () => {
    setOrdersLoading(true)
    ordersApi.myOrders()
      .then((res) => setOrders(res.data.data || []))
      .catch(() => {})
      .finally(() => setOrdersLoading(false))
  }

  const handleCancel = async (id: number) => {
    setCancellingId(id)
    try {
      await ordersApi.cancelOrder(id)
      loadOrders()
    } catch {
      // ignore
    } finally {
      setCancellingId(null)
    }
  }

  const handleTrack = useCallback(async () => {
    if (!trackId) return
    setTrackError('')
    setTracking(true)
    try {
      const res = await deliveryApi.getLocation(parseInt(trackId))
      setLocation(res.data.data)
    } catch {
      setTrackError('Could not retrieve location. Check the delivery ID.')
      setLocation(null)
    } finally {
      setTracking(false)
    }
  }, [trackId])

  // Auto-refresh location every 5s when tracking
  useEffect(() => {
    if (tab !== 'track' || !trackId || !location) return
    const timer = setInterval(() => { handleTrack() }, 5000)
    return () => clearInterval(timer)
  }, [tab, trackId, location, handleTrack])

  return (
    <>
      <nav className="navbar">
        <span className="navbar-brand">FoodApp | Customer</span>
        <div className="navbar-links">
          <span className="text-sm" style={{ color: '#93c5fd' }}>{user?.email}</span>
          <button onClick={logout}>Logout</button>
        </div>
      </nav>

      <div className="container">
        <div className="page-header">
          <div className="page-title">Welcome back!</div>
        </div>

        <div className="tabs">
          <button className={`tab ${tab === 'restaurants' ? 'active' : ''}`} onClick={() => setTab('restaurants')}>
            Restaurants
          </button>
          <button className={`tab ${tab === 'orders' ? 'active' : ''}`} onClick={() => setTab('orders')}>
            My Orders
          </button>
          <button className={`tab ${tab === 'track' ? 'active' : ''}`} onClick={() => setTab('track')}>
            Track Delivery
          </button>
        </div>

        {/* Restaurants Tab */}
        {tab === 'restaurants' && (
          <>
            {restLoading ? (
              <div className="text-muted">Loading restaurants...</div>
            ) : restaurants.length === 0 ? (
              <div className="card"><div className="text-muted">No restaurants available.</div></div>
            ) : (
              <div className="grid-3">
                {restaurants.map((r) => (
                  <div key={r.id} className="card restaurant-card" onClick={() => setSelectedRestaurant(r)}>
                    <div className="restaurant-name">{r.name}</div>
                    <div className="restaurant-meta">{r.address}</div>
                    {r.cuisine_types && (
                      <div className="restaurant-meta">{r.cuisine_types.join(', ')}</div>
                    )}
                    <div className="flex-gap mt-2">
                      <span className={r.is_active ? 'badge badge-active' : 'badge badge-danger'}>
                        {r.is_active ? 'Open' : 'Closed'}
                      </span>
                      {r.rating && <span className="badge badge-neutral">★ {r.rating}</span>}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </>
        )}

        {/* Orders Tab */}
        {tab === 'orders' && (
          <>
            <div className="flex-between mb-4">
              <span className="text-muted">{orders.length} order(s)</span>
              <button className="btn btn-secondary btn-sm" onClick={loadOrders}>Refresh</button>
            </div>
            {ordersLoading ? (
              <div className="text-muted">Loading orders...</div>
            ) : orders.length === 0 ? (
              <div className="card"><div className="text-muted">No orders yet. Go order some food!</div></div>
            ) : (
              orders.map((order) => (
                <div key={order.id} className="card">
                  <div className="order-row">
                    <div className="order-info">
                      <div className="flex-gap mb-1">
                        <strong>Order #{order.id}</strong>
                        <span className={statusBadgeClass(order.status)}>{order.status}</span>
                      </div>
                      <div className="text-sm text-muted">Restaurant: #{order.restaurant_id}</div>
                      {order.delivery_address && (
                        <div className="text-sm text-muted">{order.delivery_address}</div>
                      )}
                      <div className="text-sm mt-1">
                        <strong>${order.total_amount?.toFixed(2)}</strong>
                        {order.items && <span className="text-muted"> · {order.items.length} items</span>}
                      </div>
                    </div>
                    <div className="order-actions">
                      {(order.status === 'PENDING' || order.status === 'CONFIRMED') && (
                        <button
                          className="btn btn-danger btn-sm"
                          disabled={cancellingId === order.id}
                          onClick={() => handleCancel(order.id)}
                        >
                          {cancellingId === order.id ? 'Cancelling...' : 'Cancel'}
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              ))
            )}
          </>
        )}

        {/* Track Delivery Tab */}
        {tab === 'track' && (
          <div className="card">
            <div style={{ fontWeight: 600, marginBottom: '1rem' }}>Track your delivery</div>
            <div className="flex-gap mb-4">
              <input
                className="input"
                style={{ maxWidth: 240 }}
                value={trackId}
                onChange={(e) => setTrackId(e.target.value)}
                placeholder="Enter delivery ID"
                type="number"
              />
              <button className="btn btn-primary" onClick={handleTrack} disabled={tracking || !trackId}>
                {tracking ? 'Tracking...' : 'Track'}
              </button>
            </div>

            {trackError && <div className="error-msg">{trackError}</div>}

            {location && (
              <div className="location-display">
                <div className="text-sm text-muted mb-1">Driver Location</div>
                <div className="location-coords">
                  Lat: {location.lat.toFixed(6)}, Lng: {location.lng.toFixed(6)}
                </div>
                {location.updated_at && (
                  <div className="text-sm text-muted mt-1">
                    Updated: {new Date(location.updated_at).toLocaleTimeString()}
                  </div>
                )}
                <div className="status-bar mt-3" style={{ marginBottom: 0 }}>
                  ETA: Estimating based on current location... (auto-refreshes every 5s)
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {selectedRestaurant && (
        <MenuModal restaurant={selectedRestaurant} onClose={() => setSelectedRestaurant(null)} />
      )}
    </>
  )
}
