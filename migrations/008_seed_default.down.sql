-- Only the group seeded by the up migration is removed; storage strategies are
-- created by the application bootstrap, not by this migration.
DELETE FROM groups WHERE id = 1;
