/*
SQLyog Ultimate v12.09 (64 bit)
MySQL - 5.5.42-log : Database - loan_marketing
*********************************************************************
*/


/*!40101 SET NAMES utf8 */;

/*!40101 SET SQL_MODE=''*/;

/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
CREATE DATABASE /*!32312 IF NOT EXISTS*/`loan_marketing` /*!40100 DEFAULT CHARACTER SET utf8 */;

USE `gva`;

/*Table structure for table `task_info` */

DROP TABLE IF EXISTS `task_info`;

CREATE TABLE `task_info` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '索引ID',
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  `user_id` varchar(128) NOT NULL DEFAULT '' COMMENT '用户id',
  `miner_id` varchar(128) NOT NULL DEFAULT '' COMMENT '矿工id',
  `device_id` varchar(128) NOT NULL DEFAULT '' COMMENT '设备id',
  `file_name` varchar(128) NOT NULL DEFAULT '' COMMENT '文件名',
  `ip_address` varchar(32) NOT NULL DEFAULT '' COMMENT '目的地址',
  `cid` varchar(128) NOT NULL DEFAULT '' COMMENT '请求cid',
  `bandwidth_up` varchar(32) NOT NULL DEFAULT '' COMMENT '上行带宽',
  `bandwidth_down` varchar(32) NOT NULL DEFAULT '' COMMENT '下行带宽',
  `time_need` varchar(32) NOT NULL DEFAULT '' COMMENT '期望时间',
  `time` timestamp  NULL DEFAULT NULL COMMENT '时间',
  `service_country` varchar(56) NOT NULL DEFAULT '' COMMENT '服务商国家',
  `region` varchar(56) NOT NULL DEFAULT '' COMMENT '地区',
  `status` varchar(56) NOT NULL DEFAULT '' COMMENT '状态',
  `price` float(32) NOT NULL DEFAULT '0' COMMENT '价格',
  `file_size` float(32) NOT NULL DEFAULT '0' COMMENT '文件大小',
  `download_url` varchar(256) NOT NULL DEFAULT '' COMMENT '下载地址',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `income_daily_r`;
CREATE TABLE `income_daily_r`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  `user_id` varchar(128) NOT NULL DEFAULT '' COMMENT '用户id',
  `device_id` varchar(128) NOT NULL DEFAULT '' COMMENT '设备id',
  `time` timestamp  NULL DEFAULT NULL COMMENT '时间',
  `income` float(32) NOT NULL DEFAULT '0' COMMENT '收益',
  `online_time` float(32) NOT NULL DEFAULT '0' COMMENT '在线时长',
  `pkg_loss_ratio` float(32) NOT NULL DEFAULT '0' COMMENT '丢包率',
  `latency` float(32) NOT NULL DEFAULT '0' COMMENT '延迟',
  `nat_ratio` float(32) NOT NULL DEFAULT '0' COMMENT 'Nat',
  `disk_usage` float(32) NOT NULL DEFAULT '0' COMMENT '磁盘使用率',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_income_daily_r_deleted_at`(`deleted_at` ASC) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
DROP TABLE IF EXISTS `hour_daily_r`;
CREATE TABLE `hour_daily_r`  (
   `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
   `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
   `updated_at` timestamp NULL DEFAULT NULL COMMENT '更新时间',
   `deleted_at` datetime(3) NULL DEFAULT NULL,
   `user_id` varchar(128) NOT NULL DEFAULT '' COMMENT '用户id',
   `device_id` varchar(128) NOT NULL DEFAULT '' COMMENT '设备id',
   `time` timestamp NULL DEFAULT NULL COMMENT '时间',
   `hour_income` float(32) NOT NULL DEFAULT '0' COMMENT '收益',
   `online_time` float(32) NOT NULL DEFAULT '0' COMMENT '在线时长',
   `pkg_loss_ratio` float(32) NOT NULL DEFAULT '0' COMMENT '丢包率',
   `latency` float(32) NOT NULL DEFAULT '0' COMMENT '延迟',
   `nat_ratio` float(32) NOT NULL DEFAULT '0' COMMENT 'Nat',
   `disk_usage` float(32) NOT NULL DEFAULT '0' COMMENT '磁盘使用率',
   PRIMARY KEY (`id`) USING BTREE,
   INDEX `idx_Hour_daily_r_deleted_at`(`deleted_at` ASC) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
