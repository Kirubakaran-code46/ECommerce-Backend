-- Seed initial products
INSERT INTO products (id, name, description, price, currency, stock_quantity)
VALUES
    ('a0000000-0000-0000-0000-000000000001', 'Laptop Pro 16', 'High-performance developer workstation with 32GB RAM and 1TB SSD', 5000000, 'INR', 10),
    ('a0000000-0000-0000-0000-000000000002', 'Mechanical Gaming Keyboard', 'Tactile mechanical switches with customizable RGB backlighting', 300000, 'INR', 25),
    ('a0000000-0000-0000-0000-000000000003', 'Ergonomic Wireless Mouse', 'Precision optical sensor with dual bluetooth and 2.4GHz connectivity', 200000, 'INR', 50),
    ('a0000000-0000-0000-0000-000000000004', 'Ultra-Wide 4K Monitor', '34-inch curved IPS display with USB-C 90W power delivery', 1500000, 'INR', 8),
    ('a0000000-0000-0000-0000-000000000005', 'Noise-Cancelling Headphones', 'Active noise cancelling wireless over-ear headphones with 40h battery life', 500000, 'INR', 15),
    ('a0000000-0000-0000-0000-000000000006', 'USB-C Multiport Adapter', '7-in-1 aluminium hub with 4K HDMI, Gigabit Ethernet, and SD card reader', 150000, 'INR', 40),
    ('a0000000-0000-0000-0000-000000000007', 'Limited Edition Smart Watch', 'Titanium case smart wearable with sapphire glass (Sold Out)', 2500000, 'INR', 0)
ON CONFLICT (id) DO NOTHING;

-- Seed initial promotional coupons
INSERT INTO coupons (id, code, discount_type, discount_value, minimum_cart_value, maximum_discount, expires_at, active)
VALUES
    ('c0000000-0000-0000-0000-000000000001', 'SAVE10', 'percentage', 10, 100000, 500000, '2030-12-31 23:59:59+00', true),
    ('c0000000-0000-0000-0000-000000000002', 'SAVE20', 'percentage', 20, 500000, 300000, '2030-12-31 23:59:59+00', true),
    ('c0000000-0000-0000-0000-000000000003', 'FLAT500', 'fixed_amount', 50000, 200000, NULL, '2030-12-31 23:59:59+00', true),
    ('c0000000-0000-0000-0000-000000000004', 'EXPIRED10', 'percentage', 10, 0, NULL, '2024-01-01 00:00:00+00', true),
    ('c0000000-0000-0000-0000-000000000005', 'INACTIVE15', 'percentage', 15, 0, NULL, '2030-12-31 23:59:59+00', false)
ON CONFLICT (id) DO NOTHING;
