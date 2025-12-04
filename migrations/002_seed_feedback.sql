-- Seed sample feedback data for testing and development
INSERT INTO feedback_enriched (id, created_at, source, product_area, sentiment, priority, topic, region, customer_tier, summary) VALUES
-- Billing issues
('fb-001', NOW() - INTERVAL '1 day', 'zendesk', 'billing', 'negative', 5, 'refund processing', 'NA', 'enterprise', 'Customer unable to process refund, blocking their quarterly reconciliation'),
('fb-002', NOW() - INTERVAL '2 days', 'zendesk', 'billing', 'negative', 4, 'invoice errors', 'EU', 'pro', 'Invoice showing incorrect amounts for subscription upgrade'),
('fb-003', NOW() - INTERVAL '3 days', 'nps_survey', 'billing', 'neutral', 3, 'payment methods', 'NA', 'pro', 'Would like to see more payment options like PayPal'),
('fb-004', NOW() - INTERVAL '5 days', 'zendesk', 'billing', 'negative', 5, 'refund processing', 'APAC', 'enterprise', 'Critical: Refund delayed for over 30 days, threatening to escalate'),
('fb-005', NOW() - INTERVAL '7 days', 'google_play', 'billing', 'negative', 4, 'subscription cancellation', 'EU', 'free', 'Unable to cancel subscription through app, had to contact support'),

-- Onboarding issues
('fb-006', NOW() - INTERVAL '1 day', 'nps_survey', 'onboarding', 'positive', 2, 'setup wizard', 'NA', 'pro', 'Setup wizard was very intuitive and easy to follow'),
('fb-007', NOW() - INTERVAL '2 days', 'zendesk', 'onboarding', 'negative', 3, 'missing documentation', 'EU', 'pro', 'Could not find documentation on how to import existing data'),
('fb-008', NOW() - INTERVAL '3 days', 'google_play', 'onboarding', 'neutral', 2, 'tutorial completion', 'APAC', 'free', 'Tutorial is helpful but a bit too long'),
('fb-009', NOW() - INTERVAL '4 days', 'zendesk', 'onboarding', 'negative', 4, 'data import', 'NA', 'enterprise', 'Bulk data import feature is broken, preventing team onboarding'),
('fb-010', NOW() - INTERVAL '6 days', 'nps_survey', 'onboarding', 'positive', 1, 'quick start guide', 'EU', 'free', 'Quick start guide helped me get up and running in minutes'),

-- Performance issues
('fb-011', NOW() - INTERVAL '1 day', 'zendesk', 'performance', 'negative', 5, 'app crashes', 'NA', 'enterprise', 'App crashes when loading large datasets, making it unusable'),
('fb-012', NOW() - INTERVAL '2 days', 'google_play', 'performance', 'negative', 4, 'slow load times', 'APAC', 'pro', 'Dashboard takes over 30 seconds to load'),
('fb-013', NOW() - INTERVAL '3 days', 'zendesk', 'performance', 'negative', 4, 'memory usage', 'EU', 'enterprise', 'Application consuming excessive memory, causing system slowdown'),
('fb-014', NOW() - INTERVAL '4 days', 'nps_survey', 'performance', 'neutral', 3, 'mobile responsiveness', 'NA', 'free', 'Mobile version is slower than desktop'),
('fb-015', NOW() - INTERVAL '5 days', 'google_play', 'performance', 'negative', 5, 'app crashes', 'APAC', 'pro', 'App crashes immediately on launch after latest update'),

-- Feature requests
('fb-016', NOW() - INTERVAL '1 day', 'nps_survey', 'features', 'positive', 2, 'export functionality', 'NA', 'enterprise', 'Love the new CSV export feature, saves us hours'),
('fb-017', NOW() - INTERVAL '2 days', 'zendesk', 'features', 'neutral', 3, 'api access', 'EU', 'pro', 'Would really benefit from API access for automation'),
('fb-018', NOW() - INTERVAL '3 days', 'nps_survey', 'features', 'neutral', 2, 'dark mode', 'APAC', 'free', 'Dark mode would be a nice addition'),
('fb-019', NOW() - INTERVAL '4 days', 'zendesk', 'features', 'neutral', 3, 'collaboration tools', 'NA', 'enterprise', 'Need better team collaboration features like comments'),
('fb-020', NOW() - INTERVAL '6 days', 'google_play', 'features', 'positive', 2, 'notifications', 'EU', 'pro', 'Push notifications for important events are very useful'),

-- Security concerns
('fb-021', NOW() - INTERVAL '1 day', 'zendesk', 'security', 'negative', 5, 'data breach concerns', 'EU', 'enterprise', 'Need SOC2 compliance documentation urgently for audit'),
('fb-022', NOW() - INTERVAL '2 days', 'zendesk', 'security', 'negative', 4, 'authentication', 'NA', 'enterprise', 'Require SSO integration for our organization'),
('fb-023', NOW() - INTERVAL '3 days', 'nps_survey', 'security', 'neutral', 3, 'two-factor auth', 'APAC', 'pro', 'Would feel more secure with 2FA option'),
('fb-024', NOW() - INTERVAL '5 days', 'zendesk', 'security', 'negative', 5, 'access controls', 'EU', 'enterprise', 'Lack of role-based permissions is a security risk'),

-- UI/UX feedback
('fb-025', NOW() - INTERVAL '1 day', 'nps_survey', 'ui_ux', 'positive', 1, 'design improvements', 'NA', 'pro', 'New dashboard design looks great and is more intuitive'),
('fb-026', NOW() - INTERVAL '2 days', 'google_play', 'ui_ux', 'negative', 3, 'navigation', 'EU', 'free', 'Navigation menu is confusing, hard to find features'),
('fb-027', NOW() - INTERVAL '3 days', 'nps_survey', 'ui_ux', 'neutral', 2, 'color scheme', 'APAC', 'pro', 'Color scheme could be improved for better contrast'),
('fb-028', NOW() - INTERVAL '4 days', 'zendesk', 'ui_ux', 'negative', 3, 'mobile layout', 'NA', 'free', 'Mobile layout is cramped and hard to use'),
('fb-029', NOW() - INTERVAL '6 days', 'nps_survey', 'ui_ux', 'positive', 1, 'accessibility', 'EU', 'pro', 'Appreciate the accessibility improvements in latest release'),

-- Integration issues
('fb-030', NOW() - INTERVAL '1 day', 'zendesk', 'integrations', 'negative', 4, 'slack integration', 'NA', 'enterprise', 'Slack integration not syncing properly'),
('fb-031', NOW() - INTERVAL '2 days', 'zendesk', 'integrations', 'negative', 4, 'jira sync', 'EU', 'enterprise', 'JIRA sync failing intermittently'),
('fb-032', NOW() - INTERVAL '3 days', 'nps_survey', 'integrations', 'neutral', 3, 'zapier support', 'APAC', 'pro', 'Would like to see Zapier integration'),
('fb-033', NOW() - INTERVAL '5 days', 'zendesk', 'integrations', 'negative', 5, 'api rate limits', 'NA', 'enterprise', 'API rate limits too restrictive for our use case');
