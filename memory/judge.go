package memory

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"
)

// Verdict representa o resultado do JUDGE para um agente numa simulação.
type Verdict string

const (
	VerdictHit     Verdict = "hit"     // agente previu corretamente
	VerdictMiss    Verdict = "miss"    // agente errou previsão
	VerdictPartial Verdict = "partial" // acerto parcial
)

// SimulationJudgement consolida os veredictos de uma simulação completa.
type SimulationJudgement struct {
	SimulationID  string
	CompanyID     string
	Verdicts      map[string]Verdict // agentID → veredicto
	FractureHit   bool               // fracture point ocorreu como previsto
	RoundAccuracy float64            // quão próximo do round esperado
	JudgedAt      time.Time
}

// Judge avalia o resultado real de uma simulação contra o esperado
// e ajusta os pesos dos agentes de forma direcionada.
// É seguro chamar em goroutine — usa sua própria transação.
func (c *Calibrator) Judge(ctx context.Context, judgement SimulationJudgement) error {
	// resolve domain from simulation_jobs — same pattern as recalibrateForSimulation
	var domain string
	if err := c.db.QueryRowContext(ctx,
		`SELECT department FROM simulation_jobs WHERE id = ?`, judgement.SimulationID,
	).Scan(&domain); err != nil {
		domain = "market"
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("judge: begin tx: %w", err)
	}
	defer tx.Rollback()

	for agentID, verdict := range judgement.Verdicts {
		delta := verdictToDelta(verdict, judgement.RoundAccuracy)
		if err := c.applyJudgeDelta(ctx, tx, domain, agentID, delta); err != nil {
			// não-fatal — continua para os outros agentes
			continue
		}
	}

	// registra no grafo causal se fracture ocorreu como previsto
	if judgement.FractureHit {
		c.Graph.RecordCausality(
			judgement.CompanyID,
			"fracture_predicted",
			"fracture_confirmed",
		)
	}

	return tx.Commit()
}

// verdictToDelta converte veredicto em delta de peso [-1, 1].
// RoundAccuracy (0-1) pondera o quanto o timing foi preciso.
func verdictToDelta(v Verdict, roundAccuracy float64) float64 {
	acc := math.Max(0, math.Min(1, roundAccuracy))
	switch v {
	case VerdictHit:
		return 0.3 + (0.4 * acc) // +0.3 a +0.7 dependendo da precisão do round
	case VerdictPartial:
		return 0.1 * acc // pequeno ganho se parcialmente correto
	case VerdictMiss:
		return -0.2 * (1 - acc) // penalidade proporcional ao erro
	default:
		return 0
	}
}

// applyJudgeDelta aplica o delta ao AccuracyWeight do agente via EMA direcional.
func (c *Calibrator) applyJudgeDelta(ctx context.Context, tx *sql.Tx, domain, agentID string, delta float64) error {
	var current float64
	var sampleCount int

	err := tx.QueryRowContext(ctx, `
		SELECT accuracy_weight, sample_count
		FROM archetype_calibration
		WHERE archetype_id = ? AND domain = ?
	`, agentID, domain).Scan(&current, &sampleCount)

	if err == sql.ErrNoRows {
		// primeira vez — insere com peso neutro
		current = 1.0
		sampleCount = 0
	} else if err != nil {
		return err
	}

	// EMA direcional: alpha maior para hits (aprende rápido de acertos)
	// alpha menor para misses (não pune demais por um erro)
	alpha := 1.0 / float64(sampleCount+1)
	if delta > 0 {
		alpha = math.Min(alpha*1.5, 0.3) // aprende mais rápido de acertos
	} else {
		alpha = math.Min(alpha, 0.15) // mais conservador em punições
	}

	newWeight := current + alpha*delta
	newWeight = math.Max(0.3, math.Min(2.0, newWeight)) // clamp [0.3, 2.0]
	sampleCount++

	_, err = tx.ExecContext(ctx, `
		INSERT INTO archetype_calibration (archetype_id, domain, accuracy_weight, sample_count, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(archetype_id, domain) DO UPDATE SET
			accuracy_weight = excluded.accuracy_weight,
			sample_count    = excluded.sample_count,
			updated_at      = excluded.updated_at
	`, agentID, domain, newWeight, sampleCount, time.Now().Unix())

	return err
}

// BuildJudgement constrói um SimulationJudgement comparando
// fracture points previstos vs ocorridos numa simulação finalizada.
// Chamado pelo handler após Finalize().
func BuildJudgement(
	simulationID string,
	companyID string,
	predictedFractureRound int,
	actualFractureRound int,
	agentParticipation map[string]bool, // agentID → participou?
) SimulationJudgement {

	fractureHit := actualFractureRound > 0 && predictedFractureRound > 0
	var roundAccuracy float64
	if fractureHit {
		diff := math.Abs(float64(actualFractureRound - predictedFractureRound))
		roundAccuracy = math.Max(0, 1.0-(diff/10.0)) // 10 rounds de tolerância
	}

	verdicts := make(map[string]Verdict, len(agentParticipation))
	for agentID, participated := range agentParticipation {
		if !participated {
			continue
		}
		if fractureHit && roundAccuracy > 0.7 {
			verdicts[agentID] = VerdictHit
		} else if fractureHit && roundAccuracy > 0.3 {
			verdicts[agentID] = VerdictPartial
		} else {
			verdicts[agentID] = VerdictMiss
		}
	}

	return SimulationJudgement{
		SimulationID:  simulationID,
		CompanyID:     companyID,
		Verdicts:      verdicts,
		FractureHit:   fractureHit,
		RoundAccuracy: roundAccuracy,
		JudgedAt:      time.Now(),
	}
}
