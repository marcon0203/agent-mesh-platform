import { useLocation, useNavigate, Link } from "react-router-dom"
import { Card, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { LogoMark } from "@/components/Logo"
import { useAuth } from "@/lib/auth"

// 假登录页：平台还没有真实账号体系，点一下按钮就"登录"成功，
// 用来先把 Portal / Console 的路由结构和交互跑通。
export default function LoginPage() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const from = (location.state as { from?: string } | null)?.from ?? "/console"

  const handleLogin = () => {
    login()
    navigate(from, { replace: true })
  }

  return (
    <div className="flex min-h-screen items-center justify-center px-6">
      <Card className="w-full max-w-sm p-8">
        <div className="mb-6 flex items-center gap-2.5">
          <LogoMark className="size-7 drop-shadow-[0_0_6px_rgba(90,166,255,0.5)]" />
          <span className="font-display text-lg font-semibold">
            枢络<span className="ml-1 font-mono text-xs font-normal text-muted-foreground">AgentMesh</span>
          </span>
        </div>

        <CardTitle className="mb-2">登录控制台</CardTitle>
        <CardDescription className="mb-6">
          平台账号体系还未接入，这里先用一个假登录跑通控制台的路由和交互。
        </CardDescription>

        <Button className="w-full" size="lg" onClick={handleLogin}>
          登录
        </Button>

        <Link to="/" className="mt-4 block text-center text-xs text-muted-foreground hover:text-foreground">
          ← 返回首页
        </Link>
      </Card>
    </div>
  )
}
