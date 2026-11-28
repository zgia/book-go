-- ============================================================
-- zGia! Book Library · 数据库初始化脚本（MySQL 5.7+ / 8.0+）
--
-- 用法：
--   mysql -u <user> -p <database> < book.sql
--
-- 约定：
--   * 表引擎 InnoDB，字符集 utf8mb4；
--   * 时间字段统一 DATETIME：created_at / updated_at 默认当前时间，
--     updated_at 在行更新时自动刷新；deleted_at 为 NULL 表示未删除（软删除）；
--   * 列名 snake_case：外键 <表>_id，布尔 is_*，计数 *_count；
--   * 正文搜索索引为 SQLite 专属（FTS5），MySQL 下不使用。
-- ============================================================

SET NAMES utf8mb4;
-- 与 [time] ZONE 保持一致，按服务器实际时区调整。
SET time_zone = '+08:00';

-- ---------- 作者 ----------
CREATE TABLE IF NOT EXISTS `author` (
  `id`          BIGINT       NOT NULL AUTO_INCREMENT,
  `name`        VARCHAR(128) NOT NULL DEFAULT '',
  `former_name` VARCHAR(500) NOT NULL DEFAULT '',
  `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at`  DATETIME     NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ---------- 图书 ----------
CREATE TABLE IF NOT EXISTS `book` (
  `id`          BIGINT       NOT NULL AUTO_INCREMENT,
  `category_id` BIGINT       NOT NULL DEFAULT 0,
  `title`       VARCHAR(255) NOT NULL DEFAULT '',
  `alias`       VARCHAR(255) NOT NULL DEFAULT '',
  `author_id`   BIGINT       NOT NULL DEFAULT 0,
  `summary`     TEXT         NULL,
  `source`      VARCHAR(255) NOT NULL DEFAULT '',
  `latest`      VARCHAR(255) NOT NULL DEFAULT '',
  `rate`        INT          NOT NULL DEFAULT 0,
  `word_count`  BIGINT       NOT NULL DEFAULT 0,
  `is_finished` TINYINT      NOT NULL DEFAULT 0,
  `cover`       VARCHAR(255) NOT NULL DEFAULT '',
  `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at`  DATETIME     NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_book_category` (`category_id`),
  KEY `idx_book_author` (`author_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ---------- 分类 ----------
CREATE TABLE IF NOT EXISTS `category` (
  `id`        BIGINT       NOT NULL AUTO_INCREMENT,
  `parent_id` BIGINT       NOT NULL DEFAULT 0,
  `title`     VARCHAR(128) NOT NULL DEFAULT '',
  `is_hidden` TINYINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ---------- 卷 ----------
CREATE TABLE IF NOT EXISTS `volume` (
  `id`         BIGINT       NOT NULL AUTO_INCREMENT,
  `book_id`    BIGINT       NOT NULL DEFAULT 0,
  `title`      VARCHAR(255) NOT NULL DEFAULT '',
  `summary`    TEXT         NULL,
  `cover`      VARCHAR(255) NULL DEFAULT '',
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME     NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_volume_book` (`book_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ---------- 章节 ----------
CREATE TABLE IF NOT EXISTS `chapter` (
  `id`         BIGINT       NOT NULL AUTO_INCREMENT,
  `book_id`    BIGINT       NOT NULL DEFAULT 0,
  `volume_id`  BIGINT       NOT NULL DEFAULT 0,
  `title`      VARCHAR(255) NOT NULL DEFAULT '',
  `word_count` BIGINT       NOT NULL DEFAULT 0,
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME     NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_chapter_book` (`book_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ---------- 正文 ----------
-- 与 chapter 一对一：chapter_id 对应 chapter.id，不设自增主键。
-- 正文单章可能超过 64KB，使用 MEDIUMTEXT。
CREATE TABLE IF NOT EXISTS `content` (
  `chapter_id` BIGINT   NOT NULL DEFAULT 0,
  `txt`        MEDIUMTEXT NULL,
  KEY `idx_content_chapter` (`chapter_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ---------- 用户 ----------
CREATE TABLE IF NOT EXISTS `user` (
  `id`           BIGINT       NOT NULL AUTO_INCREMENT,
  `username`     VARCHAR(32)  NOT NULL DEFAULT '',
  `realname`     VARCHAR(128) NOT NULL DEFAULT '',
  `mobile`       VARCHAR(64)  NOT NULL DEFAULT '',
  `password`     VARCHAR(32)  NOT NULL DEFAULT '',
  `salt`         VARCHAR(6)   NOT NULL DEFAULT '',
  `last_ip`      VARCHAR(16)  NOT NULL DEFAULT '',
  `last_real_ip` VARCHAR(16)  NOT NULL DEFAULT '',
  `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at`   DATETIME     NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_username` (`username`),
  UNIQUE KEY `uk_user_mobile` (`mobile`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
