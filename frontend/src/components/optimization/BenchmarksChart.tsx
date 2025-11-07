import type { Benchmark } from '../../services/optimizations'

function ms(val?: number) {
  if (!Number.isFinite(val as number)) return '—'
  return `${Number(val).toFixed(0)} ms`
}

export default function BenchmarksChart({ items }: { items: Benchmark[] }) {
  if (!items || items.length === 0) return <div>Sin benchmarks</div>
  const max = Math.max(...items.map((i) => i.execution_time_ms)) || 1
  return (
    <div>
      <h3>Benchmarks</h3>
      <ul style={{ listStyle: 'none', padding: 0, display: 'grid', gap: 8 }}>
        {items.map((b) => (
          <li key={b.id} style={{ display: 'grid', gridTemplateColumns: '1fr auto', gap: 8, alignItems: 'center' }}>
            <div>
              <div style={{ fontSize: 12, color: '#6b7280' }}>{b.query_name}</div>
              <div style={{ background: '#f3f4f6', height: 10, borderRadius: 999 }}>
                <div style={{ width: `${(b.execution_time_ms / max) * 100}%`, height: 10, background: '#6b7280', borderRadius: 999 }} />
              </div>
            </div>
            <div style={{ fontFamily: 'monospace', fontSize: 12 }}>{ms(b.execution_time_ms)}</div>
          </li>
        ))}
      </ul>
    </div>
  )
}
