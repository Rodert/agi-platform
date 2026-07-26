-- Redeem codes already exist in the baseline schema. This index supports the
-- management-list filters that distinguish unused, redeemed and expired codes.
ALTER TABLE `redeem_codes`
  ADD KEY `idx_redeem_code_status` (`used_by`, `expires_at`);
