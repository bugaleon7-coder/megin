-- Pprof 动态开关、CPU 采样和火焰图页面。

INSERT INTO sys_base_menus (
  created_at, updated_at, deleted_at, menu_level, parent_id, path, name, hidden,
  component, sort, active_name, keep_alive, default_menu, title, icon, close_tab, transition_type
)
SELECT
  NOW(3), NOW(3), NULL, 2, parent.id, 'pprof', 'pprofMonitor', 0,
  'view/superAdmin/pprof/index.vue', 7, '', 0, 0, 'Pprof 性能分析', 'cpu', 0, ''
FROM sys_base_menus AS parent
WHERE parent.name = 'observability'
  AND parent.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_base_menus AS menu
    WHERE menu.name = 'pprofMonitor' AND menu.deleted_at IS NULL
  )
LIMIT 1;

INSERT INTO sys_authority_menus (sys_base_menu_id, sys_authority_authority_id)
SELECT menu.id, 888
FROM sys_base_menus AS menu
WHERE menu.name = 'pprofMonitor'
  AND menu.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_authority_menus AS relation
    WHERE relation.sys_base_menu_id = menu.id
      AND relation.sys_authority_authority_id = 888
  );

INSERT INTO sys_apis (created_at, updated_at, deleted_at, path, description, api_group, method)
SELECT NOW(3), NOW(3), NULL, source.path, source.description, 'Pprof 性能分析', source.method
FROM (
  SELECT '/system/pprof/status' AS path, '获取 pprof 状态' AS description, 'GET' AS method
  UNION ALL SELECT '/system/pprof/toggle', '启停 pprof', 'POST'
  UNION ALL SELECT '/system/pprof/cpu/start', '启动 CPU Profile', 'POST'
  UNION ALL SELECT '/system/pprof/cpu/result', '获取 CPU Profile 结果', 'GET'
) AS source
WHERE NOT EXISTS (
  SELECT 1 FROM sys_apis
  WHERE sys_apis.path = source.path
    AND sys_apis.method = source.method
    AND sys_apis.deleted_at IS NULL
);

INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
SELECT 'p', '888', source.path, source.method, '', '', ''
FROM (
  SELECT '/system/pprof/status' AS path, 'GET' AS method
  UNION ALL SELECT '/system/pprof/toggle', 'POST'
  UNION ALL SELECT '/system/pprof/cpu/start', 'POST'
  UNION ALL SELECT '/system/pprof/cpu/result', 'GET'
) AS source
WHERE NOT EXISTS (
  SELECT 1 FROM casbin_rule
  WHERE casbin_rule.ptype = 'p'
    AND casbin_rule.v0 = '888'
    AND casbin_rule.v1 = source.path
    AND casbin_rule.v2 = source.method
);
