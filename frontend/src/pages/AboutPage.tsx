export default function AboutPage() {
  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <div className="bg-white rounded-lg shadow-xl overflow-hidden">
        <div className="bg-gradient-to-r from-blue-600 to-indigo-600 px-8 py-12 text-white">
          <h1 className="text-4xl font-bold mb-4">🚀 Agentic Fork Squad</h1>
          <p className="text-xl opacity-90">Multi-Agent Database Optimization System</p>
        </div>

        <div className="px-8 py-8 space-y-8">
          <section>
            <h2 className="text-2xl font-bold text-gray-900 mb-4">The Problem</h2>
            <p className="text-gray-700 leading-relaxed">
              Database administrators face a critical challenge: testing optimizations directly in production 
              (risky) or creating full database copies for testing (slow and expensive). Traditional approaches 
              leave DBAs without safe, fast ways to validate optimization strategies.
            </p>
          </section>

          <section>
            <h2 className="text-2xl font-bold text-gray-900 mb-4">The Solution</h2>
            <p className="text-gray-700 leading-relaxed mb-4">
              AFS leverages Tiger Cloud's zero-copy database forks to enable multiple AI agents to propose 
              and benchmark different optimizations in parallel. A consensus engine selects the best solution 
              based on real performance metrics.
            </p>
            <div className="bg-blue-50 rounded-lg p-6">
              <h3 className="font-semibold text-gray-900 mb-3">How it works:</h3>
              <ol className="space-y-2 text-gray-700">
                <li>1. User submits a slow SQL query</li>
                <li>2. System assigns specialized AI agents (Vertex AI: gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash)</li>
                <li>3. Each agent creates an isolated database fork (Tiger Cloud)</li>
                <li>4. Agents propose different optimizations (indexes, partitioning, materialized views)</li>
                <li>5. Each proposal is benchmarked in its fork</li>
                <li>6. Consensus engine ranks proposals by performance, storage, complexity, and risk</li>
                <li>7. Winning optimization is applied to main database</li>
                <li>8. Forks are cleaned up instantly (zero-copy)</li>
              </ol>
            </div>
          </section>

          <section>
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Key Features</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FeatureCard 
                icon="🤖"
                title="Multi-Agent Intelligence"
                description="3 specialized AI agents work in parallel, each bringing unique optimization strategies"
              />
              <FeatureCard 
                icon="⚡"
                title="Zero-Copy Forks"
                description="Instant database forks (<10s) regardless of size, using Tiger Cloud's Fluid Storage"
              />
              <FeatureCard 
                icon="📊"
                title="Objective Benchmarking"
                description="Real performance metrics, not estimates. Multiple test queries per proposal"
              />
              <FeatureCard 
                icon="🎯"
                title="Intelligent Consensus"
                description="Multi-criteria scoring: 50% performance, 20% storage, 20% complexity, 10% risk"
              />
              <FeatureCard 
                icon="🔄"
                title="Real-Time Updates"
                description="WebSocket integration provides live progress updates as agents work"
              />
              <FeatureCard 
                icon="🔎"
                title="Hybrid Search"
                description="Full-text + vector similarity search for intelligent query pattern matching"
              />
            </div>
          </section>

          <section>
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Technology Stack</h2>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <TechBadge name="Go 1.25+" />
              <TechBadge name="React 19" />
              <TechBadge name="PostgreSQL 16" />
              <TechBadge name="Tiger Cloud" />
              <TechBadge name="Vertex AI" />
              <TechBadge name="Fiber v2" />
              <TechBadge name="TypeScript 5" />
              <TechBadge name="Tailwind CSS" />
            </div>
          </section>

          <section>
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Metrics</h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <StatCard label="Fork Creation" value="<10s" subtext="Zero-copy technology" />
              <StatCard label="Task Completion" value="4-5 min" subtext="End-to-end workflow" />
              <StatCard label="Storage Efficiency" value="3x" subtext="vs traditional copies" />
            </div>
          </section>

          <section className="border-t pt-8">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Project Info</h2>
            <div className="space-y-2 text-gray-700">
              <p><strong>Challenge:</strong> Tiger Cloud Challenge 2024</p>
              <p><strong>Repository:</strong> <a href="https://github.com/HCo-Innova/AgenticForkSquad" className="text-blue-600 hover:underline">github.com/HCo-Innova/AgenticForkSquad</a></p>
              <p><strong>License:</strong> AGPL-3.0</p>
            </div>
          </section>
        </div>
      </div>
    </div>
  )
}

function FeatureCard({ icon, title, description }: { icon: string; title: string; description: string }) {
  return (
    <div className="bg-gray-50 rounded-lg p-4 hover:shadow-md transition-shadow">
      <div className="text-3xl mb-2">{icon}</div>
      <h3 className="font-semibold text-gray-900 mb-1">{title}</h3>
      <p className="text-sm text-gray-600">{description}</p>
    </div>
  )
}

function TechBadge({ name }: { name: string }) {
  return (
    <div className="bg-blue-100 text-blue-800 px-3 py-2 rounded-lg text-center font-medium text-sm">
      {name}
    </div>
  )
}

function StatCard({ label, value, subtext }: { label: string; value: string; subtext: string }) {
  return (
    <div className="bg-gradient-to-br from-blue-50 to-indigo-50 rounded-lg p-4 text-center">
      <p className="text-sm text-gray-600 mb-1">{label}</p>
      <p className="text-3xl font-bold text-blue-600 mb-1">{value}</p>
      <p className="text-xs text-gray-500">{subtext}</p>
    </div>
  )
}
