package pkg

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"

)
func Check_user_credentials(msdb *sql.DB, username, serverIP, emailID string) (bool, error) {
	var isValidUser bool
	query := `DECLARE @IsValid BIT;
              EXEC dbo.ValidateUserCredentials @Username = ?, @ServerIP = ?, @Email = ?, @IsValid = @IsValid OUTPUT;
              SELECT @IsValid;`

	row := msdb.QueryRow(query, username, serverIP, emailID)
	if err := row.Scan(&isValidUser); err != nil {
		return false, err
	}
	log.Info().Msg("User credentials for" + username + "validated")

	return isValidUser, nil
}
func CheckOldPassword(msdb *sql.DB, username, serverIP, oldPassword, database string) (bool, error) {
	msWithUserCredstr := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s",
		serverIP,
		username,
		oldPassword,
		os.Getenv("MS_DB_PORT"),
		database)

	msWithUserCred, err := sql.Open("mssql", msWithUserCredstr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to MS SQL Server with user credentials")
		return false, err
	}
	defer msWithUserCred.Close()

	if err = msWithUserCred.Ping(); err != nil {
		log.Error().Err(err).Msgf("Failed to ping MS SQL Server: %v", err)
		return false, err
	}

	log.Info().Msg("Old Password is valid")
	return true, nil
}

func FindRelatedServers(msdb *sql.DB, serverIP string) ([]string, error) {
	query := "EXEC FindRelatedServers @ServerIP=?"
	rows, err := msdb.Query(query, serverIP)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var serverReplicas []string
	for rows.Next() {
		var serverIP string
		if err := rows.Scan(&serverIP); err != nil {
			return nil, err
		}
		serverReplicas = append(serverReplicas, serverIP)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return serverReplicas, nil
}
func CheckLoginExpiration(msdb *sql.DB, loginName, sqlInstance string) (bool, error) {
	var isExpired sql.NullBool
	var isValid sql.NullBool

	query := `
		DECLARE @IsExpired BIT, @IsValid BIT;
		EXEC dbo.CheckLoginExpiration @LoginName = ?, @SqlInstance = ?, @IsExpired = @IsExpired OUTPUT, @IsValid = @IsValid OUTPUT;
		SELECT @IsExpired, @IsValid;
	`
	err := msdb.QueryRow(query, loginName, sqlInstance).Scan(&isExpired, &isValid)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Error().Err(err).Msg("Login does not exist")
			return false, err
		}
		return false, err
	}

	if !isValid.Valid || !isExpired.Valid {
		log.Error().Msg("Failed to retrieve login expiration status")
		return false, err
	}

	if !isValid.Bool || isExpired.Bool {
		log.Error().Msg("Login expired or invalid")
		return false, err
	}
	log.Info().Msg("Login is still valid")
	return true, nil
}