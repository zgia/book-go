SET NAMES utf8mb4;

DROP TABLE IF EXISTS `author`;
CREATE TABLE
  `author` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name` varchar(128) NOT NULL DEFAULT '' COMMENT '作者名',
    `former_name` varchar(500) NOT NULL DEFAULT '' COMMENT '曾用名',
    `createdat` int unsigned NOT NULL DEFAULT '0',
    `updatedat` int unsigned NOT NULL DEFAULT '0',
    `deletedat` int unsigned NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`) USING BTREE
  ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '作者';

DROP TABLE IF EXISTS `book`;
CREATE TABLE
  `book` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `categoryid` smallint unsigned NOT NULL DEFAULT '0' COMMENT '分类',
    `title` varchar(255) NOT NULL DEFAULT '' COMMENT '名称',
    `alias` varchar(255) NOT NULL DEFAULT '' COMMENT '名称',
    `authorid` int unsigned NOT NULL DEFAULT '0' COMMENT '作者ID',
    `summary` text COMMENT '简介',
    `source` varchar(255) NOT NULL DEFAULT '' COMMENT '来源',
    `latest` varchar(255) NOT NULL DEFAULT '' COMMENT '最新章节',
    `rate` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '评分',
    `wordcount` int unsigned NOT NULL DEFAULT '0' COMMENT '字数',
    `isfinished` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '小说是否完本',
    `cover` varchar(255) NOT NULL DEFAULT '' COMMENT '封面',
    `createdat` int unsigned NOT NULL DEFAULT '0',
    `updatedat` int unsigned NOT NULL DEFAULT '0',
    `deletedat` int unsigned NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`)
  ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '图书';

DROP TABLE IF EXISTS `category`;
CREATE TABLE
  `category` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `parentid` int unsigned NOT NULL DEFAULT '0',
    `title` varchar(128) NOT NULL DEFAULT '',
    `ishidden` int unsigned NOT NULL DEFAULT '0' COMMENT '是否隐藏当前分类',
    PRIMARY KEY (`id`)
  ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '分类';

DROP TABLE IF EXISTS `chapter`;
CREATE TABLE
  `chapter` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `bookid` int unsigned NOT NULL DEFAULT '0',
    `volumeid` int unsigned NOT NULL DEFAULT '0',
    `title` varchar(255) NOT NULL DEFAULT '' COMMENT '章节标题',
    `wordcount` int unsigned NOT NULL DEFAULT '0' COMMENT '字数',
    `createdat` int unsigned NOT NULL DEFAULT '0',
    `updatedat` int unsigned NOT NULL DEFAULT '0',
    `deletedat` int unsigned NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    KEY `idx_book` (`bookid`)
  ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '章节';

DROP TABLE IF EXISTS `content`;
CREATE TABLE
  `content` (
    `chapterid` int unsigned NOT NULL AUTO_INCREMENT,
    `txt` mediumtext COMMENT '章节内容',
    PRIMARY KEY (`chapterid`)
  ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '内容';

DROP TABLE IF EXISTS `user`;
CREATE TABLE
  `user` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `username` varchar(32) NOT NULL DEFAULT '' COMMENT '用户名称',
    `realname` varchar(128) NOT NULL DEFAULT '' COMMENT '用户真实名称',
    `mobile` varchar(11) NOT NULL DEFAULT '' COMMENT '手机',
    `password` varchar(32) NOT NULL DEFAULT '' COMMENT '密码',
    `salt` varchar(6) NOT NULL DEFAULT '' COMMENT '密码salt',
    `lastip` varchar(16) NOT NULL DEFAULT '' COMMENT '最后登录IP地址',
    `lastrealip` varchar(16) NOT NULL DEFAULT '' COMMENT '最后登录真实IP地址',
    `createdat` int unsigned NOT NULL DEFAULT '0',
    `updatedat` int unsigned NOT NULL DEFAULT '0',
    `deletedat` int unsigned NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`) USING BTREE
  ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '用户';

DROP TABLE IF EXISTS `volume`;
CREATE TABLE
  `volume` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `bookid` int unsigned NOT NULL DEFAULT '0',
    `title` varchar(255) NOT NULL DEFAULT '' COMMENT '卷名',
    `summary` text COMMENT '简介',
    `cover` varchar(255) DEFAULT '' COMMENT '封面',
    `createdat` int unsigned NOT NULL DEFAULT '0',
    `updatedat` int unsigned NOT NULL DEFAULT '0',
    `deletedat` int unsigned NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    KEY `idx_book` (`bookid`)
  ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '卷';