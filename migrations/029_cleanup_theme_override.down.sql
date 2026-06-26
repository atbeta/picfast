-- WARNING: This migration is irreversible. We cannot restore the
-- dropped theme_override values. Backup users.settings before
-- downgrading if needed.
-- (No-op: there is nothing to do on downgrade; the previous
-- up migration is a one-way data cleanup.)
SELECT 1;
