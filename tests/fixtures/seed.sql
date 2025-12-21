-- Test fixtures for goinsight

-- Clear existing test data
DELETE FROM feedback_enriched WHERE source = 'test';
DELETE FROM product_area_impact WHERE model_version = 'test_v1';
DELETE FROM account_risk_scores WHERE model_version = 'test_v1';

-- Insert test feedback data
INSERT INTO feedback (id, source, feedback, sentiment, created_at) VALUES
('test_1', 'email', 'Excellent product quality', 'positive', NOW()),
('test_2', 'email', 'Great support experience', 'positive', NOW()),
('test_3', 'chat', 'Poor response time', 'negative', NOW()),
('test_4', 'chat', 'Pricing is too high', 'negative', NOW()),
('test_5', 'survey', 'Average features', 'neutral', NOW());

-- Insert test enriched feedback
INSERT INTO feedback_enriched (
  id, source, feedback, sentiment, product_area, priority,
  topic, region, customer_tier, summary, created_at
) VALUES
('enriched_1', 'email', 'Excellent product', 'positive', 'billing', 2,
 'pricing', 'US', 'enterprise', 'Customer satisfied with product', NOW()),
('enriched_2', 'email', 'Support is slow', 'negative', 'support', 1,
 'response_time', 'EU', 'pro', 'Need faster support response', NOW()),
('enriched_3', 'chat', 'Love the features', 'positive', 'features', 3,
 'feature_request', 'APAC', 'standard', 'Customer wants more features', NOW()),
('enriched_4', 'survey', 'Could be better', 'neutral', 'general', 4,
 'general_feedback', 'US', 'starter', 'Mixed feedback from user', NOW()),
('enriched_5', 'chat', 'Integration fails', 'negative', 'integration', 1,
 'integration', 'EU', 'enterprise', 'Critical: integration broken', NOW());

-- Insert test account risk scores
INSERT INTO account_risk_scores (
  account_id, churn_probability, health_score, risk_category, predicted_at, model_version
) VALUES
('account_1', 0.15, 0.85, 'low', NOW(), 'test_v1'),
('account_2', 0.45, 0.55, 'medium', NOW(), 'test_v1'),
('account_3', 0.75, 0.25, 'high', NOW(), 'test_v1'),
('account_4', 0.85, 0.15, 'critical', NOW(), 'test_v1'),
('account_5', 0.25, 0.75, 'low', NOW(), 'test_v1');

-- Insert test product area impacts
INSERT INTO product_area_impact (
  product_area, segment, priority_score, feedback_count,
  avg_sentiment_score, negative_count, critical_count, predicted_at, model_version
) VALUES
('billing', 'enterprise', 0.85, 15, 0.45, 8, 2, NOW(), 'test_v1'),
('support', 'enterprise', 0.90, 20, 0.40, 12, 3, NOW(), 'test_v1'),
('features', 'pro', 0.65, 10, 0.60, 3, 0, NOW(), 'test_v1'),
('integration', 'enterprise', 0.95, 25, 0.35, 18, 5, NOW(), 'test_v1'),
('general', 'standard', 0.40, 5, 0.70, 1, 0, NOW(), 'test_v1');

-- Commit transaction
COMMIT;
