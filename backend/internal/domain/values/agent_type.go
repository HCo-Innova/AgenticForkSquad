package values

// AgentType defines the specific type of LLM agent used in the execution.
type AgentType string

const (
	AgentCerebro  AgentType = "cerebro"
	AgentOperativo AgentType = "operativo"
	AgentBulk      AgentType = "bulk"
)

// AgentSpecialization describes the agent's area of expertise.
type AgentSpecialization string

const (
	SpecializationPlanningQA  AgentSpecialization = "Planning/QA"
	SpecializationOperational AgentSpecialization = "SQL/Bench/Transforms"
	SpecializationBulk        AgentSpecialization = "Boilerplate/Refactors"
)

// GetSpecialization returns the specific expertise for a given agent type.
func (a AgentType) GetSpecialization() AgentSpecialization {
	switch a {
	case AgentCerebro:
		return SpecializationPlanningQA
	case AgentOperativo:
		return SpecializationOperational
	case AgentBulk:
		return SpecializationBulk
	default:
		return ""
	}
}

// ParseAgentType maps a string to the AgentType using role/model aliases.
func ParseAgentType(s string) AgentType {
	switch s {
	case "cerebro", "gemini-2.5-pro", "gemini25pro":
		return AgentCerebro
	case "operativo", "gemini-2.5-flash", "gemini25flash":
		return AgentOperativo
	case "bulk", "gemini-2.0-flash", "gemini20flash":
		return AgentBulk
	default:
		return AgentOperativo
	}
}
