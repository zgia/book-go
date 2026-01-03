PRAGMA foreign_keys = off;

BEGIN TRANSACTION;

-- 表：作者，author
DROP TABLE IF EXISTS author;

CREATE TABLE
    IF NOT EXISTS author (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL DEFAULT (''), -- 作者名称
        former_name TEXT NOT NULL DEFAULT (''), -- 曾用名
        createdat INTEGER NOT NULL DEFAULT (0),
        updatedat INTEGER NOT NULL DEFAULT (0),
        deletedat INTEGER NOT NULL DEFAULT (0)
    );

-- 表：图书，book
DROP TABLE IF EXISTS book;

CREATE TABLE
    IF NOT EXISTS book (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        categoryid INTEGER NOT NULL DEFAULT (0), -- 分类ID
        title TEXT NOT NULL DEFAULT (''), -- 标题
        alias TEXT DEFAULT ('') NOT NULL, -- 别名
        authorid INTEGER NOT NULL DEFAULT (0), -- 作者
        summary TEXT NOT NULL DEFAULT (''), -- 摘要
        source TEXT NOT NULL DEFAULT (''), -- 来源
        latest TEXT NOT NULL DEFAULT (''), -- 最后更新
        rate INTEGER NOT NULL DEFAULT (0), -- 打分：1、2、3、4、5
        wordcount INTEGER NOT NULL DEFAULT (0), -- 字数
        isfinished INTEGER NOT NULL DEFAULT (0), -- 是否完本
        cover TEXT NOT NULL DEFAULT (''), -- 封面：路径
        createdat INTEGER NOT NULL DEFAULT (0),
        updatedat INTEGER NOT NULL DEFAULT (0),
        deletedat INTEGER NOT NULL DEFAULT (0)
    );

-- 表：分类，category
DROP TABLE IF EXISTS category;

CREATE TABLE
    IF NOT EXISTS category (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        parentid INTEGER NOT NULL DEFAULT (0), -- 父ID
        title TEXT NOT NULL DEFAULT (''), -- 分类名称
        ishidden INTEGER NOT NULL DEFAULT (0) -- 是否隐藏的分类
    );

-- 表：章节，chapter
DROP TABLE IF EXISTS chapter;

CREATE TABLE
    IF NOT EXISTS chapter (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        bookid INTEGER NOT NULL DEFAULT (0), -- 图书ID
        volumeid INTEGER NOT NULL DEFAULT (0), -- 卷ID
        title TEXT DEFAULT ('') NOT NULL, -- 章节标题
        wordcount INTEGER NOT NULL DEFAULT (0), -- 章节字数
        createdat INTEGER NOT NULL DEFAULT (0),
        updatedat INTEGER NOT NULL DEFAULT (0),
        deletedat INTEGER NOT NULL DEFAULT (0)
    );

-- 表：章节内容，content
DROP TABLE IF EXISTS content;

CREATE TABLE
    IF NOT EXISTS content (
        chapterid INTEGER PRIMARY KEY NOT NULL DEFAULT (0), -- 章节ID
        txt TEXT -- 章节内容
    );

-- 表：用户，user
DROP TABLE IF EXISTS user;

CREATE TABLE
    IF NOT EXISTS user (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL DEFAULT (''), -- 用户名
        realname TEXT NOT NULL DEFAULT (''), -- 真实名称
        mobile TEXT NOT NULL DEFAULT (''), -- 手机
        password TEXT NOT NULL DEFAULT (''), -- 密码
        salt TEXT NOT NULL DEFAULT (''), -- 加密盐
        lastip TEXT NOT NULL DEFAULT (''), -- 最后IP
        lastrealip TEXT NOT NULL DEFAULT (''), -- 最后真实IP
        createdat INTEGER NOT NULL DEFAULT (0),
        updatedat INTEGER NOT NULL DEFAULT (0),
        deletedat INTEGER NOT NULL DEFAULT (0)
    );

-- 表：卷，volume
DROP TABLE IF EXISTS volume;

CREATE TABLE
    IF NOT EXISTS volume (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        bookid INTEGER NOT NULL DEFAULT (0), -- 图书ID
        title TEXT NOT NULL DEFAULT (''), -- 卷名称
        summary TEXT NOT NULL DEFAULT (''), -- 卷摘要
        cover TEXT NOT NULL DEFAULT (''), -- 卷封面
        createdat INTEGER NOT NULL DEFAULT (0),
        updatedat INTEGER NOT NULL DEFAULT (0),
        deletedat INTEGER DEFAULT (0) NOT NULL
    );

COMMIT TRANSACTION;

PRAGMA foreign_keys = on;