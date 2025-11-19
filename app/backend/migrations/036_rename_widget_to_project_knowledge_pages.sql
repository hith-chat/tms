-- +goose Up
-- +goose StatementBegin

-- Step 1: Drop indexes that reference widget_id
DROP INDEX IF EXISTS idx_widget_knowledge_pages_widget;
DROP INDEX IF EXISTS idx_widget_knowledge_pages_project_widget;

-- Step 2: Drop the unique constraint on (widget_id, page_id)
ALTER TABLE widget_knowledge_pages DROP CONSTRAINT IF EXISTS widget_knowledge_pages_widget_id_page_id_key;

-- Step 3: Drop the widget_id column
ALTER TABLE widget_knowledge_pages DROP COLUMN widget_id;

-- Step 4: Rename the table
ALTER TABLE widget_knowledge_pages RENAME TO project_knowledge_pages;

-- Step 5: Add unique constraint on (project_id, page_id)
ALTER TABLE project_knowledge_pages ADD CONSTRAINT project_knowledge_pages_project_id_page_id_key UNIQUE (project_id, page_id);

-- Step 6: Rename remaining indexes for consistency
ALTER INDEX IF EXISTS idx_widget_knowledge_pages_page RENAME TO idx_project_knowledge_pages_page;
ALTER INDEX IF EXISTS idx_widget_knowledge_pages_tenant RENAME TO idx_project_knowledge_pages_tenant;
ALTER INDEX IF EXISTS idx_widget_knowledge_pages_project RENAME TO idx_project_knowledge_pages_project;
ALTER INDEX IF EXISTS idx_widget_knowledge_pages_created RENAME TO idx_project_knowledge_pages_created;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Reverse the migration

-- Step 1: Rename indexes back
ALTER INDEX IF EXISTS idx_project_knowledge_pages_page RENAME TO idx_widget_knowledge_pages_page;
ALTER INDEX IF EXISTS idx_project_knowledge_pages_tenant RENAME TO idx_widget_knowledge_pages_tenant;
ALTER INDEX IF EXISTS idx_project_knowledge_pages_project RENAME TO idx_widget_knowledge_pages_project;
ALTER INDEX IF EXISTS idx_project_knowledge_pages_created RENAME TO idx_widget_knowledge_pages_created;

-- Step 2: Drop the new unique constraint
ALTER TABLE project_knowledge_pages DROP CONSTRAINT IF EXISTS project_knowledge_pages_project_id_page_id_key;

-- Step 3: Rename table back
ALTER TABLE project_knowledge_pages RENAME TO widget_knowledge_pages;

-- Step 4: Add widget_id column back (will be NULL for existing rows)
ALTER TABLE widget_knowledge_pages ADD COLUMN widget_id UUID REFERENCES chat_widgets(id) ON DELETE CASCADE;

-- Step 5: Re-add the original unique constraint
ALTER TABLE widget_knowledge_pages ADD CONSTRAINT widget_knowledge_pages_widget_id_page_id_key UNIQUE (widget_id, page_id);

-- Step 6: Re-create the dropped indexes
CREATE INDEX idx_widget_knowledge_pages_widget ON widget_knowledge_pages(widget_id);
CREATE INDEX idx_widget_knowledge_pages_project_widget ON widget_knowledge_pages(project_id, widget_id);

-- +goose StatementEnd
