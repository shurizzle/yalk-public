package pg

import (
	"chat/logger"
	"chat/shared"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

func UserCreate(db *sql.DB, username string, hashed_password string) (id string) {
	sqlStatement := `INSERT INTO login_users (username, password)
	VALUES ($1, $2)
	RETURNING id`
	err := db.QueryRow(sqlStatement, username, hashed_password).Scan(&id)
	if err != nil {
		logger.LogColor("DATABASE", "Error creating new user")
		return
	}
	logger.LogColor("DATABASE", fmt.Sprintf("Added user %s with id %v", username, id))
	return id
}

func ProfileCreate(db *sql.DB, id string, display_name string, color string, is_admin string) {
	sqlStatement := `INSERT INTO users_settings (id, display_name, color, is_admin)
	VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(sqlStatement, id, display_name, color, is_admin)
	if err != nil {
		logger.LogColor("DATABASE", "Error CREATING new profile")
		return
	}
	logger.LogColor("DATABASE", fmt.Sprintf("Create profile with id %v", id))
}

func UserRead(db *sql.DB, user_id string, self bool) (user_profile shared.User_Profile, err error) {
	var username string
	var is_admin string
	var display_name string
	var color string
	var joined_chats []string
	var status string
	var status_fixed string
	var last_login time.Time
	var last_offline time.Time

	sqlStatement := `SELECT * FROM public.all_user_profiles WHERE id = $1`

	var row *sql.Row

	if user_id == "" {
		return user_profile, err
	}
	row = db.QueryRow(sqlStatement, user_id)
	// Here means: it assigns err with the row.Scan()
	// then "; err" means use "err" in the "switch" statement
	switch err := row.Scan(&user_id, &username, &display_name, &color, &is_admin, pq.Array(&joined_chats), &status, &status_fixed, &last_login, &last_offline); err {
	case sql.ErrNoRows:
		logger.LogColor("DATABASE", "No USER found!")
		return user_profile, err
	case nil:
		user_profile := shared.User_Profile{
			ID:          user_id,
			Username:    username,
			IsAdmin:     is_admin,
			DisplayName: display_name,
			Color:       color,
			Status:      status,
			StatusFixed: status_fixed,
			LastOffline: last_offline,
		}
		if self {
			user_profile.JoinedChats = joined_chats
			user_profile.LastLogin = last_login
		}
		return user_profile, nil
	default:
		logger.LogColor("DATABASE", "Error in pg.UserRead")
		return user_profile, err
	}
}

func ProfileUpdate(db *sql.DB, user_id string, display_name string, color string, is_admin string) {
	sqlStatement := "UPDATE users_settings SET display_name = $2, color = $3"

	if is_admin != "" {
		sqlStatement += (", " + "is_admin = " + is_admin)
	}

	sqlStatement += " WHERE id = $1;"
	_, err := db.Exec(sqlStatement, user_id, display_name, color)
	if err != nil {
		logger.LogColor("DATABASE", "Error UPDATING profile")
	}
	logger.LogColor("DATABASE", fmt.Sprintf("Updated id %v", user_id))
}

func UserReadAll(db *sql.DB) (all_users map[string]shared.User_Profile) {
	// sqlStatement := `SELECT sub.*, users_settings.display_name, users_settings.color FROM
	// 				(
	// 					SELECT * FROM chat_channel_1 ORDER BY timestamp DESC LIMIT 100
	// 				) AS sub
	// 				INNER JOIN users_settings
	// 					ON sub.user_id = users_settings.id
	// 					ORDER BY timestamp;`
	sqlStatement := `SELECT * FROM all_user_profiles;`

	rows, err := db.Query(sqlStatement)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	all_users = make(map[string]shared.User_Profile)
	for rows.Next() {
		var user_id string
		var username string
		var display_name string
		var color string
		var is_admin string
		var joined_chats []string
		var status string
		var status_fixed string
		var last_login time.Time
		var last_offline time.Time

		err = rows.Scan(&user_id, &username, &display_name, &color, &is_admin, pq.Array(&joined_chats), &status, &status_fixed, &last_login, &last_offline)
		if err != nil {
			panic(err)
		}
		user_profile := shared.User_Profile{
			ID:          user_id,
			Username:    username,
			DisplayName: display_name,
			Color:       color,
			IsAdmin:     is_admin,
			// JoinedChats: joined_chats,
			Status:      status,
			StatusFixed: status_fixed,
			// LastLogin:   last_login,
			LastOffline: last_offline,
		}

		all_users[user_id] = user_profile
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return
}

func UserStatusRead(db *sql.DB, user_id string) (status string, err error) {
	sqlStatement := `SELECT * FROM user_status WHERE id=$1;`
	var status_fixed string // * UNUSED -- For future functions
	var last_login string   // * UNUSED -- For future functions
	var last_offline string // * UNUSED -- For future functions

	row := db.QueryRow(sqlStatement, user_id)
	switch err := row.Scan(&user_id, &status, &status_fixed, &last_login, &last_offline); err {
	case sql.ErrNoRows:
		logger.LogColor("DATABASE", "No USER IN STATUS were returned")
		return "", err
	case nil:
		return status, nil

	default:
		logger.LogColor("DATABASE", "Error in postgre.UserReadStatus")
		return "", err
	}
}

func UserStatusUpdate(db *sql.DB, user_id string, status string, status_fixed bool) error {
	sqlStatement := `UPDATE user_status 
					SET status = $2, status_fixed  = $3
					WHERE id = $1`
	_, err := db.Exec(sqlStatement, user_id, status, status_fixed)
	if err != nil {
		logger.LogColor("DATABASE", "Error UPDATING status")
		return err
	}
	logger.LogColor("DATABASE", fmt.Sprintf("Updated status for id %v", user_id))
	return nil
}

func UserDelete(db *sql.DB, user_id string) {
	sqlStatement := `DELETE FROM login_users
	WHERE id = $1;`
	_, err := db.Exec(sqlStatement, user_id)
	if err != nil {
		logger.LogColor("DATABASE", "Error DELETING user")

	}
	logger.LogColor("DATABASE", fmt.Sprintf("Deleted id %v", user_id))

}

func ProfileDelete(db *sql.DB, user_id string) {
	sqlStatement := `DELETE FROM users_settings
	WHERE id = $1;`
	_, err := db.Exec(sqlStatement, user_id)
	if err != nil {
		logger.LogColor("DATABASE", "Error DELETING profile")

	}
	logger.LogColor("DATABASE", fmt.Sprintf("Deleted id %v", user_id))
}
