// 星座连线图标，取自 docs/design/portal-reference.html 的门户视觉参考，
// 呼应"能力互联成网"的产品定位（枢络 AgentMesh）。
export function LogoMark({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 28 28" fill="none" className={className}>
      <line x1="14" y1="6" x2="5" y2="20" stroke="#5AA6FF" strokeWidth="1" opacity="0.6" />
      <line x1="14" y1="6" x2="23" y2="20" stroke="#5AA6FF" strokeWidth="1" opacity="0.6" />
      <line x1="5" y1="20" x2="23" y2="20" stroke="#3FE0D0" strokeWidth="1" opacity="0.6" />
      <circle cx="14" cy="6" r="3" fill="#5AA6FF" />
      <circle cx="5" cy="20" r="3" fill="#3FE0D0" />
      <circle cx="23" cy="20" r="3" fill="#2F6FED" />
    </svg>
  )
}
