package skills

func init() {
	// Original 5
	Register(HealthcareSkill())
	Register(FintechSkill())
	Register(RetailSkill())
	Register(LegalSkill())
	Register(EducationSkill())
	// New 8
	Register(AgroSkill())
	Register(ConstructionSkill())
	Register(LogisticsSkill())
	Register(SaaSSkill())
	Register(EnergySkill())
	Register(ManufacturingSkill())
	Register(MediaSkill())
	Register(TourismSkill())
}
