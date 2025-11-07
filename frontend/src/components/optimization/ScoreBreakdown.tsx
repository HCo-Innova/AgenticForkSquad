import type { Proposal } from '../../services/optimizations'

export default function ScoreBreakdown({ proposal }: { proposal: Proposal }) {
  const impact = (proposal.estimated_impact || {}) as any
  const breakdown = (impact.score_breakdown || {}) as Record<string, number>
  const entries = Object.entries(breakdown)
  if (entries.length === 0) return <div>Sin desglose de puntaje</div>
  return (
    <div>
      <h3>Score Breakdown</h3>
      <ul style={{ listStyle: 'none', padding: 0, display: 'grid', gap: 6 }}>
        {entries.map(([k, v]) => (
          <li key={k} style={{ display: 'grid', gridTemplateColumns: '160px 1fr auto', gap: 8, alignItems: 'center' }}>
            <div style={{ color: '#6b7280' }}>{k}</div>
            <div style={{ background: '#f3f4f6', height: 8, borderRadius: 999 }}>
              <div style={{ width: `${Math.max(0, Math.min(100, Number(v)))}%`, height: 8, background: '#2563eb', borderRadius: 999 }} />
            </div>
            <div style={{ fontFamily: 'monospace', fontSize: 12 }}>{Number.isFinite(Number(v)) ? Number(v).toFixed(1) : '—'}</div>
          </li>
        ))}
      </ul>
    </div>
  )
}
