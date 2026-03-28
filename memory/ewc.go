package memory

import (
	"context"
	"database/sql"
	"math"
	"time"
)

// EWC implementa Elastic Weight Consolidation.
// Protege pesos importantes de simulações anteriores enquanto aprende novas.
type EWC struct {
	db    *sql.DB
	alpha float64 // taxa de proteção [0.1, 0.9] — maior = mais conservador
}

func NewEWC(db *sql.DB, alpha float64) *EWC {
	if alpha < 0.1 {
		alpha = 0.1
	}
	if alpha > 0.9 {
		alpha = 0.9
	}
	return &EWC{db: db, alpha: alpha}
}

// FisherWeight representa a importância de um peso para simulações anteriores.
type FisherWeight struct {
	ArchetypeID string
	Domain      string  // setor/domínio onde o peso foi calibrado
	Importance  float64 // Fisher Information — quanto esse peso importa historicamente
	AnchorValue float64 // valor âncora a ser preservado
}

// ComputeFisherWeights calcula a importância de cada peso baseado
// na variância histórica — pesos estáveis = alta importância = mais protegidos.
func (e *EWC) ComputeFisherWeights(ctx context.Context, domain string) ([]FisherWeight, error) {
	rows, err := e.db.QueryContext(ctx, `
		SELECT archetype_id, accuracy_weight, sample_count
		FROM archetype_calibration
		WHERE domain = ?
		ORDER BY sample_count DESC
	`, domain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var weights []FisherWeight
	for rows.Next() {
		var id string
		var w float64
		var n int
		if err := rows.Scan(&id, &w, &n); err != nil {
			continue
		}
		// Fisher Information aproximada: mais amostras = mais importante
		// Pesos próximos de 1.0 (neutro) têm menor importância
		deviation := math.Abs(w - 1.0)
		importance := math.Min(1.0, float64(n)/50.0) * (0.5 + deviation)

		weights = append(weights, FisherWeight{
			ArchetypeID: id,
			Domain:      domain,
			Importance:  importance,
			AnchorValue: w,
		})
	}
	return weights, nil
}

// ProtectedUpdate aplica um delta de peso com proteção EWC.
// Pesos com alta importância histórica resistem mais a mudanças.
func (e *EWC) ProtectedUpdate(current, delta, importance float64) float64 {
	// proteção proporcional à importância — peso importante muda menos
	protection := e.alpha * importance
	effectiveDelta := delta * (1.0 - protection)
	newWeight := current + effectiveDelta
	return math.Max(0.3, math.Min(2.0, newWeight))
}

// ConsolidateWeights persiste os pesos protegidos após uma nova simulação.
// Chamado após Judge.Judge() para garantir que aprendizado novo
// não apague padrões bem estabelecidos.
func (e *EWC) ConsolidateWeights(ctx context.Context, targetDomain string, fishers []FisherWeight) error {
	if len(fishers) == 0 {
		return nil
	}

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, f := range fishers {
		var current float64
		var sampleCount int

		err := tx.QueryRowContext(ctx, `
			SELECT accuracy_weight, sample_count
			FROM archetype_calibration
			WHERE archetype_id = ? AND domain = ?
		`, f.ArchetypeID, targetDomain).Scan(&current, &sampleCount)

		if err == sql.ErrNoRows {
			// novo domínio — inicializa ancorado no valor histórico ponderado
			anchoredStart := 1.0 + (f.AnchorValue-1.0)*f.Importance*e.alpha
			anchoredStart = math.Max(0.3, math.Min(2.0, anchoredStart))
			_, err = tx.ExecContext(ctx, `
				INSERT INTO archetype_calibration
				(archetype_id, domain, accuracy_weight, sample_count, updated_at)
				VALUES (?, ?, ?, 1, ?)
			`, f.ArchetypeID, targetDomain, anchoredStart, time.Now().Unix())
			if err != nil {
				continue
			}
		} else if err == nil {
			// domínio existente — aplica proteção EWC sobre peso atual
			delta := f.AnchorValue - current
			protected := e.ProtectedUpdate(current, delta*0.1, f.Importance)
			_, err = tx.ExecContext(ctx, `
				UPDATE archetype_calibration
				SET accuracy_weight = ?, updated_at = ?
				WHERE archetype_id = ? AND domain = ?
			`, protected, time.Now().Unix(), f.ArchetypeID, targetDomain)
			if err != nil {
				continue
			}
		}
	}

	return tx.Commit()
}
