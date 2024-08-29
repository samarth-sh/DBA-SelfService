CREATE OR REPLACE PROCEDURE create_users_table()
LANGUAGE plpgsql
AS $$
BEGIN
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username TEXT NOT NULL,
        password TEXT NOT NULL,
        serverIP TEXT NOT NULL,
        read_permission INTEGER DEFAULT 0,
        write_permission INTEGER DEFAULT 0,
        admin_permission INTEGER DEFAULT 0
    );
END;
$$;

CREATE OR REPLACE PROCEDURE insert_into_users(IN uname TEXT, IN pass TEXT, IN sIP TEXT)
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO users (username, password, serverIP) VALUES (uname, pass, sIP);
END;
$$;

CREATE OR REPLACE PROCEDURE create_logs_table()
LANGUAGE plpgsql
AS $$
BEGIN
    CREATE TABLE IF NOT EXISTS logs (
        request_id SERIAL PRIMARY KEY,
        username TEXT NOT NULL,
        serverIP TEXT NOT NULL,
        request_type TEXT DEFAULT 'Password Update',
        request_status TEXT DEFAULT 'Pending',
        message TEXT,
        request_time TIMESTAMPTZ DEFAULT NOW()
    );
END;
$$;

CREATE OR REPLACE PROCEDURE create_admin_table()
LANGUAGE plpgsql
AS $$
BEGIN
    CREATE TABLE IF NOT EXISTS admin (
        id SERIAL PRIMARY KEY,
        username TEXT NOT NULL,
        password TEXT NOT NULL
    );
END;
$$;


CREATE OR REPLACE PROCEDURE insert_into_admin(IN uname TEXT, IN pass TEXT)
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO admin (username, password) VALUES (uname, pass);
END;
$$;


CREATE OR REPLACE FUNCTION user_exists(uname TEXT)
RETURNS BOOLEAN
LANGUAGE plpgsql
AS $$
DECLARE
    user_exists BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM users
        WHERE username = uname
    ) INTO user_exists;

    RETURN user_exists;
END;
$$;

CREATE OR REPLACE FUNCTION get_user_password(uname TEXT)
RETURNS TEXT
LANGUAGE plpgsql
AS $$
DECLARE
    user_password TEXT;
BEGIN
    -- Assigning the result of the SELECT to a variable
    SELECT u.password INTO user_password
    FROM users u
    WHERE u.username = uname;

    -- If you want to return the password, use RAISE NOTICE or return the value
    RETURN user_password;

END;
$$;


CREATE OR REPLACE FUNCTION get_serverip(uname TEXT)
RETURNS TEXT
LANGUAGE plpgsql
AS $$
DECLARE
    server_ip TEXT;
BEGIN
    SELECT u.serverIP INTO server_ip
    FROM users u
    WHERE u.username = uname;

    RETURN server_ip;
END;
$$;


CREATE OR REPLACE PROCEDURE update_user_password(IN hashedPassword TEXT, IN uname TEXT, IN sIP TEXT)
LANGUAGE plpgsql
AS $$
BEGIN 
    UPDATE users
    SET password = hashedPassword
    WHERE username = uname
    AND serverIP = sIP;
END;
$$;


CREATE OR REPLACE PROCEDURE log_updates(
    IN req_type TEXT, 
    IN uname TEXT, 
    IN sIP TEXT, 
    IN req_status TEXT, 
    IN msg TEXT
)
LANGUAGE plpgsql
AS $$
BEGIN 
    INSERT INTO logs (request_type, username, serverIP, request_status, message)
    VALUES (req_type, uname, sIP, req_status, msg);
END;
$$;


CREATE OR REPLACE FUNCTION check_admin_credentials(ad_uname TEXT, ad_pass TEXT)
RETURNS BOOLEAN
LANGUAGE plpgsql
AS $$
DECLARE
    admin_exists BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM admin
        WHERE username = ad_uname
        AND password = ad_pass
    ) INTO admin_exists;

    RETURN admin_exists;
END;
$$;


CREATE OR REPLACE FUNCTION get_all_logs()
RETURNS TABLE (
    requestID INTEGER,
    username TEXT,
    serverIP TEXT,
    requestType TEXT,
    requestStatus TEXT,
    message TEXT,
    requestTime TIMESTAMPTZ
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT 
        request_id AS requestID, 
        username AS username, 
        serverip AS serverIP, 
        request_type AS requestType, 
        request_status AS requestStatus, 
        message AS message,
        request_time AS requestTime
    FROM logs;
END;
$$;









