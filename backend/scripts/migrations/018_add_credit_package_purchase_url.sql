ALTER TABLE `credit_packages`
  ADD COLUMN `purchase_url` VARCHAR(500) NOT NULL DEFAULT '' AFTER `note`;
