-- Add indexes for common query patterns on audit_logs table

-- Index for tenant-based queries (most common access pattern)
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_created ON audit_logs(tenant_id, created_at DESC);

-- Index for user activity analysis
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_tenant ON audit_logs(user_id, tenant_id, created_at DESC);

-- Index for resource-based queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id, created_at DESC);

-- Index for action-based filtering
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action, created_at DESC);

-- Composite index for time-range queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_time_range ON audit_logs(created_at DESC, tenant_id);
