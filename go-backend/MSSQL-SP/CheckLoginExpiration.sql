CREATE OR ALTER PROCEDURE dbo.CheckLoginExpiration
    @LoginName NVARCHAR(255),
    @SqlInstance NVARCHAR(255),
    @IsExpired BIT OUTPUT,
    @IsValid BIT OUTPUT
AS
BEGIN
    SET @IsExpired = 0;
    SET @IsValid = 0;

    SELECT 
        @IsExpired = lei.is_expired, 
        @IsValid = CASE 
            WHEN lei.is_expired = 1 THEN 0 
            ELSE 1 
        END
    FROM dbo.server_login_expiry_collection_computed c
    JOIN dbo.all_server_login_expiry_info lei
        ON lei.sql_instance = c.sql_instance 
        AND lei.collection_time = c.collection_time_latest
    WHERE 
        lei.login_name = @LoginName 
        AND c.sql_instance = @SqlInstance;

    IF @IsValid = 0 OR @IsExpired = 1
    BEGIN
        RAISERROR(18456, 16, 1);
        RETURN;
    END
END;

