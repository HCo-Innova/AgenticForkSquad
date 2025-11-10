import type { Proposal } from '../../services/optimizations'

function pct(val?: number) {
  if (!Number.isFinite(val as number)) return '—'
  const n = Number(val)
  return `${n.toFixed(1)}%`
}

export default function ComparisonTable({ proposals, onSelect, selectedId }: { proposals: Proposal[]; onSelect?: (id: number) => void; selectedId?: number }) {
  if (!proposals.length) {
    return (
      <div className="text-center py-8 text-gray-500">
        <p>Sin desglose de puntaje</p>
        <p className="text-sm mt-2">Los puntajes aparecerán cuando el consenso se complete.</p>
      </div>
    )
  }

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
              const breakdown = (impact.score_breakdown || {}) as Record<string, number>
              
              // Usar performance score del consenso si existe, sino intentar estimated
              const speed = breakdown.performance ?? Number(impact.query_time_improvement || impact.speedup_pct || 0)
              
              // Usar storage score (invertido: 100=bajo overhead) o MB directo
              const storageMB = Number(impact.storage_overhead_mb || 0)
              const storageScore = breakdown.storage ?? 0
              
              // Score total del consenso
              const score = breakdown.weighted_total ?? Number(impact.score || 0)
              
              return (
                <tr key={p.id} onClick={onSelect ? () => onSelect(p.id) : undefined} style={{ borderTop: '1px solid #eee', cursor: onSelect ? 'pointer' : 'default', background: selectedId === p.id ? '#f0f9ff' : undefined }}>
                  <td style={{ padding: 8 }}>#{p.id}</td>
                  <td style={{ padding: 8 }}>{p.proposal_type}</td>
                  <td style={{ padding: 8 }}>
                    {speed > 0 ? (
                      <>
                        <Bar value={speed} color="#16a34a" /> {pct(speed)}
                      </>
                    ) : (
                      <span className="text-gray-400">—</span>
                    )}
                  </td>
                  <td style={{ padding: 8 }}>
                    {storageMB > 0 ? (
                      <>
                        <Bar value={storageMB} color="#d97706" max={50} /> {storageMB.toFixed(1)} MB
                      </>
                    ) : storageScore > 0 ? (
                      <>
                        <Bar value={100 - storageScore} color="#d97706" /> {(100 - storageScore).toFixed(1)}%
                      </>
                    ) : (
                      <span className="text-gray-400">—</span>
                    )}
                  </td>
                  <td style={{ padding: 8 }}>
                    {score > 0 ? (
                      <>
                        <Bar value={score} color="#2563eb" max={100} /> {score.toFixed(1)}
                      </>
                    ) : (
                      <span className="text-gray-400">—</span>
                    )}
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
