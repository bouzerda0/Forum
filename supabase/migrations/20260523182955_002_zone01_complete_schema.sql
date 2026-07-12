/*
  # Complete Zone 01 Super-App Schema (Modules 2-5)

  ## Overview
  This migration adds database tables for all 5 modules of the Zone 01 Oujda Super-App.

  ## Module 2: Nexus Habitat (Roommate Matching)
  - `nexus_habitat_posts`: Roommate search posts
    - Lifestyle tags/rules (sleep schedule, cleanliness, etc.)
    - Available spots (1-10)
    - Manual accept/reject workflow
  - `nexus_habitat_requests`: Roommate requests with status tracking

  ## Module 3: Nexus Pulse (Food Orders & Events)
  - `nexus_pulse_food_posts`: Food order coordination (max 5 users)
  - `nexus_pulse_food_joins`: Users joining food orders
  - `nexus_pulse_events`: Campus event organization

  ## Module 4: Nexus Forum (Memes & Polls)
  - `nexus_forum_memes`: Meme posts with image uploads (NO ANONYMITY)
  - `nexus_forum_polls`: Community polls with multiple choice
  - `nexus_forum_poll_options`: Poll options
  - `nexus_forum_poll_votes`: Individual poll votes

  ## Module 5: Nexus Spirit & Arena
  - `nexus_arena_matches`: Sports match coordination (ONE per day limit)
  - `nexus_arena_participants`: Match participants
  - Prayer times fetched from external API (no table needed)

  ## Security (RLS Enabled)
  - Row Level Security on all tables
  - Strict ownership and authentication checks
  - No anonymous posts allowed
*/

-- ============================================
-- MODULE 2: NEXUS HABITAT (Roommate Matching)
-- ============================================

CREATE TABLE IF NOT EXISTS nexus_habitat_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    location TEXT NOT NULL,
    available_spots INTEGER NOT NULL CHECK (available_spots >= 1 AND available_spots <= 10),
    filled_spots INTEGER NOT NULL DEFAULT 0,
    rent_cost DECIMAL(10,2),
    lifestyle_tags TEXT[] NOT NULL DEFAULT '{}',
    rules TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed', 'filled')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nexus_habitat_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES nexus_habitat_posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    message TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ,
    UNIQUE(post_id, user_id)
);

CREATE INDEX idx_habitat_posts_user_id ON nexus_habitat_posts(user_id);
CREATE INDEX idx_habitat_posts_status ON nexus_habitat_posts(status);
CREATE INDEX idx_habitat_requests_post_id ON nexus_habitat_requests(post_id);
CREATE INDEX idx_habitat_requests_user_id ON nexus_habitat_requests(user_id);

ALTER TABLE nexus_habitat_posts ENABLE ROW LEVEL SECURITY;
ALTER TABLE nexus_habitat_requests ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view open habitat posts"
    ON nexus_habitat_posts FOR SELECT
    TO authenticated
    USING (status = 'open' OR user_id = auth.uid());

CREATE POLICY "Users can create habitat posts"
    ON nexus_habitat_posts FOR INSERT
    TO authenticated
    WITH CHECK (user_id = auth.uid());

CREATE POLICY "Users can update own habitat posts"
    ON nexus_habitat_posts FOR UPDATE
    TO authenticated
    USING (user_id = auth.uid())
    WITH CHECK (user_id = auth.uid());

CREATE POLICY "Users can delete own habitat posts"
    ON nexus_habitat_posts FOR DELETE
    TO authenticated
    USING (user_id = auth.uid());

CREATE POLICY "Post owners can view all requests"
    ON nexus_habitat_requests FOR SELECT
    TO authenticated
    USING (
        EXISTS (
            SELECT 1 FROM nexus_habitat_posts
            WHERE nexus_habitat_posts.id = nexus_habitat_requests.post_id
            AND nexus_habitat_posts.user_id = auth.uid()
        )
        OR user_id = auth.uid()
    );

CREATE POLICY "Users can create requests on open posts"
    ON nexus_habitat_requests FOR INSERT
    TO authenticated
    WITH CHECK (
        user_id = auth.uid()
        AND EXISTS (
            SELECT 1 FROM nexus_habitat_posts
            WHERE nexus_habitat_posts.id = nexus_habitat_requests.post_id
            AND nexus_habitat_posts.status = 'open'
        )
    );

CREATE POLICY "Post owners can update requests"
    ON nexus_habitat_requests FOR UPDATE
    TO authenticated
    USING (
        EXISTS (
            SELECT 1 FROM nexus_habitat_posts
            WHERE nexus_habitat_posts.id = nexus_habitat_requests.post_id
            AND nexus_habitat_posts.user_id = auth.uid()
        )
    );

-- ============================================
-- MODULE 3: NEXUS PULSE (Food Orders & Events)
-- ============================================

-- Food Orders (Max 5 users per order)
CREATE TABLE IF NOT EXISTS nexus_pulse_food_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    restaurant_name TEXT NOT NULL,
    restaurant_link TEXT,
    description TEXT NOT NULL,
    deadline_time TIMESTAMPTZ NOT NULL,
    max_participants INTEGER NOT NULL DEFAULT 5 CHECK (max_participants = 5),
    current_participants INTEGER NOT NULL DEFAULT 1,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'locked', 'completed', 'cancelled')),
    delivery_fee_split BOOLEAN DEFAULT true,
    notes TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nexus_pulse_food_joins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES nexus_pulse_food_posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    order_details TEXT NOT NULL,
    estimated_cost DECIMAL(10,2),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(post_id, user_id)
);

-- Events
CREATE TABLE IF NOT EXISTS nexus_pulse_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    event_type TEXT NOT NULL CHECK (event_type IN ('movie', 'board_games', 'sports', 'study_group', 'party', 'other')),
    event_date TIMESTAMPTZ NOT NULL,
    location TEXT NOT NULL,
    max_attendees INTEGER,
    current_attendees INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'upcoming' CHECK (status IN ('upcoming', 'ongoing', 'completed', 'cancelled')),
    image_url TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nexus_pulse_event_attendees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES nexus_pulse_events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'attending' CHECK (status IN ('attending', 'maybe', 'not_attending')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(event_id, user_id)
);

CREATE INDEX idx_food_posts_user_id ON nexus_pulse_food_posts(user_id);
CREATE INDEX idx_food_posts_status ON nexus_pulse_food_posts(status);
CREATE INDEX idx_food_joins_post_id ON nexus_pulse_food_joins(post_id);
CREATE INDEX idx_events_user_id ON nexus_pulse_events(user_id);
CREATE INDEX idx_events_date ON nexus_pulse_events(event_date);
CREATE INDEX idx_event_attendees_event_id ON nexus_pulse_event_attendees(event_id);

ALTER TABLE nexus_pulse_food_posts ENABLE ROW LEVEL SECURITY;
ALTER TABLE nexus_pulse_food_joins ENABLE ROW LEVEL SECURITY;
ALTER TABLE nexus_pulse_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE nexus_pulse_event_attendees ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view open food posts"
    ON nexus_pulse_food_posts FOR SELECT
    TO authenticated
    USING (status IN ('open', 'locked') OR user_id = auth.uid());

CREATE POLICY "Users can create food posts"
    ON nexus_pulse_food_posts FOR INSERT
    TO authenticated
    WITH CHECK (user_id = auth.uid());

CREATE POLICY "Users can update own food posts"
    ON nexus_pulse_food_posts FOR UPDATE
    TO authenticated
    USING (user_id = auth.uid())
    WITH CHECK (user_id = auth.uid());

CREATE POLICY "Users can join food posts"
    ON nexus_pulse_food_joins FOR INSERT
    TO authenticated
    WITH CHECK (
        user_id = auth.uid()
        AND EXISTS (
            SELECT 1 FROM nexus_pulse_food_posts
            WHERE nexus_pulse_food_posts.id = nexus_pulse_food_joins.post_id
            AND nexus_pulse_food_posts.status = 'open'
            AND nexus_pulse_food_posts.current_participants < 5
        )
    );

CREATE POLICY "Users can view events"
    ON nexus_pulse_events FOR SELECT
    TO authenticated
    USING (true);

CREATE POLICY "Users can create events"
    ON nexus_pulse_events FOR INSERT
    TO authenticated
    WITH CHECK (user_id = auth.uid());

CREATE POLICY "Users can update own events"
    ON nexus_pulse_events FOR UPDATE
    TO authenticated
    USING (user_id = auth.uid());

CREATE POLICY "Users can attend events"
    ON nexus_pulse_event_attendees FOR INSERT
    TO authenticated
    WITH CHECK (user_id = auth.uid());

-- ============================================
-- MODULE 4: NEXUS FORUM (Memes & Polls)
-- ============================================

-- Memes (NO ANONYMITY - user_id always visible)
CREATE TABLE IF NOT EXISTS nexus_forum_memes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    image_url TEXT NOT NULL,
    description TEXT DEFAULT '',
    likes_count INTEGER NOT NULL DEFAULT 0,
    comments_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nexus_forum_meme_likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    meme_id UUID NOT NULL REFERENCES nexus_forum_memes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(meme_id, user_id)
);

CREATE TABLE IF NOT EXISTS nexus_forum_meme_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    meme_id UUID NOT NULL REFERENCES nexus_forum_memes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Polls
CREATE TABLE IF NOT EXISTS nexus_forum_polls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    allow_multiple_choices BOOLEAN DEFAULT false,
    closes_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'closed')),
    total_votes INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nexus_forum_poll_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    poll_id UUID NOT NULL REFERENCES nexus_forum_polls(id) ON DELETE CASCADE,
    option_text TEXT NOT NULL,
    vote_count INTEGER NOT NULL DEFAULT 0,
    display_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS nexus_forum_poll_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    poll_id UUID NOT NULL REFERENCES nexus_forum_polls(id) ON DELETE CASCADE,
    option_id UUID NOT NULL REFERENCES nexus_forum_poll_options(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(option_id, user_id)
);

CREATE INDEX idx_memes_user_id ON nexus_forum_memes(user_id);
CREATE INDEX idx_memes_created_at ON nexus_forum_memes(created_at DESC);
CREATE INDEX idx_meme_likes_meme_id ON nexus_forum_meme_likes(meme_id);
CREATE INDEX idx_meme_comments_meme_id ON nexus_forum_meme_comments(meme_id);
CREATE INDEX idx_polls_user_id ON nexus_forum_polls(user_id);
CREATE INDEX idx_poll_options_poll_id ON nexus_forum_poll_options(poll_id);
CREATE INDEX idx_poll_votes_poll_id ON nexus_forum_poll_votes(poll_id);

ALTER TABLE nexus_forum_memes ENABLE ROW LEVEL SECURITY;
ALTER TABLE nexus_forum_meme_likes ENABLE ROW LEVEL SECURITY;
ALTER TABLE nexus_forum_meme_comments ENABLE ROW LEVEL SECURITY;
ALTER TABLE nexus_forum_polls ENABLE ROW LEVEL SECURITY;
ALTER TABLE nexus_forum_poll_options ENABLE ROW LEVEL SECURITY;
ALTER TABLE nexus_forum_poll_votes ENABLE ROW LEVEL SECURITY;

CREATE POLICY "All authenticated users can view memes"
    ON nexus_forum_memes FOR SELECT
    TO authenticated
    USING (true);

CREATE POLICY "Users can create memes with identity"
    ON nexus_forum_memes FOR INSERT
    TO authenticated
    WITH CHECK (user_id = auth.uid());

CREATE POLICY "Users can like memes"
    ON nexus_forum_meme_likes FOR INSERT
    TO authenticated
    WITH CHECK (user_id = auth.uid());

CREATE POLICY "Users can comment on memes"
    ON nexus_forum_meme_comments FOR INSERT
    TO authenticated
    WITH CHECK (user_id = auth.uid());

CREATE POLICY "All authenticated users can view polls"
    ON nexus_forum_polls FOR SELECT
    TO authenticated
    USING (true);

CREATE POLICY "Users can create polls"
    ON nexus_forum_polls FOR INSERT
    TO authenticated
    WITH CHECK (user_id = auth.uid());

CREATE POLICY "Users can vote on active polls"
    ON nexus_forum_poll_votes FOR INSERT
    TO authenticated
    WITH CHECK (
        user_id = auth.uid()
        AND EXISTS (
            SELECT 1 FROM nexus_forum_polls
            WHERE nexus_forum_polls.id = nexus_forum_poll_votes.poll_id
            AND nexus_forum_polls.status = 'active'
        )
    );

-- ============================================
-- MODULE 5: NEXUS ARENA (Sports Matches)
-- ============================================

CREATE TABLE IF NOT EXISTS nexus_arena_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    sport_type TEXT NOT NULL CHECK (sport_type IN ('football', 'basketball', 'volleyball', 'tennis', 'other')),
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    match_date DATE NOT NULL,
    match_time TIME NOT NULL,
    location TEXT NOT NULL,
    max_players INTEGER NOT NULL CHECK (max_players >= 2 AND max_players <= 50),
    current_players INTEGER NOT NULL DEFAULT 1,
    skill_level TEXT DEFAULT 'all_levels' CHECK (skill_level IN ('beginner', 'intermediate', 'advanced', 'all_levels')),
    status TEXT NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'full', 'completed', 'cancelled')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(match_date)
);

CREATE TABLE IF NOT EXISTS nexus_arena_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID NOT NULL REFERENCES nexus_arena_matches(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'waitlist', 'cancelled')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(match_id, user_id)
);

CREATE INDEX idx_arena_matches_date ON nexus_arena_matches(match_date);
CREATE INDEX idx_arena_participants_match_id ON nexus_arena_participants(match_id);

ALTER TABLE nexus_arena_matches ENABLE ROW LEVEL SECURITY;
ALTER TABLE nexus_arena_participants ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view scheduled matches"
    ON nexus_arena_matches FOR SELECT
    TO authenticated
    USING (true);

CREATE POLICY "Users can create one match per day"
    ON nexus_arena_matches FOR INSERT
    TO authenticated
    WITH CHECK (
        user_id = auth.uid()
        AND NOT EXISTS (
            SELECT 1 FROM nexus_arena_matches
            WHERE nexus_arena_matches.match_date = nexus_arena_matches.match_date
            AND nexus_arena_matches.status IN ('scheduled', 'full')
        )
    );

CREATE POLICY "Users can join matches"
    ON nexus_arena_participants FOR INSERT
    TO authenticated
    WITH CHECK (
        user_id = auth.uid()
        AND EXISTS (
            SELECT 1 FROM nexus_arena_matches
            WHERE nexus_arena_matches.id = nexus_arena_participants.match_id
            AND nexus_arena_matches.status IN ('scheduled', 'full')
        )
    );

-- Function to enforce one match per day
CREATE OR REPLACE FUNCTION check_one_match_per_day()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM nexus_arena_matches
        WHERE match_date = NEW.match_date
        AND status IN ('scheduled', 'full')
        AND id != NEW.id
    ) THEN
        RAISE EXCEPTION 'Only one match allowed per day';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_one_match_per_day
    BEFORE INSERT OR UPDATE ON nexus_arena_matches
    FOR EACH ROW
    EXECUTE FUNCTION check_one_match_per_day();

-- Functions to update counts
CREATE OR REPLACE FUNCTION update_habitat_filled_spots()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE nexus_habitat_posts
    SET filled_spots = (
        SELECT COUNT(*)
        FROM nexus_habitat_requests
        WHERE post_id = NEW.post_id AND status = 'accepted'
    ),
    status = CASE
        WHEN (
            SELECT COUNT(*)
            FROM nexus_habitat_requests
            WHERE post_id = NEW.post_id AND status = 'accepted'
        ) >= (SELECT available_spots FROM nexus_habitat_posts WHERE id = NEW.post_id)
        THEN 'filled'
        ELSE 'open'
    END,
    updated_at = NOW()
    WHERE id = NEW.post_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER after_habitat_request_accepted
    AFTER UPDATE ON nexus_habitat_requests
    FOR EACH ROW
    WHEN (NEW.status = 'accepted' AND OLD.status != 'accepted')
    EXECUTE FUNCTION update_habitat_filled_spots();

CREATE OR REPLACE FUNCTION update_food_participants()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE nexus_pulse_food_posts
    SET current_participants = (
        SELECT COUNT(*) + 1
        FROM nexus_pulse_food_joins
        WHERE post_id = NEW.post_id
    ),
    status = CASE
        WHEN (
            SELECT COUNT(*) + 1
            FROM nexus_pulse_food_joins
            WHERE post_id = NEW.post_id
        ) >= 5 THEN 'locked'
        ELSE 'open'
    END
    WHERE id = NEW.post_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER after_food_join
    AFTER INSERT ON nexus_pulse_food_joins
    FOR EACH ROW
    EXECUTE FUNCTION update_food_participants();