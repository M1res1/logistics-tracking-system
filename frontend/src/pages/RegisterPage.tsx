import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'

const ROLES = [
  { value: 'CUSTOMER',   label: '🛒 Customer',         desc: 'Browse & order food' },
  { value: 'DRIVER',     label: '🚗 Driver',            desc: 'Deliver orders & earn' },
  { value: 'RESTAURANT', label: '🍽️ Restaurant Owner',  desc: 'Manage your restaurant' },
]

export default function RegisterPage() {
  const { register } = useAuth()
  const navigate = useNavigate()
  const [form, setForm] = useState({ email: '', password: '', phone: '', user_type: 'CUSTOMER' })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const set = (k: string, v: string) => setForm((p) => ({ ...p, [k]: v }))

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await register(form)
      navigate('/login')
    } catch {
      setError('Registration failed. Email may already be in use.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="auth-wrapper">
      <div className="auth-card" style={{ maxWidth: 480 }}>
        <div className="auth-logo">🍕</div>
        <div className="auth-title">Create account</div>
        <div className="auth-subtitle">Join FoodApp today</div>

        {error && <div className="alert alert-error">{error}</div>}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Email</label>
            <div className="input-group">
              <span className="input-icon">✉️</span>
              <input type="email" className="input" value={form.email}
                onChange={(e) => set('email', e.target.value)} placeholder="you@example.com" required />
            </div>
          </div>
          <div className="grid-2">
            <div className="form-group">
              <label>Password</label>
              <input type="password" className="input" value={form.password}
                onChange={(e) => set('password', e.target.value)} placeholder="Min 8 chars" minLength={8} required />
            </div>
            <div className="form-group">
              <label>Phone</label>
              <input type="tel" className="input" value={form.phone}
                onChange={(e) => set('phone', e.target.value)} placeholder="+998901234567" />
            </div>
          </div>

          <div className="form-group">
            <label>Account type</label>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {ROLES.map((r) => (
                <label key={r.value} style={{
                  display: 'flex', alignItems: 'center', gap: '0.75rem',
                  padding: '0.75rem 1rem', border: '1.5px solid',
                  borderColor: form.user_type === r.value ? 'var(--primary)' : 'var(--gray-200)',
                  borderRadius: '10px', cursor: 'pointer',
                  background: form.user_type === r.value ? 'var(--primary-light)' : 'white',
                  transition: 'all .15s',
                }}>
                  <input type="radio" name="user_type" value={r.value}
                    checked={form.user_type === r.value}
                    onChange={() => set('user_type', r.value)}
                    style={{ display: 'none' }} />
                  <span style={{ fontSize: '1.25rem' }}>{r.label.split(' ')[0]}</span>
                  <div>
                    <div style={{ fontWeight: 600, fontSize: '0.875rem' }}>{r.label.slice(r.label.indexOf(' ') + 1)}</div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--gray-500)' }}>{r.desc}</div>
                  </div>
                  {form.user_type === r.value && (
                    <span style={{ marginLeft: 'auto', color: 'var(--primary)', fontWeight: 700 }}>✓</span>
                  )}
                </label>
              ))}
            </div>
          </div>

          <button type="submit" className="btn btn-primary btn-block btn-lg" disabled={loading} style={{ marginTop: '0.75rem' }}>
            {loading ? 'Creating account...' : 'Create Account →'}
          </button>
        </form>

        <div className="auth-link">
          Already have an account? <Link to="/login">Sign in</Link>
        </div>
      </div>
    </div>
  )
}
