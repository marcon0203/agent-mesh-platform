import { Navigate, Outlet, useLocation } from "react-router-dom"
import { useAuth } from "@/lib/auth"

// 挡住 /console/* 路由：未登录（假登录，见 lib/auth.tsx）时重定向到 /login，
// 并带上 from 参数，登录后跳回原本想去的页面。
export function ProtectedRoute() {
  const { isAuthenticated } = useAuth()
  const location = useLocation()

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location.pathname }} replace />
  }
  return <Outlet />
}
