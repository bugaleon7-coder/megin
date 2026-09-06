-- 数据库表列表和字段结构查询接口权限。

INSERT INTO sys_apis (created_at, updated_at, deleted_at, path, description, api_group, method)
SELECT NOW(3), NOW(3), NULL, '/system/database-query/tables', '获取数据库表列表', '数据库查询', 'GET'
WHERE NOT EXISTS (
  SELECT 1 FROM sys_apis
  WHERE path = '/system/database-query/tables' AND method = 'GET' AND deleted_at IS NULL
);

INSERT INTO sys_apis (created_at, updated_at, deleted_at, path, description, api_group, method)
SELECT NOW(3), NOW(3), NULL, '/system/database-query/table-structure', '获取数据库表结构', '数据库查询', 'GET'
WHERE NOT EXISTS (
  SELECT 1 FROM sys_apis
  WHERE path = '/system/database-query/table-structure' AND method = 'GET' AND deleted_at IS NULL
);

INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
SELECT 'p', '888', '/system/database-query/tables', 'GET', '', '', ''
WHERE NOT EXISTS (
  SELECT 1 FROM casbin_rule
  WHERE ptype = 'p' AND v0 = '888' AND v1 = '/system/database-query/tables' AND v2 = 'GET'
);

INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
SELECT 'p', '888', '/system/database-query/table-structure', 'GET', '', '', ''
WHERE NOT EXISTS (
  SELECT 1 FROM casbin_rule
  WHERE ptype = 'p' AND v0 = '888' AND v1 = '/system/database-query/table-structure' AND v2 = 'GET'
);
