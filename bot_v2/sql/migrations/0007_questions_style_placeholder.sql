-- Migration: 0007_questions_style_placeholder
-- Add styles and placeholders parallel arrays to questions_config.
-- styles[i] and placeholders[i] correspond to questions[i].

ALTER TABLE questions_config
    ADD COLUMN IF NOT EXISTS styles TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS placeholders TEXT[] NOT NULL DEFAULT '{}';
