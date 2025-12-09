-- Migration: Add ML prediction tables for tens-insight integration
-- Stores ML predictions for product area priority scores by segment

-- Create product_area_impact table
CREATE TABLE IF NOT EXISTS product_area_impact (
    product_area VARCHAR NOT NULL,
    segment VARCHAR NOT NULL,
    priority_score FLOAT NOT NULL CHECK (priority_score >= 0 AND priority_score <= 100),
    feedback_count INTEGER NOT NULL CHECK (feedback_count >= 0),
    avg_sentiment_score FLOAT NOT NULL CHECK (avg_sentiment_score >= -1 AND avg_sentiment_score <= 1),
    negative_count INTEGER NOT NULL CHECK (negative_count >= 0),
    critical_count INTEGER NOT NULL CHECK (critical_count >= 0),
    predicted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    model_version VARCHAR NOT NULL,
    PRIMARY KEY (product_area, segment)
);

-- Create indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_product_area_priority ON product_area_impact(priority_score DESC);
CREATE INDEX IF NOT EXISTS idx_product_area_predicted_at ON product_area_impact(predicted_at DESC);
CREATE INDEX IF NOT EXISTS idx_product_area_segment ON product_area_impact(segment);

-- Add comments for documentation
COMMENT ON TABLE product_area_impact IS 'ML predictions for product area priority scores by segment (populated by tens-insight)';
COMMENT ON COLUMN product_area_impact.priority_score IS 'Priority score for this product area/segment combination (0-100)';
COMMENT ON COLUMN product_area_impact.avg_sentiment_score IS 'Average sentiment score (-1 to 1, negative to positive)';
