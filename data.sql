-- Sample Users (User 3 is soft-deleted)
INSERT INTO users (id, name, email, password_hash, email_verified_at, deleted_at, created_at, updated_at) VALUES
(1, 'Alice Admin', 'alice@example.com', '$2y$10$...', '2025-04-01 10:00:00', NULL, '2025-04-01 09:00:00', '2025-04-01 10:00:00'),
(2, 'Bob Developer', 'bob@example.com', '$2y$10$...', '2025-04-01 11:00:00', NULL, '2025-04-01 11:00:00', '2025-04-01 11:00:00'),
(3, 'Charlie Deleted', 'charlie_deleted_20250402191500@example.com', '$2y$10$...', NULL, '2025-04-02 19:15:00', '2025-04-01 12:00:00', '2025-04-02 19:15:00'); -- Note: Email modified for uniqueness on soft delete

-- Sample API Keys
INSERT INTO api_keys (id, user_id, key_value, label, is_active, deleted_at, created_at, updated_at) VALUES
(1, 1, 'akey_alice_prod_dk38f...', 'Alice Production Key', TRUE, NULL, '2025-04-01 10:05:00', '2025-04-01 10:05:00'),
(2, 2, 'akey_bob_dev_j9s7a...', 'Bob Dev Server', TRUE, NULL, '2025-04-01 11:05:00', '2025-04-01 11:05:00'),
(3, 1, 'akey_alice_old_h4g5f...', 'Alice Old Key', FALSE, '2025-04-02 09:30:00', '2025-03-15 08:00:00', '2025-04-02 09:30:00'); -- Soft deleted API key

-- Sample Checks (Variety of states)
INSERT INTO checks (id, user_id, uuid, name, description, expected_interval, grace_period, last_ping_at, status, is_enabled, deleted_at, created_at, updated_at) VALUES
(1, 1, 'd1b0a7e8-f17e-4a76-8a2f-4e1e7f8a2b1c', 'Hourly Backup Verification', 'Checks if backup script ran', 3600, 300, '2025-04-02 19:10:00', 'up', TRUE, NULL, '2025-04-01 14:00:00', '2025-04-02 19:10:00'), -- User 1, Up, Pinged recently
(2, 1, 'a7c8e0f2-a3b4-4c5d-8e6f-1a2b3c4d5e7f', 'Daily Cron Job', 'Monitors the main daily task', 86400, 600, '2025-03-31 08:30:00', 'down', TRUE, NULL, '2025-04-01 14:05:00', '2025-03-31 08:30:00'), -- User 1, Down, Last ping > 1 day ago
(3, 1, 'b3d9f1a0-b4c5-4d6e-9f7a-2b3c4d5e6f8a', 'Monthly Report Gen', 'Generates monthly report', 2592000, 86400, '2025-04-01 05:00:00', 'up', FALSE, NULL, '2025-04-01 14:10:00', '2025-04-01 05:00:00'), -- User 1, Paused, Was up
(4, 2, 'c9e1a2b3-c4d5-4e6f-8a7b-3c4d5e6f7a8b', 'Dev Server Heartbeat', NULL, 900, 60, '2025-04-02 19:05:00', 'up', TRUE, NULL, '2025-04-01 15:00:00', '2025-04-02 19:05:00'), -- User 2, Up, 15 min interval, pinged recently
(5, 2, 'e5f1b3c4-d5e6-4f7a-8b9c-4d5e6f7a8b9c', 'Staging Deployment Check', 'Monitors staging deployment script', 1800, 120, NULL, 'new', TRUE, NULL, '2025-04-02 10:00:00', '2025-04-02 10:00:00'), -- User 2, New, Never pinged
(6, 1, 'f0a1b2c3-d4e5-4f6a-8b7c-5d6e7f8a9b0d', 'Old Test Check', 'This check was deleted', 60, 10, '2025-04-01 18:00:00', 'up', TRUE, '2025-04-02 11:00:00', '2025-04-01 17:55:00', '2025-04-02 11:00:00'); -- User 1, Soft Deleted

-- Sample Pings (Link to checks, provide some history)
INSERT INTO pings (check_id, received_at, source_ip, user_agent, payload, created_at) VALUES
(1, '2025-04-02 19:10:00', '192.168.1.100', 'curl/7.68.0', '{"status": "OK", "files_processed": 1234}', '2025-04-02 19:10:00'), -- Recent ping for check 1
(1, '2025-04-02 18:09:55', '192.168.1.100', 'curl/7.68.0', '{"status": "OK", "files_processed": 1230}', '2025-04-02 18:09:55'), -- Older ping for check 1
(2, '2025-03-31 08:30:00', '10.0.0.5', 'Python App Monitor v1.1', NULL, '2025-03-31 08:30:00'), -- Last ping for check 2 (now down)
(3, '2025-04-01 05:00:00', '172.16.0.10', 'Reporting Script', '{"report_id": "MAR2025", "duration_ms": 55000}', '2025-04-01 05:00:00'), -- Last ping for check 3 (now paused)
(4, '2025-04-02 19:05:00', '192.168.5.55', 'curl/7.68.0', NULL, '2025-04-02 19:05:00'), -- Recent ping for check 4
(4, '2025-04-02 18:50:00', '192.168.5.55', 'curl/7.68.0', NULL, '2025-04-02 18:50:00'), -- Older ping for check 4
(6, '2025-04-01 18:00:00', '127.0.0.1', 'Test Runner', NULL, '2025-04-01 18:00:00'); -- Last ping for check 6 (now deleted)


-- Sample Notification Channels
INSERT INTO notification_channels (id, user_id, type, value, label, is_verified, verification_token, is_enabled, deleted_at, created_at, updated_at) VALUES
(1, 1, 'email', 'alice@example.com', 'Primary Email', TRUE, NULL, TRUE, NULL, '2025-04-01 09:00:00', '2025-04-01 10:00:00'), -- User 1, Email, Verified
(2, 1, 'webhook', 'https://hooks.example.com/T123/B456/C789', 'Team Alert Hook', FALSE, 'verify_hook_abc123', TRUE, NULL, '2025-04-01 14:30:00', '2025-04-01 14:30:00'), -- User 1, Webhook, Unverified
(3, 2, 'webhook', 'https://hooks.slack.com/services/X1/Y2/Z3', 'Bob Slack', TRUE, NULL, TRUE, NULL, '2025-04-01 11:10:00', '2025-04-01 11:15:00'), -- User 2, Slack (as webhook), Verified
(4, 2, 'email', 'bob.secondary@example.net', 'Secondary Email', FALSE, 'verify_email_def456', FALSE, NULL, '2025-04-01 16:00:00', '2025-04-01 16:05:00'); -- User 2, Email, Unverified, Disabled

-- Sample Assignments of Channels to Checks
INSERT INTO check_notification_channel (check_id, notification_channel_id) VALUES
(1, 1), -- Check 1 (Hourly Backup) notifies Alice's Email
(2, 1), -- Check 2 (Daily Cron) notifies Alice's Email
(2, 2), -- Check 2 (Daily Cron) also notifies Team Alert Hook (if it were verified/enabled)
(4, 3); -- Check 4 (Dev Heartbeat) notifies Bob's Slack

-- Sample Notification Logs
INSERT INTO notifications_log (check_id, notification_channel_id, notification_type, status, attempted_at, error_message) VALUES
(2, 1, 'down', 'sent', '2025-04-01 09:31:00', NULL), -- Notified Alice via email that Check 2 went down yesterday
(2, 2, 'down', 'failed', '2025-04-01 09:31:05', 'Webhook endpoint returned 404'), -- Attempted webhook for Check 2 failed
(1, 1, 'up', 'sent', '2025-03-28 10:05:00', NULL); -- Example 'up' notification from the past