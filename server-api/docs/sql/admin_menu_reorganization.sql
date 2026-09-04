-- 管理后台菜单重组（可重复执行）
--
-- 执行前请备份数据库。脚本仅调整 sys_base_menus 与 sys_authority_menus：
-- 不删除菜单、不变更业务路由，也不回收任何既有角色权限。
-- 适用于已通过 go_app_starter.sql 初始化的数据库。

START TRANSACTION;

-- 系统管理下的功能分组。RouterHolder 仅承担嵌套路由，不显示独立页面。
INSERT INTO sys_base_menus
    (created_at, updated_at, menu_level, parent_id, path, name, hidden, component, sort,
     active_name, keep_alive, default_menu, title, icon, close_tab, transition_type)
SELECT NOW(3), NOW(3), 1, system_menu.id, 'access-control', 'accessControl', 0, 'view/routerHolder.vue', 1,
       '', 0, 0, '权限与账号', 'lock', 0, ''
FROM sys_base_menus AS system_menu
WHERE system_menu.name = 'superAdmin'
  AND NOT EXISTS (SELECT 1 FROM sys_base_menus WHERE name = 'accessControl' AND deleted_at IS NULL);

INSERT INTO sys_base_menus
    (created_at, updated_at, menu_level, parent_id, path, name, hidden, component, sort,
     active_name, keep_alive, default_menu, title, icon, close_tab, transition_type)
SELECT NOW(3), NOW(3), 1, system_menu.id, 'system-settings', 'systemSettings', 0, 'view/routerHolder.vue', 2,
       '', 0, 0, '系统配置', 'setting', 0, ''
FROM sys_base_menus AS system_menu
WHERE system_menu.name = 'superAdmin'
  AND NOT EXISTS (SELECT 1 FROM sys_base_menus WHERE name = 'systemSettings' AND deleted_at IS NULL);

INSERT INTO sys_base_menus
    (created_at, updated_at, menu_level, parent_id, path, name, hidden, component, sort,
     active_name, keep_alive, default_menu, title, icon, close_tab, transition_type)
SELECT NOW(3), NOW(3), 1, system_menu.id, 'observability', 'observability', 0, 'view/routerHolder.vue', 3,
       '', 0, 0, '运行监控', 'monitor', 0, ''
FROM sys_base_menus AS system_menu
WHERE system_menu.name = 'superAdmin'
  AND NOT EXISTS (SELECT 1 FROM sys_base_menus WHERE name = 'observability' AND deleted_at IS NULL);

-- 顶层仅保留：仪表盘、权限与账号、系统配置、运行监控。
UPDATE sys_base_menus
SET sort = 1, updated_at = NOW(3)
WHERE name = 'dashboard';

UPDATE sys_base_menus
SET title = '系统管理', icon = 'setting', hidden = 1, updated_at = NOW(3)
WHERE name = 'superAdmin';

UPDATE sys_base_menus
SET title = '开发工具', hidden = 1, updated_at = NOW(3)
WHERE name = 'systemTools';

UPDATE sys_base_menus
SET title = '扩展中心', hidden = 1, updated_at = NOW(3)
WHERE name = 'plugin';

UPDATE sys_base_menus
SET title = '开发示例', hidden = 1, updated_at = NOW(3)
WHERE name = 'example';

UPDATE sys_base_menus
SET hidden = 1, updated_at = NOW(3)
WHERE name IN ('about', 'person', 'https://www.gin-vue-admin.com');

UPDATE sys_base_menus
SET parent_id = 0, menu_level = 0, hidden = 0, updated_at = NOW(3)
WHERE name IN ('accessControl', 'systemSettings', 'observability');

UPDATE sys_base_menus SET sort = 2 WHERE name = 'accessControl';
UPDATE sys_base_menus SET sort = 3 WHERE name = 'systemSettings';
UPDATE sys_base_menus SET sort = 4 WHERE name = 'observability';

-- 权限与账号：用户、角色、菜单、接口与 API Token。
UPDATE sys_base_menus AS child
JOIN sys_base_menus AS parent ON parent.name = 'accessControl'
SET child.parent_id = parent.id, child.menu_level = 2, child.updated_at = NOW(3)
WHERE child.name IN ('user', 'authority', 'menu', 'api', 'apiToken');

UPDATE sys_base_menus SET sort = 1 WHERE name = 'user';
UPDATE sys_base_menus SET sort = 2 WHERE name = 'authority';
UPDATE sys_base_menus SET sort = 3 WHERE name = 'menu';
UPDATE sys_base_menus SET sort = 4 WHERE name = 'api';
UPDATE sys_base_menus SET sort = 5 WHERE name = 'apiToken';

-- 系统配置：字典、参数、全局配置与版本。
UPDATE sys_base_menus AS child
JOIN sys_base_menus AS parent ON parent.name = 'systemSettings'
SET child.parent_id = parent.id, child.menu_level = 2, child.updated_at = NOW(3)
WHERE child.name IN ('dictionary', 'sysParams', 'system', 'sysVersion');

UPDATE sys_base_menus SET sort = 1 WHERE name = 'dictionary';
UPDATE sys_base_menus SET sort = 2 WHERE name = 'sysParams';
UPDATE sys_base_menus SET sort = 3 WHERE name = 'system';
UPDATE sys_base_menus SET sort = 4 WHERE name = 'sysVersion';

-- 运行监控：服务状态、限流策略与操作/登录/错误日志。
UPDATE sys_base_menus AS child
JOIN sys_base_menus AS parent ON parent.name = 'observability'
SET child.parent_id = parent.id, child.menu_level = 2, child.updated_at = NOW(3)
WHERE child.name IN ('state', 'apiRateLimit', 'operation', 'loginLog', 'sysError');

UPDATE sys_base_menus SET sort = 1 WHERE name = 'state';
UPDATE sys_base_menus SET sort = 2 WHERE name = 'apiRateLimit';
UPDATE sys_base_menus SET sort = 3 WHERE name = 'operation';
UPDATE sys_base_menus SET sort = 4 WHERE name = 'loginLog';
UPDATE sys_base_menus SET sort = 5 WHERE name = 'sysError';

-- 业务用户列表属于示例模块，避免以一级菜单分散在业务区。
UPDATE sys_base_menus AS child
JOIN sys_base_menus AS parent ON parent.name = 'example'
SET child.parent_id = parent.id, child.menu_level = 1, child.sort = 1, child.updated_at = NOW(3)
WHERE child.name = 'userInfo';

-- 让已拥有分组内任一菜单的角色自动获得新分组的可见权限。
INSERT IGNORE INTO sys_authority_menus (sys_base_menu_id, sys_authority_authority_id)
SELECT parent.id, child_auth.sys_authority_authority_id
FROM sys_base_menus AS parent
JOIN sys_base_menus AS child ON child.parent_id = parent.id
JOIN sys_authority_menus AS child_auth ON child_auth.sys_base_menu_id = child.id
WHERE parent.name IN ('accessControl', 'systemSettings', 'observability')
  AND parent.deleted_at IS NULL
  AND child.deleted_at IS NULL;

COMMIT;

-- 执行后请重新登录后台，前端会重新拉取菜单与路由。
