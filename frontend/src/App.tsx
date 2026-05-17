import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './contexts/AuthContext'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import CustomerPage from './pages/CustomerPage'
import DriverPage from './pages/DriverPage'
import RestaurantPage from './pages/RestaurantPage'

function RoleRouter() {
  const { user, loading } = useAuth()
  if (loading) return <div className="container" style={{ paddingTop: '2rem' }}>Loading...</div>
  if (!user) return <Navigate to="/login" replace />
  if (user.user_type === 'CUSTOMER') return <CustomerPage />
  if (user.user_type === 'DRIVER') return <DriverPage />
  if (user.user_type === 'RESTAURANT') return <RestaurantPage />
  return <div>Unknown role</div>
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/*" element={<RoleRouter />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
