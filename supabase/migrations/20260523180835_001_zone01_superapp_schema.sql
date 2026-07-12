/*
  # Zone 01 Oujda Super-App Schema

  ## Overview
  This migration creates the database foundation for the Zone 01 Oujda coding school
  super-app. It transforms the existing forum into a comprehensive community platform
  with multiple modules.

  ## New Tables

  ### Module 1: Nexus Transit (Carpooling)
  - `nexus_transit_posts`: Stores ride-sharing offers
    - `id`: Primary key
    - `user_id`: Driver offering the ride
    - `direction`: 'aller' (to campus) or 'retour' (from campus)
    - `departure_time`: When the ride departs
    - `departure_location`: Pickup point description
    - `total_seats`: Hardcoded to 4 seats
    - `available_seats`: Remaining seats (calculated)
    - `status`: 'open', 'full', 'cancelled'
    - `notes`: Additional driver notes
    - `created_at`: Timestamp
  
  - `nexus_transit_bookings`: Passenger bookings for rides
    - `id`: Primary key
    - `post_id`: Reference to the ride post
    - `user_id`: Passenger who booked
    - `booking_order`: Order of booking (1-4) for FIFO
    - `status`: 'confirmed', 'cancelled'
    - `created_at`: Timestamp

  ### Future Modules (预留 tables for specs)
  - Module 2: TBD (will add tables when specs provided)
  - Module 3: TBD
  - Module 4: TBD
  - Module 5: TBD

  ## Security (RLS Enabled)
  - Row Level Security enabled on all tables
  - Policies restrict access based on authentication and ownership
  - Users can only modify their own posts/bookings
  - All authenticated users can view open posts

  ## Notes
  1. The `nexus_transit_posts.total_seats` is enforced at the application level to always be 4
  2. The `available_seats` field is maintained via database triggers
  3. Booking order ensures FIFO enforcement
*/

-- ============================================
-- NEXUS TRANSIT MODULE (Carpooling)
-- ============================================

CREATE TABLE IF NOT EXISTS nexus_transit_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    direction TEXT NOT NULL CHECK (direction IN ('aller', 'retour')),
    departure_time TIMESTAMPTZ NOT NULL,
    departure_location TEXT NOT NULL,
    total_seats INTEGER NOT NULL DEFAULT 4 CHECK (total_seats = 4),
    available_seats INTEGER NOT NULL DEFAULT 4,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'full', 'cancelled')),
    notes TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nexus_transit_bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES nexus_transit_posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    booking_order INTEGER NOT NULL CHECK (booking_order >= 1 AND booking_order <= 4),
    status TEXT NOT NULL DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'cancelled')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(post_id, user_id),
    UNIQUE(post_id, booking_order)
);

-- Create indexes for performance
CREATE INDEX idx_nexus_transit_posts_user_id ON nexus_transit_posts(user_id);
CREATE INDEX idx_nexus_transit_posts_status ON nexus_transit_posts(status);
CREATE INDEX idx_nexus_transit_posts_departure_time ON nexus_transit_posts(departure_time);
CREATE INDEX idx_nexus_transit_bookings_post_id ON nexus_transit_bookings(post_id);
CREATE INDEX idx_nexus_transit_bookings_user_id ON nexus_transit_bookings(user_id);

-- Enable RLS
ALTER TABLE nexus_transit_posts ENABLE ROW LEVEL SECURITY;
ALTER TABLE nexus_transit_bookings ENABLE ROW LEVEL SECURITY;

-- RLS Policies for nexus_transit_posts
CREATE POLICY "Users can view all open transit posts"
    ON nexus_transit_posts FOR SELECT
    TO authenticated
    USING (status = 'open' OR user_id = auth.uid());

CREATE POLICY "Users can view their own transit posts"
    ON nexus_transit_posts FOR SELECT
    TO authenticated
    USING (user_id = auth.uid());

CREATE POLICY "Users can create transit posts"
    ON nexus_transit_posts FOR INSERT
    TO authenticated
    WITH CHECK (user_id = auth.uid() AND total_seats = 4);

CREATE POLICY "Users can update own transit posts"
    ON nexus_transit_posts FOR UPDATE
    TO authenticated
    USING (user_id = auth.uid())
    WITH CHECK (user_id = auth.uid());

CREATE POLICY "Users can delete own transit posts"
    ON nexus_transit_posts FOR DELETE
    TO authenticated
    USING (user_id = auth.uid());

-- RLS Policies for nexus_transit_bookings
CREATE POLICY "Users can view bookings for visible posts"
    ON nexus_transit_bookings FOR SELECT
    TO authenticated
    USING (
        EXISTS (
            SELECT 1 FROM nexus_transit_posts
            WHERE nexus_transit_posts.id = nexus_transit_bookings.post_id
            AND (nexus_transit_posts.status = 'open' OR nexus_transit_posts.user_id = auth.uid())
        )
        OR user_id = auth.uid()
    );

CREATE POLICY "Users can create bookings on open posts"
    ON nexus_transit_bookings FOR INSERT
    TO authenticated
    WITH CHECK (
        user_id = auth.uid()
        AND EXISTS (
            SELECT 1 FROM nexus_transit_posts
            WHERE nexus_transit_posts.id = nexus_transit_bookings.post_id
            AND nexus_transit_posts.status = 'open'
            AND nexus_transit_posts.available_seats > 0
        )
    );

CREATE POLICY "Users can cancel own bookings"
    ON nexus_transit_bookings FOR UPDATE
    TO authenticated
    USING (user_id = auth.uid())
    WITH CHECK (user_id = auth.uid());

CREATE POLICY "Drivers can view bookings on their posts"
    ON nexus_transit_bookings FOR SELECT
    TO authenticated
    USING (
        EXISTS (
            SELECT 1 FROM nexus_transit_posts
            WHERE nexus_transit_posts.id = nexus_transit_bookings.post_id
            AND nexus_transit_posts.user_id = auth.uid()
        )
    );

-- Function to automatically update post status when full
CREATE OR REPLACE FUNCTION update_transit_post_status()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE nexus_transit_posts
    SET available_seats = (
        SELECT 4 - COUNT(*)
        FROM nexus_transit_bookings
        WHERE post_id = NEW.post_id AND status = 'confirmed'
    ),
    status = CASE
        WHEN (
            SELECT 4 - COUNT(*)
            FROM nexus_transit_bookings
            WHERE post_id = NEW.post_id AND status = 'confirmed'
        ) <= 0 THEN 'full'
        ELSE 'open'
    END,
    updated_at = NOW()
    WHERE id = NEW.post_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update available seats after booking
CREATE TRIGGER after_booking_created
    AFTER INSERT OR UPDATE ON nexus_transit_bookings
    FOR EACH ROW
    WHEN (NEW.status = 'confirmed')
    EXECUTE FUNCTION update_transit_post_status();

-- Function to handle booking cancellation
CREATE OR REPLACE FUNCTION handle_booking_cancellation()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE nexus_transit_posts
    SET available_seats = (
        SELECT 4 - COUNT(*)
        FROM nexus_transit_bookings
        WHERE post_id = OLD.post_id AND status = 'confirmed'
    ),
    status = 'open',
    updated_at = NOW()
    WHERE id = OLD.post_id;
    
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Trigger for cancellation
CREATE TRIGGER after_booking_cancelled
    AFTER UPDATE ON nexus_transit_bookings
    FOR EACH ROW
    WHEN (OLD.status = 'confirmed' AND NEW.status = 'cancelled')
    EXECUTE FUNCTION handle_booking_cancellation();