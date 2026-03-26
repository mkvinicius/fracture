-- Add skill column to track which vertical skill was used in a simulation
ALTER TABLE simulation_jobs ADD COLUMN skill TEXT NOT NULL DEFAULT '';
