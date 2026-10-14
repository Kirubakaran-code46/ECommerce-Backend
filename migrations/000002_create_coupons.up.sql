CREATE TABLE IF NOT EXISTS coupons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL,
    discount_type VARCHAR(20) NOT NULL CHECK (discount_type IN ('percentage', 'fixed_amount')),
    discount_value BIGINT NOT NULL CHECK (discount_value > 0),
    minimum_cart_value BIGINT NOT NULL DEFAULT 0 CHECK (minimum_cart_value >= 0),
    maximum_discount BIGINT CHECK (maximum_discount IS NULL OR maximum_discount > 0),
    expires_at TIMESTAMPTZ NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_coupons_code_lower ON coupons (LOWER(code));
