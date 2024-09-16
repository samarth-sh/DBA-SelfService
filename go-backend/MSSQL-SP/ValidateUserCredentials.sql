CREATE OR ALTER PROCEDURE dbo.ValidateUserCredentials
    @Username NVARCHAR(255),
    @ServerIP NVARCHAR(255),
    @Email NVARCHAR(255),
    @IsValid BIT OUTPUT
AS
BEGIN
    SET @IsValid = 0;

    IF EXISTS (
        SELECT 1
        FROM dbo.login_email_mapping
        WHERE login_name = @Username
        AND (
              sql_instance_ip = @ServerIP OR
              sql_instance_ip = '*'
          )
        AND CHARINDEX(@Email, owner_group_email) > 0
    )
    BEGIN
        SET @IsValid = 1;
    END
END;
    
