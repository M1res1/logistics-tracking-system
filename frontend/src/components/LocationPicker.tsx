import { useState, useEffect, useRef } from 'react'
import { MapContainer, TileLayer, Marker, useMapEvents } from 'react-leaflet'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

// Fix default marker icons
delete (L.Icon.Default.prototype as unknown as Record<string, unknown>)._getIconUrl
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
  iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
  shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
})

interface Props {
  lat: number
  lng: number
  onChange: (lat: number, lng: number, address?: string) => void
  height?: number
}

function ClickHandler({ onMove }: { onMove: (lat: number, lng: number) => void }) {
  useMapEvents({
    click(e) { onMove(e.latlng.lat, e.latlng.lng) },
  })
  return null
}

export default function LocationPicker({ lat, lng, onChange, height = 260 }: Props) {
  const [search, setSearch] = useState('')
  const [searching, setSearching] = useState(false)
  const [error, setError] = useState('')
  const mapRef = useRef<L.Map | null>(null)

  const center: [number, number] = lat && lng ? [lat, lng] : [41.2995, 69.2401]

  const handleSearch = async () => {
    if (!search.trim()) return
    setSearching(true)
    setError('')
    try {
      const url = `https://nominatim.openstreetmap.org/search?q=${encodeURIComponent(search)}&format=json&limit=1`
      const res = await fetch(url, { headers: { 'Accept-Language': 'en' } })
      const data = await res.json()
      if (data.length === 0) { setError('Address not found'); return }
      const { lat: rlat, lon: rlng, display_name } = data[0]
      const newLat = parseFloat(rlat)
      const newLng = parseFloat(rlng)
      onChange(newLat, newLng, display_name)
      mapRef.current?.flyTo([newLat, newLng], 15)
    } catch {
      setError('Search failed. Try clicking the map instead.')
    } finally {
      setSearching(false)
    }
  }

  useEffect(() => {
    if (lat && lng && mapRef.current) {
      mapRef.current.setView([lat, lng], mapRef.current.getZoom())
    }
  }, [lat, lng])

  return (
    <div>
      <div className="map-search">
        <input
          className="input"
          placeholder="Search address..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
        />
        <button className="btn btn-secondary" onClick={handleSearch} disabled={searching} style={{ whiteSpace: 'nowrap' }}>
          {searching ? '...' : '🔍 Search'}
        </button>
      </div>
      {error && <div style={{ fontSize: '0.8rem', color: 'var(--danger)', marginBottom: '0.5rem' }}>{error}</div>}
      <div className="map-wrapper" style={{ height }}>
        <MapContainer
          center={center}
          zoom={13}
          style={{ width: '100%', height: '100%' }}
          ref={mapRef}
        >
          <TileLayer
            url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
            attribution='&copy; <a href="https://openstreetmap.org">OpenStreetMap</a>'
          />
          <ClickHandler onMove={(la, ln) => onChange(la, ln)} />
          {lat && lng && <Marker position={[lat, lng]} />}
        </MapContainer>
      </div>
      <div style={{ fontSize: '0.75rem', color: 'var(--gray-400)', marginTop: '0.35rem' }}>
        📍 Click the map or search to set location
        {lat && lng && ` · ${lat.toFixed(5)}, ${lng.toFixed(5)}`}
      </div>
    </div>
  )
}
