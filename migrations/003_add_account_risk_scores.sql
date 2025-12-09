-- Migration: Add ML prediction tables for tens-insight integration
-- Stores ML predictions for account churn risk and health scores

-- Create account_risk_scores table
CREATE TABLE IF NOT EXISTS account_risk_scores (
    account_id VARCHAR PRIMARY KEY,
    churn_probability FLOAT NOT NULL CHECK (churn_probability >= 0 AND churn_probability <= 1),
    health_score FLOAT NOT NULL CHECK (health_score >= 0 AND health_score <= 100),
    risk_category VARCHAR NOT NULL CHECK (risk_category IN ('low', 'medium', 'high', 'critical')),
    predicted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    model_version VARCHAR NOT NULL
);

-- Create indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_account_risk_category ON account_risk_scores(risk_category);
CREATE INDEX IF NOT EXISTS idx_account_predicted_at ON account_risk_scores(predicted_at DESC);
CREATE INDEX IF NOT EXISTS idx_account_health_score ON account_risk_scores(health_score);

-- Add comments for documentation
COMMENT ON TABLE account_risk_scores IS 'ML predictions for account churn risk and health scores (populated by tens-insight)';
COMMENT ON COLUMN account_risk_scores.churn_probability IS 'Predicted probability of account churn (0-1)';
COMMENT ON COLUMN account_risk_scores.health_score IS 'Account health score (0-100, inverse of churn probability)';
COMMENT ON COLUMN account_risk_scores.risk_category IS 'Risk category: low (<25%), medium (25-50%), high (50-75%), critical (>75%)';
