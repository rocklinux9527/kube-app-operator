/*
 Navicat MySQL Data Transfer

 Source Server         : 127.0.0.1
 Source Server Type    : MySQL
 Source Server Version : 50744
 Source Host           : 127.0.0.1:3306
 Source Schema         : k8s

 Target Server Type    : MySQL
 Target Server Version : 50744
 File Encoding         : 65001

 Date: 09/11/2025 11:56:21
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` varchar(36) COLLATE utf8mb4_bin NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `email` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `password_hash` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- ----------------------------
-- Records of users
-- ----------------------------
BEGIN;
INSERT INTO `users` VALUES ('74ddc985-0fa8-457f-9e5b-8cf8432dff70', 'zhangsan', 'zhangsan@163.com', '$2a$10$rbTJhxuS4ADq7w7Bi4ngEe75tGFs0FZVT1xqNoupkpHcdNsf6hcsu', '2025-09-30 00:18:18.988', '2025-09-30 22:22:25.932');
INSERT INTO `users` VALUES ('b1639ef6-7252-4095-8e3e-b659eea95677', 'admin', 'admin@example.com', '$2a$10$as1PPneBb9O.YbSb8cqwte2WLMPJ1MbrfRDle9MHUm0KE9DDYv87S', '2025-09-28 21:08:42.813', '2025-09-29 21:12:38.433');
COMMIT;

SET FOREIGN_KEY_CHECKS = 1;
