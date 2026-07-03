import { useState } from "react"
import { Link, Outlet, useLocation } from "react-router-dom"
import {
  LayoutDashboard,
  Bot,
  Cpu,
  MessagesSquare,
  PanelLeftClose,
  PanelLeftOpen,
  LogOut,
  type LucideIcon,
} from "lucide-react"
import { cn } from "@/lib/utils"
import { LogoMark } from "@/components/Logo"
import { useAuth } from "@/lib/auth"

interface NavItem {
  to: string
  label: string
  icon: LucideIcon
  end?: boolean
}
interface NavGroup {
  key: string
  label: string
  items: NavItem[]
}

const NAV_GROUPS: NavGroup[] = [
  { key: "overview", label: "总览", items: [{ to: "/console", label: "控制台", icon: LayoutDashboard, end: true }] },
  {
    key: "orchestration",
    label: "编排与配置",
    items: [
      { to: "/console/builder", label: "Agent 管理", icon: Bot },
      { to: "/console/model-providers", label: "模型供应商", icon: Cpu },
    ],
  },
  {
    key: "debug",
    label: "调试与验证",
    items: [{ to: "/console/workbench", label: "Workbench", icon: MessagesSquare }],
  },
]

function isActive(item: NavItem, pathname: string) {
  return item.end ? pathname === item.to : pathname.startsWith(item.to)
}

function activeGroup(pathname: string): NavGroup {
  for (const group of NAV_GROUPS) {
    if (group.items.some((item) => isActive(item, pathname))) return group
  }
  return NAV_GROUPS[0]
}

// Console 布局参考阿里云百炼：顶部一条全局导航条（logo + 大分类 tab + 账号菜单），
// 左侧栏只展示当前大分类下的子功能——顶部切换"在哪个模块"，左侧栏切换
// "模块内哪个页面"，跟之前单层左侧栏（所有功能都摊平在一列里）是两种不同的
// 信息架构，模块变多之后顶部 tab 比一列到底的侧边栏更容易扫视。
export default function ConsoleLayout() {
  const location = useLocation()
  const { logout } = useAuth()
  const [collapsed, setCollapsed] = useState(false)
  const [userMenuOpen, setUserMenuOpen] = useState(false)

  const handleLogout = () => {
    setUserMenuOpen(false)
    logout()
  }

  const group = activeGroup(location.pathname)

  return (
    <div className="min-h-screen">
      <header className="sticky top-0 z-50 flex h-14 items-center gap-1 border-b border-glass-border bg-background/70 px-3 backdrop-blur-xl">
        <button
          onClick={() => setCollapsed((c) => !c)}
          title={collapsed ? "展开导航" : "收起导航"}
          className="flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-white/5 hover:text-foreground"
        >
          {collapsed ? <PanelLeftOpen className="size-4" /> : <PanelLeftClose className="size-4" />}
        </button>

        <Link to="/console" className="mr-4 flex shrink-0 items-center gap-2 pl-1">
          <LogoMark className="size-6 shrink-0 drop-shadow-[0_0_6px_rgba(90,166,255,0.5)]" />
          <span className="font-display text-base font-semibold">
            枢络
            <span className="ml-1 font-mono text-xs font-normal text-muted-foreground">Console</span>
          </span>
        </Link>

        <nav className="flex items-center gap-1">
          {NAV_GROUPS.map((g) => (
            <Link
              key={g.key}
              to={g.items[0].to}
              className={cn(
                "rounded-full px-3.5 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground",
                group.key === g.key && "bg-primary/15 text-[#5aa6ff]"
              )}
            >
              {g.label}
            </Link>
          ))}
        </nav>

        <div className="relative ml-auto">
          <button
            onClick={() => setUserMenuOpen((o) => !o)}
            className="flex items-center gap-2 rounded-full py-1 pl-1 pr-3 text-sm text-muted-foreground transition-colors hover:bg-white/5 hover:text-foreground"
          >
            <span className="flex size-7 items-center justify-center rounded-full bg-primary/20 font-display text-xs font-semibold text-[#5aa6ff]">
              A
            </span>
            开发者
          </button>
          {userMenuOpen && (
            <>
              <div className="fixed inset-0 z-40" onClick={() => setUserMenuOpen(false)} />
              <div className="glass-panel absolute right-0 top-11 z-50 w-40 overflow-hidden !rounded-xl p-1">
                <button
                  onClick={handleLogout}
                  className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm text-muted-foreground transition-colors hover:bg-white/5 hover:text-foreground"
                >
                  <LogOut className="size-3.5" />
                  退出登录
                </button>
              </div>
            </>
          )}
        </div>
      </header>

      <div className="flex">
        <aside
          className={cn(
            "sticky top-14 flex h-[calc(100vh-3.5rem)] shrink-0 flex-col overflow-y-auto border-r border-glass-border bg-surface/60 px-3 py-4 backdrop-blur-xl transition-[width] duration-200",
            collapsed ? "w-16" : "w-56"
          )}
        >
          <ul className="space-y-0.5">
            {group.items.map((item) => {
              const active = isActive(item, location.pathname)
              const Icon = item.icon
              return (
                <li key={item.to}>
                  <Link
                    to={item.to}
                    title={collapsed ? item.label : undefined}
                    className={cn(
                      "flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-white/5 hover:text-foreground",
                      collapsed && "justify-center",
                      active && "bg-primary/15 text-[#5aa6ff]"
                    )}
                  >
                    <Icon className="size-4 shrink-0" />
                    {!collapsed && <span>{item.label}</span>}
                  </Link>
                </li>
              )
            })}
          </ul>
        </aside>

        <main className="min-w-0 flex-1 px-8 py-8">
          <div className="mx-auto max-w-6xl">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
