CREATE OR ALTER PROCEDURE dbo.ResetUserPassword
    @LoginName NVARCHAR(255),      
    @NewPassword NVARCHAR(255),    
    @DisablePolicy BIT = 0,        
    @DisableExpiration BIT = 0     
AS
SET NOCOUNT ON;
BEGIN
    SET NOCOUNT ON;

    DECLARE @SQL NVARCHAR(MAX);

    SET @SQL = 'ALTER LOGIN [' + @LoginName + '] WITH PASSWORD = ''' + @NewPassword + '''';

    IF @DisablePolicy = 1
    BEGIN
        SET @SQL = @SQL + ' , CHECK_POLICY = OFF';
    END
    ELSE
    BEGIN
        SET @SQL = @SQL + ' , CHECK_POLICY = ON';
    END

    IF @DisableExpiration = 1
    BEGIN
        SET @SQL = @SQL + ' , CHECK_EXPIRATION = OFF';
    END
    ELSE
    BEGIN
        SET @SQL = @SQL + ' , CHECK_EXPIRATION = ON';
    END

    EXEC sp_executesql @SQL;
    
    PRINT 'Password reset for login [' + @LoginName + '] has been completed.';
END;
    