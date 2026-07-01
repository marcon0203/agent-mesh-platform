interface PageHeaderProps {
  eyebrow: string
  title: string
  description?: string
}

// 统一各页面顶部的"分类胶囊 + 大标题 + 说明文字"结构，避免每个页面各自手写一份，
// 也让 index.css 里的 .eyebrow-pill 视觉升级能一次性覆盖所有页面。
export function PageHeader({ eyebrow, title, description }: PageHeaderProps) {
  return (
    <div className="mb-10">
      <span className="eyebrow-pill">{eyebrow}</span>
      <h1 className="mt-4 font-display text-4xl font-bold tracking-tight text-foreground">{title}</h1>
      {description && (
        <p className="mt-3 max-w-xl text-sm leading-relaxed text-muted-foreground">{description}</p>
      )}
    </div>
  )
}
