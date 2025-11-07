package usecases

// WebSocket event types aligned with API spec (08-API-SPECIFICATION.md)
const (
	EventTaskCreated          = "task_created"
	EventAgentsAssigned       = "agents_assigned"
	EventForkCreated          = "fork_created"
	EventAnalysisCompleted    = "analysis_completed"
	EventProposalSubmitted    = "proposal_submitted"
	EventBenchmarkCompleted   = "benchmark_completed"
	EventConsensusReached     = "consensus_reached"
	EventOptimizationApplied  = "optimization_applied"
	EventTaskCompleted        = "task_completed"
	EventTaskFailed           = "task_failed"
	EventConnectionEstablished = "connection_established"
)
