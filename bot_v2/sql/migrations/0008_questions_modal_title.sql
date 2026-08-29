-- Migration: 0008_questions_modal_title
-- Add modal_title column to questions_config with default 'Sushi Ticket Questions'.

ALTER TABLE questions_config
    ADD COLUMN IF NOT EXISTS modal_title TEXT NOT NULL DEFAULT 'Sushi Ticket Questions';
