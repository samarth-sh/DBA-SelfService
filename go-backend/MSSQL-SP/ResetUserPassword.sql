CREATE OR ALTER PROCEDURE dbo.ResetUserPassword
    @LoginName NVARCHAR(255),      
    @NewPassword NVARCHAR(255),    
    @OldPassword NVARCHAR(255)
AS
BEGIN
    SET NOCOUNT ON;

    DECLARE @SQL NVARCHAR(MAX);

    SET @SQL = 'ALTER LOGIN [' + @LoginName + '] WITH PASSWORD = ''' + @NewPassword + ''' OLD_PASSWORD = ''' + @OldPassword + ''';';

    BEGIN TRY
        EXEC sp_executesql @SQL;
    END TRY
    BEGIN CATCH
        PRINT 'Password reset for login [' + @LoginName + '] has failed.';
        THROW;
    END CATCH;

    EXEC sp_executesql @SQL;

    PRINT 'Password reset for login [' + @LoginName + '] has been completed.';
    
END;
