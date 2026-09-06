-- Pprof 火焰图独立详情页路由（隐藏菜单）。

INSERT INTO sys_base_menus (
  created_at, updated_at, deleted_at, menu_level, parent_id, path, name, hidden,
  component, sort, active_name, keep_alive, default_menu, title, icon, close_tab, transition_type
)
SELECT
  NOW(3), NOW(3), NULL, 2, parent.id, 'pprof-record/:id', 'pprofRecord', 1,
  'view/superAdmin/pprof/detail.vue', 8, 'pprofMonitor', 0, 0, 'Pprof 火焰图-${id}', 'cpu', 0, ''
FROM sys_base_menus AS parent
WHERE parent.name = 'observability'
  AND parent.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_base_menus AS menu
    WHERE menu.name = 'pprofRecord' AND menu.deleted_at IS NULL
  )
LIMIT 1;

INSERT INTO sys_authority_menus (sys_base_menu_id, sys_authority_authority_id)
SELECT menu.id, 888
FROM sys_base_menus AS menu
WHERE menu.name = 'pprofRecord'
  AND menu.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_authority_menus AS relation
    WHERE relation.sys_base_menu_id = menu.id
      AND relation.sys_authority_authority_id = 888
  );
