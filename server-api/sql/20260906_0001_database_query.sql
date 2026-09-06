-- 数据库只读查询页面、接口及超级管理员权限。

INSERT INTO sys_base_menus (
  created_at, updated_at, deleted_at, menu_level, parent_id, path, name, hidden,
  component, sort, active_name, keep_alive, default_menu, title, icon, close_tab, transition_type
)
SELECT
  NOW(3), NOW(3), NULL, 2, parent.id, 'database-query', 'databaseQuery', 0,
  'view/superAdmin/databaseQuery/index.vue', 6, '', 0, 0, '数据库查询', 'search', 0, ''
FROM sys_base_menus AS parent
WHERE parent.name = 'observability'
  AND parent.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_base_menus AS menu
    WHERE menu.name = 'databaseQuery' AND menu.deleted_at IS NULL
  )
LIMIT 1;

INSERT INTO sys_authority_menus (sys_base_menu_id, sys_authority_authority_id)
SELECT menu.id, 888
FROM sys_base_menus AS menu
WHERE menu.name = 'databaseQuery'
  AND menu.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_authority_menus AS relation
    WHERE relation.sys_base_menu_id = menu.id
      AND relation.sys_authority_authority_id = 888
  );

INSERT INTO sys_apis (created_at, updated_at, deleted_at, path, description, api_group, method)
SELECT NOW(3), NOW(3), NULL, '/system/database-query/execute', '执行只读数据库查询', '数据库查询', 'POST'
WHERE NOT EXISTS (
  SELECT 1 FROM sys_apis
  WHERE path = '/system/database-query/execute'
    AND method = 'POST'
    AND deleted_at IS NULL
);

INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
SELECT 'p', '888', '/system/database-query/execute', 'POST', '', '', ''
WHERE NOT EXISTS (
  SELECT 1 FROM casbin_rule
  WHERE ptype = 'p'
    AND v0 = '888'
    AND v1 = '/system/database-query/execute'
    AND v2 = 'POST'
);
