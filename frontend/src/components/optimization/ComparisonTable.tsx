import type { Proposal } from '../../services/optimizations'

function pct(val?: number) {
  if (!Number.isFinite(val as number)) return '—'
  const n = Number(val)
  return `${n.toFixed(1)}%`
}

export default function ComparisonTable({ proposals, onSelect, selectedId }: { proposals: Proposal[]; onSelect?: (id: number) => void; selectedId?: number }) {
  return (
    <div>
      <h3>Comparison</h3>
      <div style={{ overflowX: 'auto' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr>
              <th style={{ textAlign: 'left', padding: 8 }}>Proposal</th>
              <th style={{ textAlign: 'left', padding: 8 }}>Agent</th>
              <th style={{ textAlign: 'left', padding: 8 }}>Estimated Speedup</th>
              <th style={{ textAlign: 'left', padding: 8 }}>Storage Overhead</th>
              <th style={{ textAlign: 'left', padding: 8 }}>Score</th>
            </tr>
          </thead>
          <tbody>
            {proposals.map((p) => {
              const impact = (p.estimated_impact || {}) as any
              const speed = Number(impact.speedup_pct)
              const storage = Number(impact.storage_overhead_pct)
              const score = Number(impact.score)
              return (
                <tr key={p.id} onClick={onSelect ? () => onSelect(p.id) : undefined} style={{ borderTop: '1px solid #eee', cursor: onSelect ? 'pointer' : 'default', background: selectedId === p.id ? '#f0f9ff' : undefined }}>
                  <td style={{ padding: 8 }}>#{p.id}</td>
                  <td style={{ padding: 8 }}>{p.proposal_type}</td>
                  <td style={{ padding: 8 }}>
                    <Bar value={speed} color="#16a34a" /> {pct(speed)}
                  </td>
                  <td style={{ padding: 8 }}>
                    <Bar value={storage} color="#d97706" /> {pct(storage)}
                  </td>
                  <td style={{ padding: 8 }}>
                    <Bar value={score} color="#2563eb" max={100} /> {Number.isFinite(score) ? score.toFixed(1) : '—'}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function Bar({ value, color, max = 100 }: { value?: number; color: string; max?: number }) {
  const v = Number.isFinite(value as number) ? Math.max(0, Math.min(100, (Number(value) / max) * 100)) : 0
  return (
    <div style={{ background: '#f3f4f6', height: 8, borderRadius: 999, width: 160, display: 'inline-block', verticalAlign: 'middle' }}>
      <div style={{ width: `${v}%`, height: 8, background: color, borderRadius: 999 }} />
    </div>
  )
}
