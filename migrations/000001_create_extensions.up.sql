--
-- Enable uuid extension
--
-- Use to primary keys and foreign keys in tables of ledger-exp database
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

--
-- Enable hstore extension
--
-- Use for metadata fields normally
CREATE EXTENSION IF NOT EXISTS "hstore";
