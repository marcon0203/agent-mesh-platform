import { Link, Outlet, useLocation } from "react-router-dom"
import { cn } from "@/lib/utils"
import { LogoMark } from "@/components/Logo"
import { useAuth } from "@/lib/auth"

const NAV_ITEMS = [
  { to: "/console", label: "控制台", end: true },
  { to: "/console/builder", label: "Agent 管理" },
  { to: "/console/model-providers", label: "模型供应商" },
  { to: "/console/workbench", label: "Workbench" },
]

// Console 是登录后的管理后台：Agent 装配/发布、模型供应商配置、Workbench 调试、
// API Key 与用量。对应产品规格文档 §6.1 里"开发者在控制台为每个应用生成 Key"、
// "Workbench 走平台登录态"的设计——和 Portal（宣传 + 能力市场浏览）分开。
export default function ConsoleLayout() {
  const location = useLocation()
  const { logout } = useAuth()

  // 退出登录后不用手动跳转：登录态一清空，ProtectedRoute 自己就会把当前
  // 受保护路由重定向到 /login（这也是大多数产品退出登录后的常见落点）。
  const handleLogout = () => logout()

  return (
    <div className="min-h-screen">
      <nav className="sticky top-0 z-50 border-b border-glass-border bg-background/70 backdrop-blur-xl">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-8 py-4">
          <Link to="/console" className="flex items-center gap-2.5 font-display text-lg font-semibold">
            <LogoMark className="size-7 drop-shadow-[0_0_6px_rgba(90,166,255,0.5)]" />
            枢络
            <span className="-ml-1 font-mono text-xs font-normal tracking-wide text-muted-foreground">
              Console
            </span>
          </Link>
          <ul className="flex items-center gap-1.5">
            {NAV_ITEMS.map((item) => {
              const active = item.end ? location.pathname === item.to : location.pathname.startsWith(item.to)
              return (
                <li key={item.to}>
                  <Link
                    to={item.to}
                    className={cn(
                      "rounded-full px-3.5 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground",
                      active && "bg-primary/15 text-[#5aa6ff]"
                    )}
                  >
                    {item.label}
                  </Link>
                </li>
              )
            })}
            <li>
              <button
                onClick={handleLogout}
                className="rounded-full px-3.5 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
              >
                退出登录
              </button>
            </li>
          </ul>
        </div>
      </nav>
      <main className="mx-auto max-w-6xl px-8 py-12">
        <Outlet />
      </main>
    </div>
  )
}
