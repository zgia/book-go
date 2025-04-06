SET NAMES utf8mb4;

DROP TABLE IF EXISTS `book`;
CREATE TABLE `book` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `categoryid` smallint(5) unsigned NOT NULL DEFAULT 0 COMMENT '分类',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '名称',
  `alias` varchar(255) NOT NULL DEFAULT '' COMMENT '名称别名',
  `author` varchar(255) NOT NULL DEFAULT '' COMMENT '作者',
  `summary` text DEFAULT NULL COMMENT '简介',
  `source` varchar(255) NOT NULL DEFAULT '' COMMENT '来源',
  `latest` varchar(255) NOT NULL DEFAULT '' COMMENT '最新章节',
  `isfinished` tinyint(1) unsigned NOT NULL DEFAULT 0 COMMENT '是否全本',
  `cover` varchar(255) NOT NULL DEFAULT '' COMMENT '封面',
  `wordcount` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '字数',
  `createdat` int(10) unsigned NOT NULL DEFAULT 0,
  `updatedat` int(10) unsigned NOT NULL DEFAULT 0,
  `deletedat` int(10) unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='图书';


DROP TABLE IF EXISTS `category`;
CREATE TABLE `category` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `parentid` int(10) unsigned NOT NULL DEFAULT 0,
  `title` varchar(128) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='分类';


DROP TABLE IF EXISTS `chapter`;
CREATE TABLE `chapter` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `bookid` int(10) unsigned NOT NULL DEFAULT 0,
  `volumeid` int(10) unsigned NOT NULL DEFAULT 0,
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '章节标题',
  `wordcount` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '字数',
  `createdat` int(10) unsigned NOT NULL DEFAULT 0,
  `updatedat` int(10) unsigned NOT NULL DEFAULT 0,
  `deletedat` int(10) unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_book` (`bookid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='章节';


DROP TABLE IF EXISTS `content`;
CREATE TABLE `content` (
  `chapterid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `txt` mediumtext DEFAULT NULL COMMENT '章节内容',
  PRIMARY KEY (`chapterid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='内容';


DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `username` varchar(32) NOT NULL DEFAULT '' COMMENT '用户名称',
  `realname` varchar(128) NOT NULL DEFAULT '' COMMENT '用户真实名称',
  `mobile` varchar(11) NOT NULL DEFAULT '' COMMENT '手机',
  `password` varchar(32) NOT NULL DEFAULT '' COMMENT '密码',
  `salt` varchar(6) NOT NULL DEFAULT '' COMMENT '密码salt',
  `lastip` varchar(16) NOT NULL DEFAULT '' COMMENT '最后登录IP地址',
  `lastrealip` varchar(16) NOT NULL DEFAULT '' COMMENT '最后登录真实IP地址',
  `createdat` int(10) unsigned NOT NULL DEFAULT 0,
  `updatedat` int(10) unsigned NOT NULL DEFAULT 0,
  `deletedat` int(10) unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户';


DROP TABLE IF EXISTS `volume`;
CREATE TABLE `volume` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `bookid` int(10) unsigned NOT NULL DEFAULT 0,
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '卷名',
  `summary` text DEFAULT NULL COMMENT '简介',
  `cover` varchar(255) DEFAULT '' COMMENT '封面',
  `createdat` int(10) unsigned NOT NULL DEFAULT 0,
  `updatedat` int(10) unsigned NOT NULL DEFAULT 0,
  `deletedat` int(10) unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_book` (`bookid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='卷';
