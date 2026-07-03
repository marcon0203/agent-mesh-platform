import { Navigate, Route, Routes } from "react-router-dom"
import PortalLayout from "@/layouts/PortalLayout"
import ConsoleLayout from "@/layouts/ConsoleLayout"
import { ProtectedRoute } from "@/components/ProtectedRoute"
import PortalHomePage from "@/pages/PortalHomePage"
import MarketplacePage from "@/pages/MarketplacePage"
import AgentListPage from "@/pages/AgentListPage"
import AgentEditorPage from "@/pages/AgentEditorPage"
import ModelProvidersPage from "@/pages/ModelProvidersPage"
import AppPlazaPage from "@/pages/AppPlazaPage"
import WorkbenchPage from "@/pages/WorkbenchPage"
import DashboardPage from "@/pages/DashboardPage"

export default function App() {
  return (
    <Routes>
      {/* Portal：对外宣传 + 能力市场浏览，不需要登录 */}
      <Route element={<PortalLayout />}>
        <Route path="/" element={<PortalHomePage />} />
        <Route path="/marketplace" element={<MarketplacePage />} />
      </Route>

      {/* Console：登录后的管理后台。未登录时 ProtectedRoute 会照常渲染这些页面，
          只是叠一层模糊 + 登录弹窗（参考阿里云百炼），不整页跳去单独的登录页。 */}
      <Route element={<ProtectedRoute />}>
        <Route element={<ConsoleLayout />}>
          <Route path="/console" element={<DashboardPage />} />
          <Route path="/console/builder" element={<AgentListPage />} />
          <Route path="/console/builder/:id" element={<AgentEditorPage />} />
          <Route path="/console/model-providers" element={<ModelProvidersPage />} />
          <Route path="/console/app-plaza" element={<AppPlazaPage />} />
          <Route path="/console/workbench" element={<WorkbenchPage />} />
        </Route>
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
