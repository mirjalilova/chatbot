CREATE TYPE role AS ENUM ('guest', 'user', 'admin', 'pro-user', 'business-user');

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  full_name VARCHAR(200),
  phone_number VARCHAR(13) UNIQUE NOT NULL,
  role role NOT NULL DEFAULT 'user',
  avatar TEXT,
  language TEXT NOT NULL DEFAULT 'en',
  ip_address VARCHAR(50),
  user_agent TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  deleted_at BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS restrictions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  type role NOT NULL,
  request_limit INT NOT NULL,
  time_limit INT
);

CREATE TABLE IF NOT EXISTS chat_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL DEFAULT 'New Chat',
    user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS chat (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_room_id UUID NOT NULL REFERENCES chat_rooms(id),
    user_request TEXT NOT NULL,
    gemini_request TEXT NOT NULL,
    responce TEXT NOT NULL,
    citation_urls TEXT[] DEFAULT '{}',
    location TEXT[] DEFAULT '{}',
    images_url TEXT[] DEFAULT '{}',
    organizations jsonb DEFAULT '[]'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at BIGINT NOT NULL DEFAULT 0
);




INSERT INTO restrictions (type, request_limit, time_limit) VALUES
('user', 10000, 100000),
('admin', 0, 0),
('pro-user', 0, 0),
('guest', 3, 0);

INSERT INTO users (id, full_name, phone_number, password) VALUES
('00000000-0000-0000-0000-000000000000', 'admin', '+998979004416', '$2a$10$hQIviMhEdjuyAAKYsImr6uw9739a96iQNfJhox/lx/7foJzsKR9JW'); -- password is 'adminpassword'


-- CREATE OR REPLACE FUNCTION update_all_chat_room_titles()
-- RETURNS TRIGGER AS $$
-- DECLARE
--     r RECORD;
--     new_title TEXT;
-- BEGIN
--     FOR r IN SELECT id FROM chat_rooms WHERE deleted_at = 0
--     LOOP
--         SELECT
--             string_agg(word, ' ') INTO new_title
--         FROM (
--             SELECT unnest(string_to_array(responce, ' ')) AS word
--             FROM chat
--             WHERE chat_room_id = r.id AND responce IS NOT NULL AND deleted_at = 0
--             ORDER BY created_at ASC
--             LIMIT 3
--         ) AS words;

--         IF new_title IS NULL THEN
--             new_title := 'New Chat';
--         END IF;

--         UPDATE chat_rooms
--         SET title = new_title,
--             updated_at = NOW()
--         WHERE id = r.id;
--     END LOOP;

--     RETURN NULL;
-- END;
-- $$ LANGUAGE plpgsql;



-- DROP TRIGGER IF EXISTS trg_update_all_chat_titles ON chat;

-- CREATE TRIGGER trg_update_all_chat_titles
-- AFTER INSERT ON chat
-- FOR EACH ROW
-- EXECUTE FUNCTION update_all_chat_room_titles();


CREATE OR REPLACE FUNCTION update_all_chat_room_titles()
RETURNS TRIGGER AS $$
DECLARE
    new_title TEXT;
BEGIN
    SELECT string_agg(word, ' ')
    INTO new_title
    FROM (
        SELECT unnest(string_to_array(c.user_request, ' ')) AS word
        FROM chat c
        WHERE c.chat_room_id = NEW.chat_room_id AND c.deleted_at = 0
        ORDER BY c.created_at ASC
        LIMIT 3
    ) AS words;

    IF new_title IS NULL OR length(trim(new_title)) = 0 THEN
        new_title := 'New Chat';
    END IF;

    UPDATE chat_rooms
    SET title = new_title,
        updated_at = NOW()
    WHERE id = NEW.chat_room_id;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_all_chat_titles
AFTER INSERT ON chat
FOR EACH ROW
EXECUTE FUNCTION update_all_chat_room_titles();




-- active users and count request
CREATE OR REPLACE FUNCTION get_daily_chat_stats(start_date DATE, end_date DATE)
RETURNS TABLE (
    day DATE,
    active_users INT,
    requests_count INT
) AS $$
BEGIN
    RETURN QUERY
    WITH days AS (
        SELECT generate_series(start_date, end_date, interval '1 day')::date AS day
    ),
    active AS (
        SELECT
            cr.updated_at::date AS day,
            COUNT(DISTINCT cr.user_id)::int AS active_users
        FROM chat_rooms cr
        WHERE cr.updated_at::date BETWEEN start_date AND end_date
        GROUP BY cr.updated_at::date
    ),
    requests AS (
        SELECT
            c.created_at::date AS day,
            COUNT(*)::int AS requests_count
        FROM chat c
        WHERE c.created_at::date BETWEEN start_date AND end_date
        GROUP BY c.created_at::date
    )
    SELECT
        d.day,
        COALESCE(a.active_users, 0)::int AS active_users,
        COALESCE(r.requests_count, 0)::int AS requests_count
    FROM days d
    LEFT JOIN active a ON a.day = d.day
    LEFT JOIN requests r ON r.day = d.day
    ORDER BY d.day DESC;
END;
$$ LANGUAGE plpgsql;
