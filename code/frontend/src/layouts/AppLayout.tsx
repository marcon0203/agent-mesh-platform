import { Link, Outlet, useLocation } from "react-router-dom"
import { cn } from "@/lib/utils"
import { LogoMark } from "@/components/Logo"

const NAV_ITEMS = [
  { to: "/marketplace", label: "能力市场" },
  { to: "/builder", label: "构建 Agent" },
  { to: "/model-providers", label: "模型供应商" },
  { to: "/workbench", label: "Workbench" },
  { to: "/dashboard", label: "控制台" },
]

export default function AppLayout() {
  const location = useLocation()

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
          <ul className="flex items-center gap-1.5">
            {NAV_ITEMS.map((item) => {
              const active = location.pathname.startsWith(item.to)
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
        </div>
      </nav>
      <main className="mx-auto max-w-6xl px-8 py-12">
        <Outlet />
      </main>
    </div>
  )
}
