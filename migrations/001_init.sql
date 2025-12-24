-- Create the feedback_enriched table
CREATE TABLE IF NOT EXISTS feedback_enriched (
    id            TEXT PRIMARY KEY,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    source        TEXT NOT NULL,
    product_area  TEXT NOT NULL,
    sentiment     TEXT NOT NULL CHECK (sentiment IN ('positive', 'neutral', 'negative')),
    priority      INT NOT NULL CHECK (priority >= 1 AND priority <= 5),
    topic         TEXT NOT NULL,
    region        TEXT NOT NULL,
    customer_tier TEXT NOT NULL,
    account_id    VARCHAR,
    summary       TEXT NOT NULL
);

-- Create indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_feedback_product_area ON feedback_enriched(product_area);
CREATE INDEX IF NOT EXISTS idx_feedback_sentiment ON feedback_enriched(sentiment);
CREATE INDEX IF NOT EXISTS idx_feedback_priority ON feedback_enriched(priority);
CREATE INDEX IF NOT EXISTS idx_feedback_created_at ON feedback_enriched(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_feedback_customer_tier ON feedback_enriched(customer_tier);
CREATE INDEX IF NOT EXISTS idx_feedback_region ON feedback_enriched(region);
CREATE INDEX IF NOT EXISTS idx_feedback_account_id ON feedback_enriched(account_id);
