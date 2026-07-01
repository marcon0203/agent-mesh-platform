# Frontend（React + TypeScript + shadcn/ui + Tailwind v4）

## 技术选型

- Vite + React 19 + TypeScript
- Tailwind CSS v4（`@tailwindcss/vite` 插件，无需 `tailwind.config.js`，主题变量直接写在 `src/index.css`）
- shadcn/ui 风格组件（手动维护于 `src/components/ui`，`components.json` 保留标准配置，
  后续可用 `npx shadcn@latest add <component>` 补充更多组件）
- react-router-dom 做客户端路由

配色沿用 `docs/design/portal-reference.html` 的科技蓝毛玻璃视觉（深空背景 + 毛玻璃卡片 + 蓝/青/紫点缀），
主题变量集中在 `src/index.css` 的 `:root` 与 `@theme inline` 中。

## 目录结构

```
src/
├── api/            与 gateway-service 对接的请求封装
├── components/ui/  shadcn 风格基础组件（button/card/badge/input）
├── data/           页面用的示例数据（后续替换为真实接口）
├── layouts/         全局导航布局
├── pages/          四个核心页面：
│                     MarketplacePage   能力市场（对应产品规格第二章）
│                     AgentBuilderPage  能力装配式构建（对应产品规格第四章）
│                     WorkbenchPage     对话调试通道（对应产品规格第六章）
│                     DashboardPage     用量与 Agent 概览
├── types/          领域类型定义
└── index.css       Tailwind v4 主题变量
```

## 本地开发

```bash
npm install
npm run dev        # 默认 http://localhost:5173，/api 请求代理到 gateway-service:8080
npm run build       # 生产构建
```

## 当前状态

页面目前使用 `src/data/mock-capabilities.ts` 里的示例数据渲染，`src/api/client.ts` 已经搭好
请求封装但尚未接入真实 gateway-service 响应，标记为 `TODO` 的位置需要对接后端联调。
