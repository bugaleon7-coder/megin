-- 前台 API 限流规则表。
-- MySQL 只保存规则；请求计数和令牌桶状态保存在各服务实例内存中。
CREATE TABLE IF NOT EXISTS `api_rate_limit_rules` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '限流规则ID',
  `name` varchar(100) NOT NULL COMMENT '规则名称',
  `scope_type` tinyint NOT NULL COMMENT '作用范围：1全局，2指定接口',
  `http_method` varchar(10) NOT NULL DEFAULT '*' COMMENT 'HTTP方法，全局规则固定为*',
  `route_path` varchar(255) NOT NULL DEFAULT '*' COMMENT 'Gin路由模板，全局规则固定为*',
  `dimension` tinyint NOT NULL COMMENT '限流维度：1 IP，2 UID',
  `rate_count` int NOT NULL COMMENT '每个周期补充的令牌数',
  `interval_seconds` int NOT NULL COMMENT '令牌补充周期秒数',
  `burst` int NOT NULL COMMENT '令牌桶容量和瞬时放行上限',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态：0禁用，1启用',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `created_by` bigint unsigned NOT NULL DEFAULT '0' COMMENT '创建管理员ID',
  `updated_by` bigint unsigned NOT NULL DEFAULT '0' COMMENT '最后修改管理员ID',
  `created_at` datetime(3) NOT NULL COMMENT '创建时间',
  `updated_at` datetime(3) NOT NULL COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_scope_route_dimension` (`scope_type`,`http_method`,`route_path`,`dimension`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='前台API限流规则';

-- 首次部署可以按实际容量调整下面两条全局规则。
INSERT IGNORE INTO `api_rate_limit_rules`
(`name`,`scope_type`,`http_method`,`route_path`,`dimension`,`rate_count`,`interval_seconds`,`burst`,`status`,`remark`,`created_by`,`updated_by`,`created_at`,`updated_at`)
VALUES
('全局IP限流',1,'*','*',1,2,1,2,1,'所有/api接口默认按IP限流',0,0,NOW(3),NOW(3)),
('全局UID限流',1,'*','*',2,1,5,1,1,'鉴权接口默认按UID限流',0,0,NOW(3),NOW(3)),
('健康检查接口IP限流示例',2,'GET','/api/health',1,1,5,1,1,'默认接口级限流示例：每个IP每5秒允许1个请求',0,0,NOW(3),NOW(3));

-- 注册后台接口元数据。按照项目的 Casbin 兼容规则，路径不保存 /admin-api 前缀。
UPDATE `sys_apis`
SET `path` = CONCAT('/system', `path`), `updated_at` = NOW(3)
WHERE `path` LIKE '/rate-limit/%'
  AND `deleted_at` IS NULL;

INSERT INTO `sys_apis` (`created_at`,`updated_at`,`path`,`description`,`api_group`,`method`)
SELECT NOW(3),NOW(3),source.path,source.description,'API限流规则',source.method
FROM (
  SELECT '/system/rate-limit/create' AS path,'新增API限流规则' AS description,'POST' AS method
  UNION ALL SELECT '/system/rate-limit/update','修改API限流规则','PUT'
  UNION ALL SELECT '/system/rate-limit/changeStatus','启停API限流规则','PUT'
  UNION ALL SELECT '/system/rate-limit/delete','删除API限流规则','DELETE'
  UNION ALL SELECT '/system/rate-limit/detail','查询API限流规则详情','GET'
  UNION ALL SELECT '/system/rate-limit/pageList','分页查询API限流规则','GET'
  UNION ALL SELECT '/system/rate-limit/refresh','手动刷新API限流规则','POST'
) AS source
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_apis`
  WHERE `sys_apis`.`path` = source.path
    AND `sys_apis`.`method` = source.method
    AND `sys_apis`.`deleted_at` IS NULL
);

-- 为默认超级管理员角色 888 授权；其他角色仍需通过后台权限管理按需分配。
UPDATE `casbin_rule`
SET `v1` = CONCAT('/system', `v1`)
WHERE `ptype` = 'p'
  AND `v1` LIKE '/rate-limit/%';

INSERT INTO `casbin_rule` (`ptype`,`v0`,`v1`,`v2`,`v3`,`v4`,`v5`)
SELECT 'p','888',source.path,source.method,'','',''
FROM (
  SELECT '/system/rate-limit/create' AS path,'POST' AS method
  UNION ALL SELECT '/system/rate-limit/update','PUT'
  UNION ALL SELECT '/system/rate-limit/changeStatus','PUT'
  UNION ALL SELECT '/system/rate-limit/delete','DELETE'
  UNION ALL SELECT '/system/rate-limit/detail','GET'
  UNION ALL SELECT '/system/rate-limit/pageList','GET'
  UNION ALL SELECT '/system/rate-limit/refresh','POST'
) AS source
WHERE NOT EXISTS (
  SELECT 1 FROM `casbin_rule`
  WHERE `casbin_rule`.`ptype` = 'p'
    AND `casbin_rule`.`v0` = '888'
    AND `casbin_rule`.`v1` = source.path
    AND `casbin_rule`.`v2` = source.method
);

-- 注册“API限流管理”后台菜单，组件路径对应 web/src/view/superAdmin/rateLimit/index.vue。
-- 菜单放在“超级管理员”目录下，并默认授予超级管理员角色 888。
INSERT INTO `sys_base_menus`
(`created_at`,`updated_at`,`menu_level`,`parent_id`,`path`,`name`,`hidden`,`component`,`sort`,`active_name`,`keep_alive`,`default_menu`,`title`,`icon`,`close_tab`,`transition_type`)
SELECT NOW(3),NOW(3),1,3,'apiRateLimit','apiRateLimit',0,'view/superAdmin/rateLimit/index.vue',4,'',0,0,'API限流管理','timer',0,''
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_base_menus`
  WHERE `name` = 'apiRateLimit'
    AND `deleted_at` IS NULL
);

INSERT INTO `sys_authority_menus` (`sys_base_menu_id`,`sys_authority_authority_id`)
SELECT menu.id,888
FROM `sys_base_menus` AS menu
WHERE menu.name = 'apiRateLimit'
  AND menu.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM `sys_authority_menus` AS authority_menu
    WHERE authority_menu.sys_base_menu_id = menu.id
      AND authority_menu.sys_authority_authority_id = 888
  );
