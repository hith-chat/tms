-- This script is run during PostgreSQL initialization
-- Create the database if it doesn't exist (this is handled by POSTGRES_DB env var)
-- This file ensures the database is ready for migrations

SELECT 'PostgreSQL is ready for Hith application' as message;
