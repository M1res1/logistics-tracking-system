import { useState, lazy, Suspense } from 'react'
import { useAuth } from '../contexts/AuthContext'
import * as deliveryApi from '../api/delivery'

const LocationPicker = lazy(() => import('../components/LocationPicker'))

type Tab = 'deliveries' | 'location'

export default function DriverPage() {
  const { user, logout } = useAuth()
  const [tab, setTab] = useState<Tab>('deliveries')

  // Deliveries
  const [deliveryId, setDeliveryId] = useState('')
  const [actionMsg, setActionMsg] = useState('')
  const [actionErr, setActionErr] = useState('')
  const [acting, setActing] = useState(false)

  // Location
  const [lat, setLat] = useState(0)
  const [lng, setLng] = useState(0)
  const [locationId, setLocationId] = useState('')
  const [locationMsg, setLocationMsg] = useState('')
  const [sendingLoc, setSendingLoc] = useState(false)
  const [watchId, setWatchId] = useState<number | null>(null)

  const doAction = async (fn: (id: number) => Promise<unknown>, label: string) => {
    if (!deliveryId) return
    setActing(true); setActionMsg(''); setActionErr('')
    try {
      await fn(parseInt(deliveryId))
      setActionMsg(`✅ Delivery #${deliveryId} ${label} successfully.`)
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'Action failed'
      setActionErr(msg)
    } finally { setActing(false) }
  }

  const sendLocation = async () => {
    if (!locationId || !lat || !lng) return
    setSendingLoc(true); setLocationMsg('')
    try {
      await deliveryApi.updateLocation(parseInt(locationId), lat, lng)
      setLocationMsg(`📍 Location sent: ${lat.toFixed(5)}, ${lng.toFixed(5)}`)
    } catch { setLocationMsg('❌ Failed to send location') }
    finally { setSendingLoc(false) }
  }

  const startGPS = () => {
    if (!navigator.geolocation) { setLocationMsg('GPS not supported'); return }
    const id = navigator.geolocation.watchPosition(
      (pos) => { setLat(pos.coords.latitude); setLng(pos.coords.longitude) },
      () => setLocationMsg('GPS error — use the map instead')
    )
    setWatchId(id)
    setLocationMsg('📡 GPS active — position updating...')
  }

  const stopGPS = () => {
    if (watchId !== null) { navigator.geolocation.clearWatch(watchId); setWatchId(null) }
    setLocationMsg('GPS stopped')
  }

  return (
    <>
      <nav className="navbar">
        <span className="navbar-brand">FoodApp</span>
        <div className="navbar-links">
          <span className="navbar-user">🚗 {user?.email}</span>
          <button onClick={logout}>Sign out</button>
        </div>
      </nav>

      <div className="container">
        <div className="page-header">
          <div className="page-title">Driver Dashboard</div>
          <div className="page-subtitle">Manage deliveries and share your location</div>
        </div>

        <div className="tabs">
          <button className={`tab ${tab === 'deliveries' ? 'active' : ''}`} onClick={() => setTab('deliveries')}>📦 Deliveries</button>
          <button className={`tab ${tab === 'location' ? 'active' : ''}`} onClick={() => setTab('location')}>📍 My Location</button>
        </div>

        {/* ── Deliveries ── */}
        {tab === 'deliveries' && (
          <div className="card">
            <div className="card-title">Manage Delivery</div>

            <div className="form-group">
              <label>Delivery ID</label>
              <input className="input" value={deliveryId} onChange={(e) => setDeliveryId(e.target.value)}
                placeholder="Enter delivery ID" type="number" style={{ maxWidth: 240 }} />
            </div>

            {actionMsg && <div className="alert alert-success">{actionMsg}</div>}
            {actionErr && <div className="alert alert-error">{actionErr}</div>}

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(160px, 1fr))', gap: '0.75rem', marginTop: '0.5rem' }}>
              {[
                { label: '✅ Accept',   fn: deliveryApi.acceptDelivery,   color: 'btn-success',   desc: 'Accept the delivery assignment' },
                { label: '📦 Pick Up',  fn: deliveryApi.pickupDelivery,   color: 'btn-primary',   desc: 'Confirm you picked up the order' },
                { label: '🏁 Complete', fn: deliveryApi.completeDelivery, color: 'btn-outline',   desc: 'Mark delivery as completed' },
              ].map(({ label, fn, color, desc }) => (
                <div key={label} style={{ background: 'var(--gray-50)', borderRadius: '10px', padding: '1rem', border: '1px solid var(--gray-200)' }}>
                  <div style={{ fontSize: '0.75rem', color: 'var(--gray-500)', marginBottom: '0.5rem' }}>{desc}</div>
                  <button className={`btn ${color} btn-block`} disabled={acting || !deliveryId}
                    onClick={() => doAction(fn, label.split(' ').slice(1).join(' ').toLowerCase())}>
                    {label}
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* ── Location ── */}
        {tab === 'location' && (
          <div className="card">
            <div className="card-title">Share Your Location</div>

            <div className="form-group">
              <label>Delivery ID</label>
              <input className="input" value={locationId} onChange={(e) => setLocationId(e.target.value)}
                placeholder="Enter delivery ID" type="number" style={{ maxWidth: 240 }} />
            </div>

            <div className="form-group">
              <label>📍 Your position — click map or use GPS</label>
              <Suspense fallback={<div className="loading-state"><span className="spinner" /></div>}>
                <LocationPicker lat={lat} lng={lng} onChange={(la, ln) => { setLat(la); setLng(ln) }} height={320} />
              </Suspense>
            </div>

            {locationMsg && (
              <div className={`alert ${locationMsg.startsWith('❌') ? 'alert-error' : 'alert-info'}`}>{locationMsg}</div>
            )}

            <div className="flex-gap">
              {watchId === null
                ? <button className="btn btn-secondary" onClick={startGPS}>📡 Use GPS</button>
                : <button className="btn btn-danger" onClick={stopGPS}>⏹ Stop GPS</button>
              }
              <button className="btn btn-primary" disabled={sendingLoc || !locationId || (!lat && !lng)}
                onClick={sendLocation}>
                {sendingLoc ? 'Sending...' : '📤 Send Location'}
              </button>
            </div>
          </div>
        )}
      </div>
    </>
  )
}
