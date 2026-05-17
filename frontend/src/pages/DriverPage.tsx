import { useState } from 'react'
import { useAuth } from '../contexts/AuthContext'
import * as deliveryApi from '../api/delivery'

type Tab = 'deliveries' | 'location'

interface ActionStatus {
  type: 'success' | 'error'
  message: string
}

export default function DriverPage() {
  const { user, logout } = useAuth()
  const [tab, setTab] = useState<Tab>('deliveries')

  // Delivery actions
  const [deliveryId, setDeliveryId] = useState('')
  const [actionStatus, setActionStatus] = useState<ActionStatus | null>(null)
  const [acting, setActing] = useState(false)

  // Location update
  const [locDeliveryId, setLocDeliveryId] = useState('')
  const [manualLat, setManualLat] = useState('')
  const [manualLng, setManualLng] = useState('')
  const [locStatus, setLocStatus] = useState<ActionStatus | null>(null)
  const [locLoading, setLocLoading] = useState(false)
  const [geoLoading, setGeoLoading] = useState(false)

  const doAction = async (action: () => Promise<unknown>, label: string) => {
    if (!deliveryId) { setActionStatus({ type: 'error', message: 'Please enter a Delivery ID.' }); return }
    setActing(true)
    setActionStatus(null)
    try {
      await action()
      setActionStatus({ type: 'success', message: `${label} successful!` })
    } catch {
      setActionStatus({ type: 'error', message: `${label} failed. Check the delivery ID and try again.` })
    } finally {
      setActing(false)
    }
  }

  const handleGetLocation = () => {
    setGeoLoading(true)
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setManualLat(pos.coords.latitude.toFixed(6))
        setManualLng(pos.coords.longitude.toFixed(6))
        setGeoLoading(false)
      },
      () => {
        setLocStatus({ type: 'error', message: 'Could not get browser location. Enter coordinates manually.' })
        setGeoLoading(false)
      }
    )
  }

  const handleSendLocation = async () => {
    if (!locDeliveryId) { setLocStatus({ type: 'error', message: 'Please enter a Delivery ID.' }); return }
    if (!manualLat || !manualLng) { setLocStatus({ type: 'error', message: 'Please provide coordinates.' }); return }
    setLocLoading(true)
    setLocStatus(null)
    try {
      await deliveryApi.updateLocation(parseInt(locDeliveryId), parseFloat(manualLat), parseFloat(manualLng))
      setLocStatus({ type: 'success', message: 'Location updated successfully!' })
    } catch {
      setLocStatus({ type: 'error', message: 'Failed to update location.' })
    } finally {
      setLocLoading(false)
    }
  }

  return (
    <>
      <nav className="navbar">
        <span className="navbar-brand">FoodApp | Driver</span>
        <div className="navbar-links">
          <span className="text-sm" style={{ color: '#93c5fd' }}>{user?.email}</span>
          <button onClick={logout}>Logout</button>
        </div>
      </nav>

      <div className="container">
        <div className="page-header">
          <div className="page-title">Driver Dashboard</div>
        </div>

        <div className="tabs">
          <button className={`tab ${tab === 'deliveries' ? 'active' : ''}`} onClick={() => setTab('deliveries')}>
            My Deliveries
          </button>
          <button className={`tab ${tab === 'location' ? 'active' : ''}`} onClick={() => setTab('location')}>
            Go Online / Location
          </button>
        </div>

        {tab === 'deliveries' && (
          <div className="card">
            <div style={{ fontWeight: 600, marginBottom: '1rem' }}>Manage Deliveries</div>
            <div className="form-group">
              <label>Delivery ID</label>
              <input
                type="number"
                className="input"
                style={{ maxWidth: 240 }}
                value={deliveryId}
                onChange={(e) => setDeliveryId(e.target.value)}
                placeholder="Enter delivery ID"
              />
            </div>

            {actionStatus && (
              <div className={`status-bar ${actionStatus.type === 'error' ? 'error' : 'success'}`} style={{ marginBottom: '1rem' }}>
                {actionStatus.message}
              </div>
            )}

            <div className="flex-gap" style={{ flexWrap: 'wrap' }}>
              <button
                className="btn btn-primary"
                disabled={acting}
                onClick={() => doAction(() => deliveryApi.acceptDelivery(parseInt(deliveryId)), 'Accept')}
              >
                Accept Delivery
              </button>
              <button
                className="btn btn-success"
                disabled={acting}
                onClick={() => doAction(() => deliveryApi.pickupDelivery(parseInt(deliveryId)), 'Pickup')}
              >
                Mark Picked Up
              </button>
              <button
                className="btn btn-secondary"
                disabled={acting}
                onClick={() => doAction(() => deliveryApi.completeDelivery(parseInt(deliveryId)), 'Complete')}
              >
                Mark Complete
              </button>
            </div>

            <hr className="divider" />
            <div className="text-muted text-sm">
              <strong>How it works:</strong> Enter a delivery ID above, then use the action buttons to progress the delivery through its lifecycle: Accept → Pickup → Complete.
            </div>
          </div>
        )}

        {tab === 'location' && (
          <div className="card">
            <div style={{ fontWeight: 600, marginBottom: '1rem' }}>Update Your Location</div>

            <div className="form-group">
              <label>Delivery ID</label>
              <input
                type="number"
                className="input"
                style={{ maxWidth: 240 }}
                value={locDeliveryId}
                onChange={(e) => setLocDeliveryId(e.target.value)}
                placeholder="Enter delivery ID"
              />
            </div>

            <div style={{ marginBottom: '1rem' }}>
              <button
                className="btn btn-secondary"
                onClick={handleGetLocation}
                disabled={geoLoading}
                style={{ marginBottom: '0.75rem' }}
              >
                {geoLoading ? 'Getting location...' : 'Use My Browser Location'}
              </button>
              <div className="text-sm text-muted mb-2" style={{ marginTop: '0.5rem' }}>
                Or enter coordinates manually:
              </div>
              <div className="grid-2">
                <div className="form-group">
                  <label>Latitude</label>
                  <input
                    className="input"
                    value={manualLat}
                    onChange={(e) => setManualLat(e.target.value)}
                    placeholder="e.g. 37.774929"
                  />
                </div>
                <div className="form-group">
                  <label>Longitude</label>
                  <input
                    className="input"
                    value={manualLng}
                    onChange={(e) => setManualLng(e.target.value)}
                    placeholder="e.g. -122.419418"
                  />
                </div>
              </div>
            </div>

            {locStatus && (
              <div className={`status-bar ${locStatus.type === 'error' ? 'error' : 'success'}`} style={{ marginBottom: '1rem' }}>
                {locStatus.message}
              </div>
            )}

            <button
              className="btn btn-primary"
              onClick={handleSendLocation}
              disabled={locLoading}
            >
              {locLoading ? 'Sending...' : 'Send Location'}
            </button>
          </div>
        )}
      </div>
    </>
  )
}
