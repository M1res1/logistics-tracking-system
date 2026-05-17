import { useState, useEffect, useCallback, lazy, Suspense } from 'react'
import { useAuth } from '../contexts/AuthContext'
import * as restaurantsApi from '../api/restaurants'
import * as ordersApi from '../api/orders'
import * as deliveryApi from '../api/delivery'
import * as paymentsApi from '../api/payments'
import type { Restaurant, MenuItem } from '../api/restaurants'
import type { Order } from '../api/orders'
import type { Wallet } from '../api/payments'

const LocationPicker = lazy(() => import('../components/LocationPicker'))

type Tab = 'restaurants' | 'orders' | 'track' | 'wallet'

const CUISINES_EMOJI: Record<string, string> = {
  pizza: '🍕', burger: '🍔', sushi: '🍣', chinese: '🥡', indian: '🍛',
  mexican: '🌮', italian: '🍝', thai: '🍜', coffee: '☕', dessert: '🍰',
  default: '🍽️',
}

function cuisineEmoji(types: string) {
  const t = (types || '').toLowerCase()
  for (const [k, v] of Object.entries(CUISINES_EMOJI)) {
    if (k !== 'default' && t.includes(k)) return v
  }
  return CUISINES_EMOJI.default
}

function statusColor(s: string) {
  switch (s?.toUpperCase()) {
    case 'PENDING': return 'badge-pending'
    case 'CONFIRMED': case 'PREPARING': case 'READY': return 'badge-active'
    case 'ASSIGNED': case 'IN_TRANSIT': return 'badge-info'
    case 'DELIVERED': return 'badge-purple'
    case 'CANCELLED': return 'badge-danger'
    default: return 'badge-neutral'
  }
}

/* ── Menu modal ─────────────────────────────────────────── */
function MenuModal({ restaurant, onClose }: { restaurant: Restaurant; onClose: () => void }) {
  const [menu, setMenu] = useState<MenuItem[]>([])
  const [cart, setCart] = useState<Map<number, number>>(new Map())
  const [step, setStep] = useState<'menu' | 'order'>('menu')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [ordering, setOrdering] = useState(false)
  const [success, setSuccess] = useState('')
  const [address, setAddress] = useState('')
  const [lat, setLat] = useState(0)
  const [lng, setLng] = useState(0)

  useEffect(() => {
    restaurantsApi.getMenu(restaurant.id)
      .then((r) => setMenu(r.data.data || []))
      .catch(() => setError('Failed to load menu'))
      .finally(() => setLoading(false))
  }, [restaurant.id])

  const setQty = (id: number, qty: number) => {
    setCart((prev) => { const m = new Map(prev); qty <= 0 ? m.delete(id) : m.set(id, qty); return m })
  }

  const total = menu.reduce((s, i) => s + (cart.get(i.id) || 0) * i.price, 0)
  const cartItems = menu.filter((i) => cart.has(i.id))
  const categories = [...new Set(menu.map((m) => m.category))]

  const handleOrder = async (e: React.FormEvent) => {
    e.preventDefault()
    if (cart.size === 0) { setError('Add at least one item'); return }
    setOrdering(true); setError('')
    try {
      await ordersApi.createOrder({
        restaurant_id: restaurant.id,
        delivery_address: address,
        lat, lng,
        items: cartItems.map((i) => ({ menu_item_id: i.id, quantity: cart.get(i.id)!, unit_price: i.price })),
      })
      setSuccess('🎉 Order placed successfully!')
      setCart(new Map())
    } catch { setError('Failed to place order.') }
    finally { setOrdering(false) }
  }

  return (
    <div className="modal-overlay" onClick={(e) => e.target === e.currentTarget && onClose()}>
      <div className="modal">
        <div className="modal-header">
          <div>
            <div className="modal-title">{restaurant.name}</div>
            <div style={{ fontSize: '0.8rem', color: 'var(--gray-500)' }}>{restaurant.address}</div>
          </div>
          <button className="modal-close" onClick={onClose}>✕</button>
        </div>

        {success && <div className="alert alert-success">{success}</div>}
        {error && <div className="alert alert-error">{error}</div>}

        {step === 'menu' && (
          <>
            {loading ? <div className="loading-state"><span className="spinner" />Loading menu...</div>
              : menu.length === 0 ? <div className="empty-state"><div className="empty-state-icon">📋</div><div>No items yet</div></div>
              : categories.map((cat) => (
                <div key={cat} style={{ marginBottom: '1rem' }}>
                  <div style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--gray-400)', textTransform: 'uppercase', letterSpacing: '.08em', marginBottom: '0.5rem' }}>{cat}</div>
                  {menu.filter((m) => m.category === cat).map((item) => (
                    <div key={item.id} className="menu-item-row">
                      <div className="menu-item-info">
                        <div className="menu-item-name">{item.name}</div>
                        {item.description && <div className="menu-item-desc">{item.description}</div>}
                        <div className="menu-item-meta">⏱ {item.prep_time_minutes} min</div>
                      </div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                        <span className="menu-item-price">${item.price.toFixed(2)}</span>
                        <div className="qty-stepper">
                          <button className="qty-btn" onClick={() => setQty(item.id, (cart.get(item.id) || 0) - 1)}>−</button>
                          <span className="qty-value">{cart.get(item.id) || 0}</span>
                          <button className="qty-btn" onClick={() => setQty(item.id, (cart.get(item.id) || 0) + 1)}>+</button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ))
            }

            {cart.size > 0 && (
              <div style={{ background: 'var(--primary-light)', borderRadius: '10px', padding: '1rem', marginTop: '1rem' }}>
                <div className="flex-between">
                  <div>
                    <div style={{ fontWeight: 700, color: 'var(--primary-dark)' }}>${total.toFixed(2)}</div>
                    <div style={{ fontSize: '0.8rem', color: 'var(--gray-500)' }}>{cart.size} item(s)</div>
                  </div>
                  <button className="btn btn-primary" onClick={() => setStep('order')}>
                    Checkout →
                  </button>
                </div>
              </div>
            )}
          </>
        )}

        {step === 'order' && !success && (
          <form onSubmit={handleOrder}>
            <div style={{ marginBottom: '1rem' }}>
              <div style={{ fontWeight: 700, marginBottom: '0.75rem' }}>Order Summary</div>
              {cartItems.map((i) => (
                <div key={i.id} className="flex-between text-sm" style={{ marginBottom: '0.4rem' }}>
                  <span>{i.name} × {cart.get(i.id)}</span>
                  <span style={{ fontWeight: 600 }}>${(i.price * cart.get(i.id)!).toFixed(2)}</span>
                </div>
              ))}
              <hr className="divider" />
              <div className="flex-between" style={{ fontWeight: 800, fontSize: '1.1rem' }}>
                <span>Total</span><span style={{ color: 'var(--primary)' }}>${total.toFixed(2)}</span>
              </div>
            </div>

            <div className="form-group">
              <label>Delivery Address</label>
              <input className="input" value={address} onChange={(e) => setAddress(e.target.value)}
                placeholder="Enter address or pick on map" required />
            </div>

            <div className="form-group">
              <label>📍 Pick location on map</label>
              <Suspense fallback={<div className="loading-state"><span className="spinner" /></div>}>
                <LocationPicker lat={lat} lng={lng} onChange={(la, ln, addr) => {
                  setLat(la); setLng(ln)
                  if (addr && !address) setAddress(addr.split(',').slice(0, 3).join(','))
                }} />
              </Suspense>
            </div>

            <div className="flex-gap" style={{ marginTop: '1rem' }}>
              <button type="button" className="btn btn-secondary" onClick={() => setStep('menu')}>← Back</button>
              <button type="submit" className="btn btn-primary" style={{ flex: 1 }} disabled={ordering}>
                {ordering ? 'Placing order...' : `Place Order · $${total.toFixed(2)}`}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  )
}

/* ── Wallet / Payment tab ─────────────────────────────────────────── */
function WalletTab({ userId }: { userId: number }) {
  const [wallet, setWallet] = useState<Wallet | null>(null)
  const [loading, setLoading] = useState(true)
  const [topupAmount, setTopupAmount] = useState('')
  const [cardNumber, setCardNumber] = useState('')
  const [expiry, setExpiry] = useState('')
  const [cvv, setCvv] = useState('')
  const [cardName, setCardName] = useState('')
  const [paying, setPaying] = useState(false)
  const [msg, setMsg] = useState('')
  const [msgType, setMsgType] = useState<'success' | 'error'>('success')

  useEffect(() => {
    paymentsApi.getWallet(userId)
      .then((r) => setWallet(r.data.data))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [userId])

  const formatCard = (v: string) => v.replace(/\D/g, '').slice(0, 16).replace(/(.{4})/g, '$1 ').trim()
  const formatExpiry = (v: string) => { const d = v.replace(/\D/g, '').slice(0, 4); return d.length >= 3 ? d.slice(0, 2) + '/' + d.slice(2) : d }

  const displayCard = cardNumber || '•••• •••• •••• ••••'
  const displayExpiry = expiry || 'MM/YY'
  const displayName = cardName || 'CARDHOLDER NAME'

  const handleTopup = async (e: React.FormEvent) => {
    e.preventDefault()
    const amount = parseFloat(topupAmount)
    if (!amount || amount <= 0) return
    if (!cardNumber.replace(/\s/g, '') || cardNumber.replace(/\s/g, '').length < 16) {
      setMsg('Enter a valid 16-digit card number'); setMsgType('error'); return
    }
    setPaying(true); setMsg('')
    try {
      await paymentsApi.topupWallet(userId, amount)
      const r = await paymentsApi.getWallet(userId)
      setWallet(r.data.data)
      setMsg(`✅ Successfully added $${amount.toFixed(2)} to your wallet!`)
      setMsgType('success')
      setTopupAmount(''); setCardNumber(''); setExpiry(''); setCvv(''); setCardName('')
    } catch {
      setMsg('Payment failed. Please check your card details.'); setMsgType('error')
    } finally { setPaying(false) }
  }

  if (loading) return <div className="loading-state"><span className="spinner" />Loading wallet...</div>

  return (
    <div>
      {/* Balance card */}
      <div className="wallet-balance-card">
        <div>
          <div className="wallet-label">Available Balance</div>
          <div className="wallet-amount">${(wallet?.balance || 0).toFixed(2)}</div>
        </div>
        <span style={{ fontSize: '3rem' }}>💳</span>
      </div>

      {msg && <div className={`alert alert-${msgType}`}>{msg}</div>}

      {/* Card visual */}
      <div className="payment-card-visual">
        <div>
          <div className="card-chip">💳</div>
          <div className="card-number">{displayCard}</div>
        </div>
        <div className="card-footer">
          <div>
            <div className="card-label">Cardholder</div>
            <div className="card-value">{displayName}</div>
          </div>
          <div>
            <div className="card-label">Expires</div>
            <div className="card-value">{displayExpiry}</div>
          </div>
          <div className="card-network">💜</div>
        </div>
      </div>

      {/* Top-up form */}
      <div className="card">
        <div className="card-title">💰 Add Money</div>
        <form onSubmit={handleTopup}>
          <div className="form-group">
            <label>Card Number</label>
            <input className="input" value={cardNumber} placeholder="1234 5678 9012 3456"
              onChange={(e) => setCardNumber(formatCard(e.target.value))} style={{ fontFamily: 'monospace', letterSpacing: '0.1em' }} />
          </div>
          <div className="grid-2">
            <div className="form-group">
              <label>Cardholder Name</label>
              <input className="input" value={cardName} placeholder="John Doe"
                onChange={(e) => setCardName(e.target.value.toUpperCase())} />
            </div>
            <div className="form-group">
              <label>Amount ($)</label>
              <input className="input" type="number" min="1" step="0.01" value={topupAmount}
                placeholder="10.00" onChange={(e) => setTopupAmount(e.target.value)} />
            </div>
          </div>
          <div className="grid-2">
            <div className="form-group">
              <label>Expiry</label>
              <input className="input" value={expiry} placeholder="MM/YY"
                onChange={(e) => setExpiry(formatExpiry(e.target.value))} maxLength={5} />
            </div>
            <div className="form-group">
              <label>CVV</label>
              <input className="input" value={cvv} placeholder="•••" type="password"
                onChange={(e) => setCvv(e.target.value.replace(/\D/g, '').slice(0, 3))} maxLength={3} />
            </div>
          </div>
          <button type="submit" className="btn btn-success btn-block" disabled={paying}>
            {paying ? 'Processing...' : `Add ${topupAmount ? '$' + parseFloat(topupAmount || '0').toFixed(2) : 'Money'}`}
          </button>
        </form>
      </div>
    </div>
  )
}

/* ── Main CustomerPage ─────────────────────────────────────────── */
export default function CustomerPage() {
  const { user, logout } = useAuth()
  const [tab, setTab] = useState<Tab>('restaurants')
  const [restaurants, setRestaurants] = useState<Restaurant[]>([])
  const [restLoading, setRestLoading] = useState(false)
  const [selectedRestaurant, setSelectedRestaurant] = useState<Restaurant | null>(null)
  const [orders, setOrders] = useState<Order[]>([])
  const [ordersLoading, setOrdersLoading] = useState(false)
  const [cancellingId, setCancellingId] = useState<number | null>(null)
  const [trackId, setTrackId] = useState('')
  const [location, setLocation] = useState<{ lat: number; lng: number; updated_at?: string } | null>(null)
  const [trackError, setTrackError] = useState('')
  const [tracking, setTracking] = useState(false)

  useEffect(() => {
    if (tab === 'restaurants') {
      setRestLoading(true)
      restaurantsApi.listRestaurants()
        .then((r) => setRestaurants(r.data.data?.restaurants || []))
        .catch(() => {})
        .finally(() => setRestLoading(false))
    }
    if (tab === 'orders') loadOrders()
  }, [tab])

  const loadOrders = () => {
    setOrdersLoading(true)
    ordersApi.myOrders()
      .then((r) => setOrders(r.data.data?.orders || []))
      .catch(() => {})
      .finally(() => setOrdersLoading(false))
  }

  const handleCancel = async (id: number) => {
    setCancellingId(id)
    try { await ordersApi.cancelOrder(id); loadOrders() }
    catch {} finally { setCancellingId(null) }
  }

  const handleTrack = useCallback(async () => {
    if (!trackId) return
    setTrackError(''); setTracking(true)
    try {
      const r = await deliveryApi.getLocation(parseInt(trackId))
      setLocation(r.data.data)
    } catch { setTrackError('Could not find location. Check the delivery ID.'); setLocation(null) }
    finally { setTracking(false) }
  }, [trackId])

  useEffect(() => {
    if (tab !== 'track' || !trackId || !location) return
    const t = setInterval(handleTrack, 5000)
    return () => clearInterval(t)
  }, [tab, trackId, location, handleTrack])

  return (
    <>
      <nav className="navbar">
        <span className="navbar-brand">FoodApp</span>
        <div className="navbar-links">
          <span className="navbar-user">👤 {user?.email}</span>
          <button onClick={logout}>Sign out</button>
        </div>
      </nav>

      <div className="container">
        <div className="page-header">
          <div className="page-title">What are you craving?</div>
          <div className="page-subtitle">Order food from the best restaurants near you</div>
        </div>

        <div className="tabs">
          {[
            { key: 'restaurants', label: '🏪 Restaurants' },
            { key: 'orders',      label: '📦 My Orders' },
            { key: 'track',       label: '📍 Track' },
            { key: 'wallet',      label: '💳 Wallet' },
          ].map(({ key, label }) => (
            <button key={key} className={`tab ${tab === key ? 'active' : ''}`} onClick={() => setTab(key as Tab)}>{label}</button>
          ))}
        </div>

        {/* ── Restaurants ── */}
        {tab === 'restaurants' && (
          restLoading ? <div className="loading-state"><span className="spinner" />Loading restaurants...</div>
          : restaurants.length === 0
            ? <div className="empty-state"><div className="empty-state-icon">🍽️</div><div className="empty-state-text">No restaurants yet.</div></div>
            : <div className="restaurant-grid">
                {restaurants.map((r) => (
                  <div key={r.id} className="card restaurant-card" onClick={() => setSelectedRestaurant(r)}>
                    <div className="restaurant-card-emoji">{cuisineEmoji(r.cuisine_types)}</div>
                    <div className="restaurant-name">{r.name}</div>
                    <div className="restaurant-meta">📍 {r.address}</div>
                    {r.cuisine_types && <div className="restaurant-meta">🍴 {r.cuisine_types}</div>}
                    <div className="restaurant-meta">🕐 {r.opening_time} – {r.closing_time}</div>
                    <div className="flex-gap mt-2">
                      <span className={r.is_active ? 'badge badge-active' : 'badge badge-danger'}>
                        {r.is_active ? '● Open' : '● Closed'}
                      </span>
                      {r.rating ? <span className="badge badge-neutral">★ {r.rating}</span> : null}
                    </div>
                  </div>
                ))}
              </div>
        )}

        {/* ── My Orders ── */}
        {tab === 'orders' && (
          <>
            <div className="flex-between mb-4">
              <span className="text-muted">{orders.length} order(s)</span>
              <button className="btn btn-secondary btn-sm" onClick={loadOrders}>↻ Refresh</button>
            </div>
            {ordersLoading ? <div className="loading-state"><span className="spinner" /></div>
              : orders.length === 0
                ? <div className="empty-state"><div className="empty-state-icon">📦</div><div className="empty-state-text">No orders yet. Go order some food!</div></div>
                : orders.map((order) => (
                  <div key={order.id} className="card order-card">
                    <div className="order-row">
                      <div className="order-info">
                        <div className="flex-gap mb-1">
                          <span className="order-id">Order #{order.id}</span>
                          <span className={`badge ${statusColor(order.status)}`}>{order.status}</span>
                        </div>
                        <div className="order-total">${order.total?.toFixed(2)}</div>
                        {order.delivery_address && <div className="text-sm text-muted mt-1">📍 {order.delivery_address}</div>}
                      </div>
                      <div className="order-actions">
                        {(order.status === 'PENDING' || order.status === 'CONFIRMED') && (
                          <button className="btn btn-danger btn-sm" disabled={cancellingId === order.id}
                            onClick={() => handleCancel(order.id)}>
                            {cancellingId === order.id ? '...' : 'Cancel'}
                          </button>
                        )}
                      </div>
                    </div>
                  </div>
                ))
            }
          </>
        )}

        {/* ── Track ── */}
        {tab === 'track' && (
          <div className="card">
            <div className="card-title">📍 Track your delivery</div>
            <div className="flex-gap mb-4">
              <input className="input" style={{ maxWidth: 220 }} value={trackId}
                onChange={(e) => setTrackId(e.target.value)} placeholder="Delivery ID" type="number" />
              <button className="btn btn-primary" onClick={handleTrack} disabled={tracking || !trackId}>
                {tracking ? 'Tracking...' : 'Track'}
              </button>
            </div>
            {trackError && <div className="alert alert-error">{trackError}</div>}
            {location && (
              <>
                <div className="location-display">
                  <div style={{ fontSize: '0.75rem', color: 'var(--gray-500)', marginBottom: '0.25rem' }}>Driver location</div>
                  <div className="location-coords">📍 {location.lat.toFixed(5)}, {location.lng.toFixed(5)}</div>
                  {location.updated_at && (
                    <div className="text-sm text-muted mt-1">Updated {new Date(location.updated_at).toLocaleTimeString()}</div>
                  )}
                </div>
                <Suspense fallback={null}>
                  <div style={{ marginTop: '1rem' }}>
                    <LocationPicker lat={location.lat} lng={location.lng} onChange={() => {}} height={300} />
                  </div>
                </Suspense>
                <div className="alert alert-info mt-3" style={{ marginBottom: 0 }}>Auto-refreshes every 5 seconds</div>
              </>
            )}
          </div>
        )}

        {/* ── Wallet ── */}
        {tab === 'wallet' && user && <WalletTab userId={user.id} />}
      </div>

      {selectedRestaurant && (
        <MenuModal restaurant={selectedRestaurant} onClose={() => setSelectedRestaurant(null)} />
      )}
    </>
  )
}
