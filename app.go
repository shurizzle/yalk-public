package main

import (
	"chat/logger"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
)

type chat struct {
	ID           string             `json:"id"`
	Type         string             `json:"type"`
	Name         string             `json:"name"`
	Users        []string           `json:"users"`
	Messages     map[string]message `json:"messages"`
	Creator      string             `json:"creator"`
	CreationDate time.Time          `json:"creationDate"`
}

type message struct {
	ID   string    `json:"message_id"`
	Time time.Time `json:"time"`
	From string    `json:"from"`
	To   string    `json:"to"`
	Type string    `json:"type"`
	Text string    `json:"text"`
}

type profile struct {
	ID          string    `json:"user_id"`
	Username    string    `json:"username"`
	IsAdmin     string    `json:"isAdmin"`
	DisplayName string    `json:"display_name"`
	Color       string    `json:"color"`
	IsOnline    bool      `json:"isOnline"`
	Status      string    `json:"status"`
	StatusText  string    `json:"statusText"`
	JoinedChats []string  `json:"joined_chats"`
	LastLogin   time.Time `json:"lastLogin"`
	LastOffline time.Time `json:"lastOffline"`
}

func messageCreate(db *sql.DB, time time.Time, message string, from string, to string) (id string) {

	sqlStatement := fmt.Sprintf(`INSERT INTO chats."%v" (time, message, event, user_id)
								VALUES ($1, $2, $3, $4)
								RETURNING id`, strings.ToUpper(to))
	err := db.QueryRow(sqlStatement, time, message, "chat_message", from).Scan(&id)
	if err != nil {
		panic(err)
	}
	logger.LogColor("DATABASE", fmt.Sprintf("Logged message %s in table '%s'", id, to))
	return
}

func messageReadAll(db *sql.DB, id string) (messages map[string]message) {
	// sqlStatement := `SELECT * FROM all_messages_profiles_last_100;`
	sqlStatement := fmt.Sprintf(`SELECT * FROM (
		SELECT * FROM chats."%s"
		ORDER BY time DESC 
		LIMIT 100
		) sub
	ORDER BY time ASC;`, id)

	rows, err := db.Query(sqlStatement)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	messages = make(map[string]message)
	for rows.Next() {
		var id string
		var time time.Time
		var text string
		var from string
		var msgType string

		err = rows.Scan(&id, &time, &text, &msgType, &from)
		if err != nil {
			panic(err)
		}
		messageStruct := message{
			ID:   id,
			Time: time,
			From: from,
			To:   id,
			Text: text,
			Type: msgType,
		}
		messages[id] = messageStruct
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return messages
}

func chatCreate(db *sql.DB, creator string, chatType string, name string, users []string) string {
	var id string
	users = append(users, creator)
	sort.Strings(users)
	if chatType == "dm" {
		sqlStatement := `SELECT COALESCE(chats.exists($1), '')`
		err := db.QueryRow(sqlStatement, pq.StringArray(users)).Scan(&id)
		if err != nil {
			logger.LogColor("DATABASE", "Cannot create new chat")
			return id
		}
		if id != "" {
			logger.LogColor("DATABASE", fmt.Sprintf("Chat %v already exists", id))
			return id
		}
	}

	sqlStatement := `SELECT chats.new($1, $2, $3, $4)`
	err := db.QueryRow(sqlStatement, name, chatType, creator, pq.Array(users)).Scan(&id)
	if err != nil {
		logger.LogColor("DATABASE", "Error creating new chat")
		return id
	}
	logger.LogColor("DATABASE", fmt.Sprintf("CHAT - NEW - [%v][%v] '%v' created with users ID joining %v", chatType, id, name, users))
	return id
}

func chatDelete(db *sql.DB, id string) (err error) {
	sqlStatement := `SELECT chats.delete($1)`

	err = db.QueryRow(sqlStatement, id).Err()
	if err != nil {
		logger.LogColor("DATABASE", "Error deleting chat")
		return
	}
	logger.LogColor("DATABASE", fmt.Sprintf("CHAT - DELETE - [%v]", id))
	return nil
}

func chatInfo(db *sql.DB, _id string, wantMsg bool) (chatroom chat, err error) {
	sqlStatement := `SELECT * FROM server_chats WHERE id = $1`
	var id string
	var chatType string
	var users []string
	var name string
	var createBy string
	var creationDate time.Time

	row := db.QueryRow(sqlStatement, strings.ToUpper(_id))
	switch err := row.Scan(&id, &chatType, pq.Array(&users), &name, &createBy, &creationDate); err {
	case sql.ErrNoRows:
		logger.LogColor("DATABASE", "CHAT - INFO not found!")
		return chatroom, err
	case nil:
		chatroom = chat{
			ID:           id,
			Name:         name,
			Users:        users,
			Type:         chatType,
			Creator:      createBy,
			CreationDate: creationDate,
		}
		if wantMsg {
			chatroom.Messages = messageReadAll(db, id)
		}
		logger.LogColor("DATABASE", fmt.Sprintf("Fetched chat %v | wantMsg: %v", id, wantMsg))
		return chatroom, nil
	default:
		logger.LogColor("DATABASE", "Error in ChatInfo")
		return chatroom, err
	}
}

func chatReadall(db *sql.DB, id string) (channels map[string]chat) {
	// ! SANITIZE USER INPUT
	// pub_ch_suffix :=
	sqlStatement := `SELECT * FROM server_chats WHERE (users @> ARRAY[$1]::text[]) OR (server_chats.type LIKE ('%_public'));`
	// sqlStatement := fmt.Sprintf(`SELECT * FROM public.server_chats WHERE users @> ARRAY['%s']::text[] OR type LIKE '%_public';`, user_id)

	rows, err := db.Query(sqlStatement, id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	channels = make(map[string]chat)
	for rows.Next() {
		var id string
		var chatType string
		var users []string
		var name string
		var createdBy string
		var creationDate any

		err = rows.Scan(&id, &chatType, pq.Array(&users), &name, &createdBy, &creationDate)
		if err != nil {
			panic(err)
		}
		channel := chat{
			ID:       id,
			Name:     name,
			Users:    users,
			Type:     chatType,
			Messages: messageReadAll(db, strings.ToUpper(id)),
		}
		channels[id] = channel
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return
}

func chatJoin(db *sql.DB, userID string, chatID string) (chat, error) {
	var id string
	var _chat chat
	var _status error
	// var chatType string
	// var users []string
	// var name string
	// var createdBy string
	// var creationDate time.Time

	sqlStatement := fmt.Sprintf(`SELECT * FROM chats.join('%v', '%v')`, strings.ToUpper(chatID), userID)
	row := db.QueryRow(sqlStatement)

	// switch err := row.Scan(&id, &chatType, pq.Array(&users), &name, &createdBy, &creationDate); err {
	switch err := row.Scan(&id); err {
	case sql.ErrNoRows:
		logger.LogColor("DATABASE", "CHAT - JOIN chat not found!")
		_status = fmt.Errorf("err")
	case nil:
		switch id {
		case "exists":
			logger.LogColor("DATABASE", "CHAT - JOIN user is already part of chat")
			_status = fmt.Errorf("err")
		case "err":
		case "":
			logger.LogColor("DATABASE", "CHAT - JOIN Error joining chat")
			_status = fmt.Errorf("err")
		default:
			dbChat, err := chatInfo(db, id, true)
			if err != nil {
				logger.LogColor("DATABASE", "CHAT - JOIN Error getting chat info")
				_status = err
			}
			logger.LogColor("DATABASE", fmt.Sprintf("User %v joined channel %v", userID, chatID))
			_chat = dbChat
		}
	default:
		logger.LogColor("DATABASE", "Error in ChatJoin")
		err = fmt.Errorf("err")
	}
	return _chat, _status
	// sqlUserJoin := `
	// 	UPDATE users_settings
	// 	SET joined_chats = array_append(joined_chats, $1)
	// 	WHERE id = $2
	// `
	// _, err = db.Exec(sqlUserJoin, chat_id, user_id)
	// if err != nil {
	// 	logger.LogColor("DATABASE", "JOIN - ERROR UPDATE UserProfile")
	// 	return chat, err
	// }

	// sqlChanJoin := `
	// 	UPDATE server_chats
	// 	SET users = array_append(users, $1)
	// 	WHERE id = $2
	// `
	// _, err = db.Exec(sqlChanJoin, user_id, chat_id)
	// if err != nil {
	// 	logger.LogColor("DATABASE", "JOIN - ERROR UPDATE syncChannels")
	// 	return chat, err
	// }
	// chat, err = ChatInfo(db, chat_id, true)
	// if err != nil {
	// 	logger.LogColor("DATABASE", "JOIN - ERROR GET ChatInfo")
	// 	return chat, err
	// }
}

func serverSettingsRead(db *sql.DB) configServer {
	sqlStatement := `SELECT * FROM server_settings;`

	rows, err := db.Query(sqlStatement)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	settings := make(map[string]string)

	for rows.Next() {
		var key string
		var value string

		err = rows.Scan(&key, &value)
		if err != nil {
			panic(err)
		}
		settings[key] = value
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return configServer{
		ServerID:       settings["server_id"],
		DefaultChannel: settings["default_channel"],
		TestKey:        settings["test_key"],
		ConnType:       settings["conn_type"],
	}
}

func userCreate(db *sql.DB, username string, pwdHash string) (id string) {
	sqlStatement := `INSERT INTO login_users (username, password)
	VALUES ($1, $2)
	RETURNING id`
	err := db.QueryRow(sqlStatement, username, pwdHash).Scan(&id)
	if err != nil {
		logger.LogColor("DATABASE", "Error creating new user")
		return
	}
	logger.LogColor("DATABASE", fmt.Sprintf("Added user %s with id %v", username, id))
	return id
}

func profileCreate(db *sql.DB, id string, displayName string, color string, isAdmin string) {
	sqlStatement := `INSERT INTO users_settings (id, display_name, color, isAdmin)
	VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(sqlStatement, id, displayName, color, isAdmin)
	if err != nil {
		logger.LogColor("DATABASE", "Error CREATING new profile")
		return
	}
	logger.LogColor("DATABASE", fmt.Sprintf("Create profile with id %v", id))
}

func userRead(db *sql.DB, id string, self bool) (userProfile profile, err error) {
	var username string
	var isAdmin string
	var displayName string
	var color string
	var joinedChat []string
	var status string
	var statusText string
	var lastLogin time.Time
	var lastOffline time.Time
	var isOnline bool

	sqlStatement := `SELECT * FROM public.all_user_profiles WHERE id = $1`

	var row *sql.Row

	if id == "" {
		return userProfile, err
	}
	row = db.QueryRow(sqlStatement, id)
	// Here means: it assigns err with the row.Scan()
	// then "; err" means use "err" in the "switch" statement
	switch err := row.Scan(&id, &username, &displayName, &color, &isAdmin, pq.Array(&joinedChat), &status, &statusText, &lastLogin, &lastOffline, &isOnline); err {
	case sql.ErrNoRows:
		logger.LogColor("DATABASE", "No USER found!")
		return userProfile, err
	case nil:
		userProfile := profile{
			ID:          id,
			Username:    username,
			IsAdmin:     isAdmin,
			DisplayName: displayName,
			Color:       color,
			Status:      status,
			StatusText:  statusText,
			LastOffline: lastOffline,
			IsOnline:    isOnline,
		}
		if self {
			userProfile.JoinedChats = joinedChat
			userProfile.LastLogin = lastLogin
		}
		return userProfile, nil
	default:
		logger.LogColor("DATABASE", "Error in UserRead")
		return userProfile, err
	}
}

func profileUpdate(db *sql.DB, id string, displayName string, color string, isAdmin string) {
	sqlStatement := "UPDATE users_settings SET display_name = $2, color = $3"

	if isAdmin != "" {
		sqlStatement += (", " + "isAdmin = " + isAdmin)
	}

	sqlStatement += " WHERE id = $1;"
	_, err := db.Exec(sqlStatement, id, displayName, color)
	if err != nil {
		logger.LogColor("DATABASE", "Error UPDATING profile")
	}
	logger.LogColor("DATABASE", fmt.Sprintf("Updated id %v", id))
}

func userReadAll(db *sql.DB) (allUser map[string]profile) {
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
	allUser = make(map[string]profile)
	for rows.Next() {
		var id string
		var username string
		var isAdmin string
		var displayName string
		var color string
		var joinedChats []string
		var status string
		var statusText string
		var lastLogin time.Time
		var lastOffline time.Time
		var isOnline bool

		err = rows.Scan(&id, &username, &displayName, &color, &isAdmin, pq.Array(&joinedChats), &status, &statusText, &lastLogin, &lastOffline, &isOnline)
		if err != nil {
			panic(err)
		}
		userProfile := profile{
			ID:          id,
			Username:    username,
			DisplayName: displayName,
			Color:       color,
			IsAdmin:     isAdmin,
			Status:      status,
			StatusText:  statusText,
			LastOffline: lastOffline,
			IsOnline:    isOnline,
		}

		allUser[id] = userProfile
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return
}

// func UserStatusRead(db *sql.DB, user_id string) (bool, string, error) {
// 	sqlStatement := `SELECT * FROM user_status WHERE id=$1;`
// 	var IsOnline bool
// 	var status string      // * UNUSED -- For future functions
// 	var statusText string  // * UNUSED -- For future functions
// 	var lastLogin string   // * UNUSED -- For future functions
// 	var lastOffline string // * UNUSED -- For future functions

// 	row := db.QueryRow(sqlStatement, user_id)
// 	switch err := row.Scan(&user_id, &status, &status_fixed, &lastLogin, &lastOffline); err {
// 	case sql.ErrNoRows:
// 		logger.LogColor("DATABASE", "No USER IN STATUS were returned")
// 		return "", err
// 	case nil:
// 		return status, nil

// 	default:
// 		logger.LogColor("DATABASE", "Error in postgre.UserReadStatus")
// 		return "", err
// 	}
// }

func userStatusUpdate(db *sql.DB, id string, status string, statusFixed bool) error {
	sqlStatement := `UPDATE user_status 
					SET status = $2
					WHERE id = $1`
	_, err := db.Exec(sqlStatement, id, status)
	if err != nil {
		logger.LogColor("DATABASE", "Error UPDATING status")
		return err
	}
	logger.LogColor("DATABASE", fmt.Sprintf("Updated status for id %v", id))
	return nil
}

func userOnlineUpdate(db *sql.DB, id string, isOnline bool) error {
	sqlStatement := `UPDATE user_status 
					SET is_online = $2
					WHERE id = $1`
	_, err := db.Exec(sqlStatement, id, isOnline)
	if err != nil {
		logger.LogColor("DATABASE", "Error UPDATING Online status")
		return err
	}
	logger.LogColor("DATABASE", fmt.Sprintf("Updated online status for id %v [online: %v]", id, isOnline))
	return nil
}

func userDelete(db *sql.DB, id string) {
	sqlStatement := `DELETE FROM login_users
	WHERE id = $1;`
	_, err := db.Exec(sqlStatement, id)
	if err != nil {
		logger.LogColor("DATABASE", "Error DELETING user")
	}
	logger.LogColor("DATABASE", fmt.Sprintf("Deleted id %v", id))

}

func profileDelete(db *sql.DB, id string) {
	sqlStatement := `DELETE FROM users_settings
	WHERE id = $1;`
	_, err := db.Exec(sqlStatement, id)
	if err != nil {
		logger.LogColor("DATABASE", "Error DELETING profile")

	}
	logger.LogColor("DATABASE", fmt.Sprintf("Deleted id %v", id))
}
