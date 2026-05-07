interface Props {
  data: number[]
  color?: string
  height?: number
}

export function Sparkline({ data, color = '#ff5722', height = 28 }: Props) {
  if (!data.length) return null
  const w = 100
  const max = Math.max(...data)
  const min = Math.min(...data)
  const range = max - min || 1
  const pts = data
    .map((v, i) => `${(i / (data.length - 1)) * w},${height - ((v - min) / range) * (height - 4) - 2}`)
    .join(' ')
  return (
    <svg
      width="100%"
      height={height}
      viewBox={`0 0 ${w} ${height}`}
      preserveAspectRatio="none"
      style={{ display: 'block' }}
    >
      <polyline fill="none" stroke={color} strokeWidth="1.2" points={pts} vectorEffect="non-scaling-stroke" />
      <polyline fill={color} fillOpacity="0.08" stroke="none" points={`0,${height} ${pts} ${w},${height}`} />
    </svg>
  )
}
