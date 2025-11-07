package values

import "testing"

func TestGetSpecialization(t *testing.T) {
	if got := AgentCerebro.GetSpecialization(); got != SpecializationPlanningQA {
		t.Errorf("expected %s, got %s", SpecializationPlanningQA, got)
	}
	if got := AgentOperativo.GetSpecialization(); got != SpecializationOperational {
		t.Errorf("expected %s, got %s", SpecializationOperational, got)
	}
	if got := AgentBulk.GetSpecialization(); got != SpecializationBulk {
		t.Errorf("expected %s, got %s", SpecializationBulk, got)
	}
	if got := AgentType("unknown").GetSpecialization(); got != "" {
		t.Errorf("expected empty specialization, got %s", got)
	}
}

func TestParseAgentType_Aliases(t *testing.T) {
	cases := []struct{ in string; want AgentType }{
		{"cerebro", AgentCerebro},
		{"gemini-2.5-pro", AgentCerebro},
		{"gemini25pro", AgentCerebro},
		{"operativo", AgentOperativo},
		{"gemini-2.5-flash", AgentOperativo},
		{"gemini25flash", AgentOperativo},
		{"bulk", AgentBulk},
		{"gemini-2.0-flash", AgentBulk},
		{"gemini20flash", AgentBulk},
	}
	for _, c := range cases {
		if got := ParseAgentType(c.in); got != c.want {
			t.Errorf("ParseAgentType(%q)=%q, want %q", c.in, got, c.want)
		}
	}
}