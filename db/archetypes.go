package db

import (
	"database/sql"
	"time"
)

// ─── Archetypes ───────────────────────────────────────────────────────────────

// ArchetypeRow mirrors a row in the archetypes table.
type ArchetypeRow struct {
	ID           string  `json:"id"`
	CompanyID    string  `json:"company_id,omitempty"` // empty = built-in
	Name         string  `json:"name"`
	AgentType    string  `json:"agent_type"` // conformist | disruptor
	Description  string  `json:"description"`
	MemoryWeight float64 `json:"memory_weight"` // calibration multiplier 0.3–2.0
	IsActive     bool    `json:"is_active"`
	CreatedAt    int64   `json:"created_at"`
	UpdatedAt    int64   `json:"updated_at"`
}

// CreateArchetype inserts a new custom archetype (company override).
// Built-in archetypes (company_id = '') are never modified through this path.
func (d *DB) CreateArchetype(a *ArchetypeRow) error {
	now := time.Now().Unix()
	_, err := d.Exec(`
		INSERT INTO archetypes
			(id, company_id, name, agent_type, description, memory_weight, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, a.ID, a.CompanyID, a.Name, a.AgentType, a.Description, a.MemoryWeight, boolToInt(a.IsActive), now, now)
	return err
}

// UpdateArchetype updates a custom archetype by ID.
// Only non-empty fields are updated; built-in archetypes (company_id = '') are protected.
func (d *DB) UpdateArchetype(id string, name, description string, memoryWeight float64, isActive bool) error {
	_, err := d.Exec(`
		UPDATE archetypes
		SET name = ?, description = ?, memory_weight = ?, is_active = ?, updated_at = ?
		WHERE id = ? AND company_id != ''
	`, name, description, memoryWeight, boolToInt(isActive), time.Now().Unix(), id)
	return err
}

// GetArchetype returns a single archetype by ID.
func (d *DB) GetArchetype(id string) (*ArchetypeRow, error) {
	var a ArchetypeRow
	var isActive int
	err := d.QueryRow(`
		SELECT id, company_id, name, agent_type, description, memory_weight, is_active, created_at, updated_at
		FROM archetypes WHERE id = ?
	`, id).Scan(&a.ID, &a.CompanyID, &a.Name, &a.AgentType, &a.Description,
		&a.MemoryWeight, &isActive, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	a.IsActive = isActive == 1
	return &a, nil
}

// ListArchetypes returns all archetypes, optionally filtered by company_id.
// Pass empty string to get all archetypes (built-in + custom).
func (d *DB) ListArchetypes(companyID string) ([]ArchetypeRow, error) {
	var rows *sql.Rows
	var err error
	if companyID == "" {
		rows, err = d.Query(`
			SELECT id, company_id, name, agent_type, description, memory_weight, is_active, created_at, updated_at
			FROM archetypes ORDER BY agent_type ASC, name ASC
		`)
	} else {
		rows, err = d.Query(`
			SELECT id, company_id, name, agent_type, description, memory_weight, is_active, created_at, updated_at
			FROM archetypes WHERE company_id = ? OR company_id = ''
			ORDER BY agent_type ASC, name ASC
		`, companyID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []ArchetypeRow
	for rows.Next() {
		var a ArchetypeRow
		var isActive int
		if err := rows.Scan(&a.ID, &a.CompanyID, &a.Name, &a.AgentType, &a.Description,
			&a.MemoryWeight, &isActive, &a.CreatedAt, &a.UpdatedAt); err != nil {
			continue
		}
		a.IsActive = isActive == 1
		result = append(result, a)
	}
	return result, nil
}

// DeleteArchetype removes a custom archetype by ID.
// Built-in archetypes (company_id = '') cannot be deleted.
func (d *DB) DeleteArchetype(id string) error {
	_, err := d.Exec(`DELETE FROM archetypes WHERE id = ? AND company_id != ''`, id)
	return err
}

// ─── Custom Rules ─────────────────────────────────────────────────────────────

// CustomRuleRow mirrors a row in the custom_rules table.
type CustomRuleRow struct {
	ID          string  `json:"id"`
	CompanyID   string  `json:"company_id"`
	Description string  `json:"description"`
	Domain      string  `json:"domain"`
	Stability   float64 `json:"stability"` // 0.0 (fragile) to 1.0 (immutable)
	IsActive    bool    `json:"is_active"`
	CreatedAt   int64   `json:"created_at"`
	UpdatedAt   int64   `json:"updated_at"`
}

// CreateCustomRule inserts a new company-specific world rule.
func (d *DB) CreateCustomRule(r *CustomRuleRow) error {
	now := time.Now().Unix()
	_, err := d.Exec(`
		INSERT INTO custom_rules
			(id, company_id, description, domain, stability, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, r.ID, r.CompanyID, r.Description, r.Domain, r.Stability, boolToInt(r.IsActive), now, now)
	return err
}

// UpdateCustomRule updates a custom rule by ID.
func (d *DB) UpdateCustomRule(id string, description, domain string, stability float64, isActive bool) error {
	_, err := d.Exec(`
		UPDATE custom_rules
		SET description = ?, domain = ?, stability = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`, description, domain, stability, boolToInt(isActive), time.Now().Unix(), id)
	return err
}

// GetCustomRule returns a single custom rule by ID.
func (d *DB) GetCustomRule(id string) (*CustomRuleRow, error) {
	var r CustomRuleRow
	var isActive int
	err := d.QueryRow(`
		SELECT id, company_id, description, domain, stability, is_active, created_at, updated_at
		FROM custom_rules WHERE id = ?
	`, id).Scan(&r.ID, &r.CompanyID, &r.Description, &r.Domain,
		&r.Stability, &isActive, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	r.IsActive = isActive == 1
	return &r, nil
}

// ListCustomRules returns all custom rules for a company.
func (d *DB) ListCustomRules(companyID string) ([]CustomRuleRow, error) {
	rows, err := d.Query(`
		SELECT id, company_id, description, domain, stability, is_active, created_at, updated_at
		FROM custom_rules WHERE company_id = ?
		ORDER BY domain ASC, created_at ASC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []CustomRuleRow
	for rows.Next() {
		var r CustomRuleRow
		var isActive int
		if err := rows.Scan(&r.ID, &r.CompanyID, &r.Description, &r.Domain,
			&r.Stability, &isActive, &r.CreatedAt, &r.UpdatedAt); err != nil {
			continue
		}
		r.IsActive = isActive == 1
		result = append(result, r)
	}
	return result, nil
}

// DeleteCustomRule removes a custom rule by ID.
func (d *DB) DeleteCustomRule(id string) error {
	_, err := d.Exec(`DELETE FROM custom_rules WHERE id = ?`, id)
	return err
}
