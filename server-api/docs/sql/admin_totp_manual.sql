ALTER TABLE `sys_users`
  ADD COLUMN `totp_secret` TEXT NULL COMMENT 'Google TOTP密钥' AFTER `origin_setting`,
  ADD COLUMN `totp_enabled` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否启用Google TOTP' AFTER `totp_secret`,
  ADD COLUMN `totp_bound_at` DATETIME NULL COMMENT 'Google TOTP绑定时间' AFTER `totp_enabled`;
