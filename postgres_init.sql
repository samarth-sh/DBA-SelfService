-- Procedure to create the pass_reset_logs table
DROP PROCEDURE IF EXISTS create_pass_reset_logs_table;
CREATE OR REPLACE PROCEDURE create_pass_reset_logs_table()
LANGUAGE plpgsql
AS $$
BEGIN
    DROP TABLE IF EXISTS pass_reset_logs;
    
    CREATE TABLE pass_reset_logs (
        id SERIAL PRIMARY KEY,
        username TEXT NOT NULL,
        serverIP TEXT NOT NULL,
        request_type TEXT NOT NULL DEFAULT 'Password Reset',
        request_status TEXT DEFAULT 'Pending',
        message TEXT,
        created_at TIMESTAMPTZ DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Kolkata') -- Adjusting to IST
    );
END;
$$;

-- Procedure to create the admin table
DROP PROCEDURE IF EXISTS create_admin_table;
CREATE OR REPLACE PROCEDURE create_admin_table()
LANGUAGE plpgsql
AS $$
BEGIN
    DROP TABLE IF EXISTS admin;
    
    CREATE TABLE admin (
        id SERIAL PRIMARY KEY,
        username TEXT NOT NULL UNIQUE,     
        password_hash TEXT NOT NULL,
        password_last_updated TIMESTAMPTZ DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Kolkata'),
        last_login TIMESTAMPTZ DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Kolkata'),
        created_at TIMESTAMPTZ DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Kolkata')
    );
END;
$$;

-- Procedure to create the pass_reset_logs table, used for logging password reset requests
DROP PROCEDURE IF EXISTS log_updates;
CREATE OR REPLACE PROCEDURE log_updates(
    IN uname TEXT, 
    IN ser_ip TEXT, 
    IN req_type TEXT, 
    IN req_status TEXT, 
    IN msg TEXT
)
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO pass_reset_logs (username, serverIP, request_type, request_status, message, created_at)
    VALUES (uname, ser_ip, req_type, req_status, msg, CURRENT_TIMESTAMP);
END;
$$;

-- used for updating password reset requests
DROP PROCEDURE IF EXISTS update_pass_reset_logs;
CREATE OR REPLACE PROCEDURE update_pass_reset_logs(
    IN uname TEXT, 
    IN ser_ip TEXT, 
    IN req_type TEXT, 
    IN req_status TEXT, 
    IN msg TEXT
)
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE pass_reset_logs SET 
        request_status = req_status,
        message = msg,
        created_at = CURRENT_TIMESTAMP
    WHERE username = uname AND serverIP = ser_ip AND request_type = req_type;
END;
$$;

-- Procedure to insert admin cerdentials into the admin table
CREATE EXTENSION IF NOT EXISTS pgcrypto; -- used for hashing and comparing passwords

DROP FUNCTION IF EXISTS insert_into_admin;
CREATE FUNCTION insert_into_admin(IN uname TEXT, IN pass TEXT)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO admin (username, password_hash)
    VALUES (uname, crypt(pass, gen_salt('bf')));
END;
$$;

-- Procedure to validate admin credentials given by the admin user
DROP FUNCTION IF EXISTS check_admin_credentials;
CREATE FUNCTION check_admin_credentials(ad_uname TEXT, ad_pass TEXT)
RETURNS BOOLEAN
LANGUAGE plpgsql
AS $$
DECLARE
    admin_exists BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM admin
        WHERE username = ad_uname
        AND password_hash = crypt(ad_pass, password_hash)
    ) INTO admin_exists;

    IF admin_exists THEN
        UPDATE admin SET last_login = CURRENT_TIMESTAMP WHERE username = ad_uname;
    END IF;

    RETURN admin_exists;
END;
$$;


-- Procedure to get all logs from the pass_reset_logs table in the form of a table
-- Procedure to get all logs from the pass_reset_logs table in the form of a table
DROP FUNCTION IF EXISTS get_all_logs;
CREATE OR REPLACE FUNCTION get_all_logs()
RETURNS TABLE (
    id INT,
    username TEXT,
    serverIP TEXT,
    request_type TEXT,
    request_status TEXT,
    created_at TIMESTAMPTZ
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY 
        SELECT 
            p.id, p.username, p.serverIP, p.request_type, p.request_status, 
            p.created_at
        FROM pass_reset_logs p;
END;
$$;