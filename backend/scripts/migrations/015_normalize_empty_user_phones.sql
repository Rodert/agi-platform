-- Empty strings conflict with the unique phone index. Unbound phones must be NULL.
UPDATE `users` SET `phone` = NULL WHERE `phone` = '';
