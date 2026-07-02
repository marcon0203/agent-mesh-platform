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

// Portal 首页 Hero 区的放大版星座图，带呼吸/流动动画（见 index.css 的
// .node-core/.node-orbit/.edge-flow），呼应 Tool/Skill/Agent 互联编排的定位。
export function ConstellationGraphic({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 300 300" className={className}>
      <defs>
        <radialGradient id="coreGrad" cx="50%" cy="50%" r="50%">
          <stop offset="0%" stopColor="#5AA6FF" />
          <stop offset="100%" stopColor="#2F6FED" />
        </radialGradient>
      </defs>
      <line className="edge-flow" x1="150" y1="150" x2="70" y2="80" stroke="#5AA6FF" strokeWidth="1" opacity="0.45" />
      <line className="edge-flow" x1="150" y1="150" x2="230" y2="80" stroke="#5AA6FF" strokeWidth="1" opacity="0.45" />
      <line className="edge-flow" x1="150" y1="150" x2="60" y2="200" stroke="#3FE0D0" strokeWidth="1" opacity="0.4" />
      <line className="edge-flow" x1="150" y1="150" x2="240" y2="200" stroke="#3FE0D0" strokeWidth="1" opacity="0.4" />
      <line className="edge-flow" x1="150" y1="150" x2="150" y2="240" stroke="#B48CFF" strokeWidth="1" opacity="0.4" />
      <line x1="70" y1="80" x2="230" y2="80" stroke="#5AA6FF" strokeWidth="0.6" opacity="0.2" />
      <line x1="60" y1="200" x2="240" y2="200" stroke="#3FE0D0" strokeWidth="0.6" opacity="0.2" />

      <circle className="node-core" cx="150" cy="150" r="8" fill="url(#coreGrad)" />
      <circle className="node-orbit" cx="70" cy="80" r="5" fill="#5AA6FF" />
      <circle className="node-orbit" cx="230" cy="80" r="5" fill="#5AA6FF" />
      <circle className="node-orbit" cx="60" cy="200" r="5" fill="#3FE0D0" />
      <circle className="node-orbit" cx="240" cy="200" r="5" fill="#3FE0D0" />
      <circle className="node-orbit" cx="150" cy="240" r="5" fill="#B48CFF" />
    </svg>
  )
}
