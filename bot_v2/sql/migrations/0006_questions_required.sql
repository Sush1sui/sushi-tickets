-- Migration: 0006_questions_required
-- Add required_flags parallel boolean array to questions_config.
-- required_flags[i] corresponds to questions[i].

ALTER TABLE questions_config
    ADD COLUMN IF NOT EXISTS required_flags BOOLEAN[] NOT NULL DEFAULT '{}';
