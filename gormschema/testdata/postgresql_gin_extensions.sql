CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";
CREATE TABLE "notebook_files" ("id" bigserial,"tenant_id" text NOT NULL,"scan_version" bigint NOT NULL,"file_name" text NOT NULL,PRIMARY KEY ("id"));
CREATE INDEX IF NOT EXISTS "idx_notebook_files_tenant_scan_filename_trgm" ON "notebook_files" USING gin("tenant_id","scan_version",file_name gin_trgm_ops);
