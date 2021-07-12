-- Create syntax for TABLE 'balances'
CREATE TABLE `balances` (
  `address` char(1) NOT NULL DEFAULT '',
  `eth_balance` decimal(36,18) unsigned NOT NULL,
  PRIMARY KEY (`address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Create syntax for TABLE 'sync_info'
CREATE TABLE `sync_info` (
  `block_num` int DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Create syntax for TABLE 'token_transfers'
CREATE TABLE `token_transfers` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `token_address` char(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
  `from_address` char(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
  `to_address` char(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
  `transaction_hash` char(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
  `value` decimal(36,18) NOT NULL,
  `block_timestamp` timestamp NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Create syntax for TABLE 'tokens'
CREATE TABLE `tokens` (
  `adress` char(64) NOT NULL DEFAULT '',
  `symbol` char(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  `name` char(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '',
  `decimals` tinyint NOT NULL,
  `total_supply` tinyint NOT NULL,
  `block_timestamp` timestamp NOT NULL,
  PRIMARY KEY (`adress`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Create syntax for TABLE 'transactions'
CREATE TABLE `transactions` (
  `hash` char(64) NOT NULL DEFAULT '',
  `from_address` char(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
  `to_address` char(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
  `value` decimal(36,18) NOT NULL,
  `gas` int unsigned NOT NULL,
  `gas_price` int unsigned NOT NULL,
  `block_timestamp` timestamp NOT NULL,
  PRIMARY KEY (`hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
