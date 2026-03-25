package engine

// SimulationMode controls the depth and cost of a simulation run.
type SimulationMode string

const (
	ModeStandard SimulationMode = "standard" // 30 rounds, 1 ensemble run
	ModePremium  SimulationMode = "premium"  // 50 rounds, 2 ensemble runs
)

// ModeConfig holds per-mode parameters.
type ModeConfig struct {
	MaxRounds      int
	EnsembleRuns   int
	CouncilInterval int // run councils every N rounds
}

// DefaultConfigForMode returns the canonical ModeConfig for the given mode.
// Unknown modes fall back to Standard.
func DefaultConfigForMode(mode SimulationMode) ModeConfig {
	switch mode {
	case ModePremium:
		return ModeConfig{
			MaxRounds:       50,
			EnsembleRuns:    2,
			CouncilInterval: 5,
		}
	default: // ModeStandard and anything unknown
		return ModeConfig{
			MaxRounds:       30,
			EnsembleRuns:    1,
			CouncilInterval: 5,
		}
	}
}
