-- Pprof 采样记录元数据。原始采样和火焰图内容保存在文件系统。

CREATE TABLE IF NOT EXISTS sys_pprof_records (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  created_at datetime(3) DEFAULT NULL,
  updated_at datetime(3) DEFAULT NULL,
  deleted_at datetime(3) DEFAULT NULL,
  status varchar(32) NOT NULL COMMENT 'waiting/collecting/generating/completed/failed',
  duration_seconds int NOT NULL,
  started_at datetime(3) DEFAULT NULL,
  finished_at datetime(3) DEFAULT NULL,
  raw_file varchar(500) DEFAULT NULL,
  flame_file varchar(500) DEFAULT NULL,
  sample_total bigint NOT NULL DEFAULT 0,
  sample_unit varchar(32) DEFAULT NULL,
  error_message text,
  created_by bigint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  KEY idx_sys_pprof_records_status (status),
  KEY idx_sys_pprof_records_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO sys_apis (created_at, updated_at, deleted_at, path, description, api_group, method)
SELECT NOW(3), NOW(3), NULL, source.path, source.description, 'Pprof 性能分析', 'GET'
FROM (
  SELECT '/system/pprof/records' AS path, '获取 Pprof 采样记录' AS description
  UNION ALL SELECT '/system/pprof/record', '查看 Pprof 采样记录'
) AS source
WHERE NOT EXISTS (
  SELECT 1 FROM sys_apis
  WHERE sys_apis.path = source.path AND sys_apis.method = 'GET' AND sys_apis.deleted_at IS NULL
);

INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
SELECT 'p', '888', source.path, 'GET', '', '', ''
FROM (
  SELECT '/system/pprof/records' AS path
  UNION ALL SELECT '/system/pprof/record'
) AS source
WHERE NOT EXISTS (
  SELECT 1 FROM casbin_rule
  WHERE casbin_rule.ptype = 'p' AND casbin_rule.v0 = '888'
    AND casbin_rule.v1 = source.path AND casbin_rule.v2 = 'GET'
);
