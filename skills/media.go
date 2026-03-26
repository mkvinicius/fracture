package skills

import "github.com/fracture/fracture/engine"

// MediaSkill retorna a skill vertical de Mídia & Entretenimento.
func MediaSkill() *Skill {
	return &Skill{
		ID:          "media",
		Name:        "Mídia & Entretenimento",
		Description: "Simulação especializada para mídia, streaming, publicidade, creator economy e entretenimento.",
		Industries: []string{
			"mídia", "media", "entretenimento", "entertainment",
			"streaming", "publicidade", "advertising",
			"creator economy", "influencer", "conteúdo digital",
			"TV aberta", "rádio", "jornal", "podcast", "games", "music",
		},

		Rules: []*engine.Rule{
			{ID: "med-001", Description: "TV Globo still reaches 100M+ Brazilians — linear TV is declining but not dead", Domain: engine.DomainMarket, Stability: 0.60, IsActive: true},
			{ID: "med-002", Description: "CONAR self-regulates advertising — government regulation is minimal but growing", Domain: engine.DomainRegulation, Stability: 0.72, IsActive: true},
			{ID: "med-003", Description: "Netflix, Globoplay, Disney+, Amazon Prime compete for Brazilian subscription share", Domain: engine.DomainMarket, Stability: 0.55, IsActive: true},
			{ID: "med-004", Description: "TikTok and Instagram Reels have captured 60%+ of youth attention time", Domain: engine.DomainCulture, Stability: 0.45, IsActive: true},
			{ID: "med-005", Description: "Creator economy: 500K+ Brazilian creators monetize content professionally", Domain: engine.DomainCulture, Stability: 0.42, IsActive: true},
			{ID: "med-006", Description: "Digital advertising surpassed TV advertising in Brazil for first time in 2023", Domain: engine.DomainMarket, Stability: 0.58, IsActive: true},
			{ID: "med-007", Description: "AI-generated content is reducing content production cost 70-90%", Domain: engine.DomainTechnology, Stability: 0.30, IsActive: true},
			{ID: "med-008", Description: "Podcast market in Brazil is top 3 globally — audio growing despite video dominance", Domain: engine.DomainCulture, Stability: 0.48, IsActive: true},
			{ID: "med-009", Description: "Gaming is larger than movies and music combined globally — Brazil is top 10 market", Domain: engine.DomainCulture, Stability: 0.55, IsActive: true},
			{ID: "med-010", Description: "Ad-supported streaming (AVOD) is winning over pure subscription (SVOD)", Domain: engine.DomainMarket, Stability: 0.48, IsActive: true},
			{ID: "med-011", Description: "Intellectual property protection is weak in Brazilian digital media", Domain: engine.DomainRegulation, Stability: 0.62, IsActive: true},
			{ID: "med-012", Description: "Live commerce via TikTok/Instagram is merging entertainment and e-commerce", Domain: engine.DomainCulture, Stability: 0.35, IsActive: true},
		},

		Agents: []SkillAgent{
			{
				Name:        "João Roberto Marinho",
				Role:        "Legacy Media Champion & Globo Ecosystem Guardian",
				Traits:      []string{"TV Globo", "linear TV", "mass audience", "journalism quality", "brand safety", "Globoplay transition"},
				Goals:       []string{"preserve Globo audience leadership", "monetize content across platforms"},
				Biases:      []string{"platform disintermediation", "creator economy fragmenting attention"},
				Power:       0.88,
				IsDisruptor: false,
			},
			{
				Name:        "MrBeast",
				Role:        "Creator Economy Disruptor & Attention Economy Champion",
				Traits:      []string{"YouTube", "creator economy", "subscriber growth", "virality engineering", "content as product", "direct audience relationship"},
				Goals:       []string{"creators own the audience relationship", "eliminate media middleman"},
				Biases:      []string{"traditional media gatekeepers", "content licensing to platforms"},
				Power:       0.85,
				IsDisruptor: true,
			},
			{
				Name:        "Marshall McLuhan",
				Role:        "Media Theory Pioneer & The Medium is the Message",
				Traits:      []string{"the medium is the message", "global village", "hot vs cool media", "Understanding Media", "technological determinism", "media as extensions of man", "the Gutenberg Galaxy"},
				Goals:       []string{"understand how media reshapes human perception and society", "reveal hidden effects of communication technology"},
				Biases:      []string{"content-focused analysis that ignores medium effects", "technology adoption without understanding social transformation"},
				Power:       0.90,
				IsDisruptor: true,
			},
			{
				Name:        "David Ogilvy",
				Role:        "Advertising Father & Brand Building Pioneer",
				Traits:      []string{"Confessions of an Advertising Man", "Ogilvy on Advertising", "the consumer is not a moron she is your wife", "brand image", "long-copy advertising", "research before creativity", "the big idea"},
				Goals:       []string{"advertising that sells not just entertains", "brand built on truth and long-term consistency"},
				Biases:      []string{"creative for creative sake without sales effectiveness", "short-term performance over brand building", "attention without persuasion"},
				Power:       0.90,
				IsDisruptor: false,
			},
			{
				Name:        "Ben Thompson",
				Role:        "Stratechery & Aggregation Theory Creator",
				Traits:      []string{"Stratechery", "Aggregation Theory", "the Moat Map", "media business model analysis", "platform vs aggregator distinction", "the Daily Update", "zero distribution cost changes everything"},
				Goals:       []string{"understand how internet changes media economics", "aggregators capture value from suppliers and users simultaneously"},
				Biases:      []string{"media companies ignoring aggregation dynamics", "content strategy without distribution strategy"},
				Power:       0.88,
				IsDisruptor: true,
			},
		},

		Context: `MEDIA & ENTERTAINMENT CONTEXT FOR FRACTURE SIMULATION:
Brazil is top 5 globally in social media usage and top 10 in streaming.
Key players: Globo (TV/streaming), Band, Record (TV),
Netflix, Disney+, Amazon Prime, Globoplay (streaming),
Meta/Instagram, TikTok, YouTube (social/creator),
Spotify, Deezer (music), Casimiro, Flávio Augusto (creators).
Key dynamics: Digital advertising surpassing TV for first time,
TikTok capturing youth attention from all other platforms,
AI-generated content threatening traditional production economics,
creator economy with 500K+ professional creators in Brazil,
gaming larger than movies+music combined,
live commerce merging entertainment and e-commerce.`,

		Queries: []string{
			"mídia digital Brasil streaming publicidade disruption 2024 2025",
			"creator economy influencer monetização Brasil TikTok YouTube",
			"TV Globo streaming Globoplay competição Netflix Brasil",
			"AI conteúdo gerado inteligência artificial mídia impacto",
			"games mercado Brasil e-sports mobile gaming crescimento",
		},
	}
}
