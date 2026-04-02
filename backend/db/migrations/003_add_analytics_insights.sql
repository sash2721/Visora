-- User Analytics table (cached analytics computed after each receipt upload)
CREATE TABLE IF NOT EXISTS user_analytics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total_spent NUMERIC(12, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(10),
    category_breakdown JSONB NOT NULL DEFAULT '[]',
    daily_spending JSONB NOT NULL DEFAULT '[]',
    period VARCHAR(20) NOT NULL DEFAULT 'monthly',
    computed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, period)
);

-- User Insights table (cached LLM-generated insights computed after each receipt upload)
CREATE TABLE IF NOT EXISTS user_insights (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    summary TEXT NOT NULL DEFAULT '',
    warnings JSONB NOT NULL DEFAULT '[]',
    period VARCHAR(20) NOT NULL DEFAULT 'monthly',
    computed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, period)
);
