-- Pprof 采样记录支持多种运行时 Profile。

ALTER TABLE sys_pprof_records
  ADD COLUMN profile_type varchar(32) NOT NULL DEFAULT 'cpu' COMMENT 'cpu/heap/allocs/goroutine/mutex/block/threadcreate/trace' AFTER deleted_at,
  ADD KEY idx_sys_pprof_records_profile_type (profile_type);
