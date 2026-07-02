import { createContext, useContext, useEffect, useState, type ReactNode } from "react"

// 前端假登录：平台目前没有真实的账号/Session 后端（gateway 的鉴权只针对
// 机器调用的 API Key，不是用户登录），这里只是先把 Portal / Console 的
// 路由结构和交互跑通。接入真实账号体系后，这个 Context 需要换成基于
// Session Cookie 的服务端校验（产品规格文档 §6.1）。
const STORAGE_KEY = "agentmesh_mock_auth"

interface AuthContextValue {
  isAuthenticated: boolean
  login: () => void
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(() => localStorage.getItem(STORAGE_KEY) === "1")

  useEffect(() => {
    if (isAuthenticated) {
      localStorage.setItem(STORAGE_KEY, "1")
    } else {
      localStorage.removeItem(STORAGE_KEY)
    }
  }, [isAuthenticated])

  const login = () => setIsAuthenticated(true)
  const logout = () => setIsAuthenticated(false)

  return <AuthContext.Provider value={{ isAuthenticated, login, logout }}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error("useAuth must be used within AuthProvider")
  return ctx
}
