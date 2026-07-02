import { Link, Outlet, useLocation } from "react-router-dom"
import { cn } from "@/lib/utils"
import { LogoMark } from "@/components/Logo"
import { Button } from "@/components/ui/button"
import { useAuth } from "@/lib/auth"

const NAV_ITEMS = [
  { to: "/", label: "首页" },
  { to: "/marketplace", label: "能力市场" },
]

// Portal 是对外的公开门户：宣传定位 + 能力市场浏览，不需要登录。
// 视觉参照 docs/design/portal-reference.html。控制台入口按钮根据登录态
// 分流到 /console（已登录）或 /login（未登录）。
export default function PortalLayout() {
  const location = useLocation()
  const { isAuthenticated } = useAuth()

  return (
    <div className="min-h-screen">
      <nav className="sticky top-0 z-50 border-b border-glass-border bg-background/70 backdrop-blur-xl">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-8 py-4">
          <Link to="/" className="flex items-center gap-2.5 font-display text-lg font-semibold">
            <LogoMark className="size-7 drop-shadow-[0_0_6px_rgba(90,166,255,0.5)]" />
            枢络
            <span className="-ml-1 font-mono text-xs font-normal tracking-wide text-muted-foreground">
              AgentMesh
            </span>
          </Link>
          <div className="flex items-center gap-1.5">
            <ul className="mr-2 flex items-center gap-1.5">
              {NAV_ITEMS.map((item) => {
                const active = location.pathname === item.to
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
            </ul>
            <Button asChild size="sm">
              <Link to={isAuthenticated ? "/console" : "/login"}>
                {isAuthenticated ? "进入控制台" : "登录"}
              </Link>
            </Button>
          </div>
        </div>
      </nav>
      <main>
        <Outlet />
      </main>
    </div>
  )
}
