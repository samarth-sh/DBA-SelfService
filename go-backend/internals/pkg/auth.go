package pkg

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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
	log.Println("User credentials checked successfully")

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
		log.Printf("Failed to connect to the server using old credentials: %v", err)
		return false, err
	}
	defer msWithUserCred.Close()

	if err = msWithUserCred.Ping(); err != nil {
		log.Printf("Old password failed authentication: %v", err)
		return false, err
	}

	log.Println("Old password is still valid")
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
			return false, fmt.Errorf("login does not exist")
		}
		return false, err
	}

	if !isValid.Valid || !isExpired.Valid {
		return false, fmt.Errorf("failed to retrieve login expiration status")
	}

	if !isValid.Bool || isExpired.Bool {
		return false, fmt.Errorf("login expired or invalid")
	}

	return true, nil
}