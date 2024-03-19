# ************************************************************
# Sequel Ace SQL dump
# 版本号： 20050
#
# https://sequel-ace.com/
# https://github.com/Sequel-Ace/Sequel-Ace
#
# 主机: localhost (MySQL 8.0.32)
# 数据库: library
# 生成时间: 2023-08-10 02:30:12 +0000
# ************************************************************


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
SET NAMES utf8mb4;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE='NO_AUTO_VALUE_ON_ZERO', SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


# 转储表 book
# ------------------------------------------------------------

DROP TABLE IF EXISTS `book`;

CREATE TABLE `book` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `categoryid` smallint unsigned NOT NULL DEFAULT '0' COMMENT '分类',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '名称',
  `author` varchar(255) NOT NULL DEFAULT '' COMMENT '作者',
  `summary` text COMMENT '简介',
  `source` varchar(255) NOT NULL DEFAULT '' COMMENT '来源',
  `wordcount` int unsigned NOT NULL DEFAULT '0' COMMENT '字数',
  `createdat` int unsigned NOT NULL DEFAULT '0',
  `updatedat` int unsigned NOT NULL DEFAULT '0',
  `deletedat` int unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='图书';



# 转储表 category
# ------------------------------------------------------------

DROP TABLE IF EXISTS `category`;

CREATE TABLE `category` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `parentid` int unsigned NOT NULL DEFAULT '0',
  `title` varchar(128) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='分类';



# 转储表 chapter
# ------------------------------------------------------------

DROP TABLE IF EXISTS `chapter`;

CREATE TABLE `chapter` (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='章节';



# 转储表 content
# ------------------------------------------------------------

DROP TABLE IF EXISTS `content`;

CREATE TABLE `content` (
  `chapterid` int unsigned NOT NULL AUTO_INCREMENT,
  `txt` mediumtext COMMENT '章节内容',
  PRIMARY KEY (`chapterid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='内容';



# 转储表 user
# ------------------------------------------------------------

DROP TABLE IF EXISTS `user`;

CREATE TABLE `user` (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户';



# 转储表 volume
# ------------------------------------------------------------

DROP TABLE IF EXISTS `volume`;

CREATE TABLE `volume` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `bookid` int unsigned NOT NULL DEFAULT '0',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '卷名',
  `summary` text COMMENT '简介',
  `createdat` int unsigned NOT NULL DEFAULT '0',
  `updatedat` int unsigned NOT NULL DEFAULT '0',
  `deletedat` int unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_book` (`bookid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='卷';




/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
