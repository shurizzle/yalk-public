package pg

import (
	"chat/logger"
	"chat/shared"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
)

func MessageCreate(db *sql.DB, timestamp time.Time, message string, user_id string, chat_id string) (id string) {

	sqlStatement := fmt.Sprintf(`INSERT INTO chats."%v" (timestamp, message, event, user_id)
								VALUES ($1, $2, $3, $4)
								RETURNING id`, strings.ToUpper(chat_id))
	err := db.QueryRow(sqlStatement, timestamp, message, "chat_message", user_id).Scan(&id)
	if err != nil {
		panic(err)
	}
	logger.LogColor("DATABASE", fmt.Sprintf("Logged message %s in table '%s'", id, chat_id))
	return
}

func MessageReadAll(db *sql.DB, chat_id string) (messages map[string]shared.Chat_Message) {
	// sqlStatement := `SELECT * FROM all_messages_profiles_last_100;`
	sqlStatement := fmt.Sprintf(`SELECT * FROM (
		SELECT * FROM chats."%s"
		ORDER BY timestamp DESC 
		LIMIT 100
		) sub
	ORDER BY timestamp ASC;`, chat_id)

	rows, err := db.Query(sqlStatement)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	messages = make(map[string]shared.Chat_Message)
	for rows.Next() {
		var id string
		var timestamp time.Time
		var text string
		var from string
		var msg_type string

		err = rows.Scan(&id, &timestamp, &text, &msg_type, &from)
		if err != nil {
			panic(err)
		}
		messageStruct := shared.Chat_Message{
			ID:        id,
			Timestamp: timestamp,
			From:      from,
			To:        chat_id,
			Text:      text,
			Type:      msg_type,
		}
		messages[id] = messageStruct
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return messages
}

func ChatCreate(db *sql.DB, creator string, chat_type string, name string, users []string) (id string) {
	var scanned_id string
	users = append(users, creator)
	sort.Strings(users)
	if chat_type == "dm" {
		sqlStatement := `SELECT COALESCE(chats.exists($1), '')`
		db.QueryRow(sqlStatement, pq.StringArray(users)).Scan(&scanned_id)
		if scanned_id != "" {
			logger.LogColor("DATABASE", fmt.Sprintf("Chat %v already exists", scanned_id))
			return scanned_id
		}
	}

	sqlStatement := `SELECT chats.new($1, $2, $3, $4)`
	err := db.QueryRow(sqlStatement, name, chat_type, creator, pq.Array(users)).Scan(&id)
	if err != nil {
		logger.LogColor("DATABASE", "Error creating new chat")
		return
	}
	logger.LogColor("DATABASE", fmt.Sprintf("CHAT - NEW - [%v][%v] '%v' created with users ID joining %v", chat_type, id, name, users))
	return id
}

func ChatDelete(db *sql.DB, id string) (err error) {
	sqlStatement := `SELECT chats.delete($1)`

	err = db.QueryRow(sqlStatement, id).Err()
	if err != nil {
		logger.LogColor("DATABASE", "Error deleting chat")
		return
	}
	logger.LogColor("DATABASE", fmt.Sprintf("CHAT - DELETE - [%v]", id))
	return nil
}

func ChatInfo(db *sql.DB, chat_id string, wantMsg bool) (chat shared.Conversation, err error) {
	sqlStatement := `SELECT * FROM server_chats WHERE id = $1`
	var id string
	var chat_type string
	var users []string
	var name string
	var created_by string
	var creation_date time.Time

	row := db.QueryRow(sqlStatement, strings.ToUpper(chat_id))
	switch err := row.Scan(&id, &chat_type, pq.Array(&users), &name, &created_by, &creation_date); err {
	case sql.ErrNoRows:
		logger.LogColor("DATABASE", "CHAT - INFO not found!")
		return chat, err
	case nil:
		chat = shared.Conversation{
			ID:           id,
			Name:         name,
			Users:        users,
			Type:         chat_type,
			Creator:      created_by,
			CreationDate: creation_date,
		}
		if wantMsg {
			chat.Messages = MessageReadAll(db, id)
		}
		logger.LogColor("DATABASE", fmt.Sprintf("Fetched chat %v | wantMsg: %v", id, wantMsg))
		return chat, nil
	default:
		logger.LogColor("DATABASE", "Error in pg.ChatInfo")
		return chat, err
	}
}

func ChatReadAll(db *sql.DB, user_id string) (channels map[string]shared.Conversation) {
	// ! SANITIZE USER INPUT
	// pub_ch_suffix :=
	sqlStatement := `SELECT * FROM server_chats WHERE (users @> ARRAY[$1]::text[]) OR (server_chats.type LIKE ('%_public'));`
	// sqlStatement := fmt.Sprintf(`SELECT * FROM public.server_chats WHERE users @> ARRAY['%s']::text[] OR type LIKE '%_public';`, user_id)

	rows, err := db.Query(sqlStatement, user_id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	channels = make(map[string]shared.Conversation)
	for rows.Next() {
		var id string
		var chat_type string
		var users []string
		var name string
		var created_by string
		var creation_date any

		err = rows.Scan(&id, &chat_type, pq.Array(&users), &name, &created_by, &creation_date)
		if err != nil {
			panic(err)
		}
		channel := shared.Conversation{
			ID:       id,
			Name:     name,
			Users:    users,
			Type:     chat_type,
			Messages: MessageReadAll(db, strings.ToUpper(id)),
		}
		channels[id] = channel
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return
}

func ChatJoin(db *sql.DB, user_id string, chat_id string) (shared.Conversation, error) {
	var id string
	var chat shared.Conversation
	var r_status error
	// var chat_type string
	// var users []string
	// var name string
	// var created_by string
	// var creation_date time.Time

	sqlStatement := fmt.Sprintf(`SELECT * FROM chats.join('%v', '%v')`, strings.ToUpper(chat_id), user_id)
	row := db.QueryRow(sqlStatement)

	// switch err := row.Scan(&id, &chat_type, pq.Array(&users), &name, &created_by, &creation_date); err {
	switch err := row.Scan(&id); err {
	case sql.ErrNoRows:
		logger.LogColor("DATABASE", "CHAT - JOIN chat not found!")
		r_status = fmt.Errorf("err")
	case nil:
		switch id {
		case "exists":
			logger.LogColor("DATABASE", "CHAT - JOIN user is already part of chat")
			r_status = fmt.Errorf("err")
		case "err":
		case "":
			logger.LogColor("DATABASE", "CHAT - JOIN Error joining chat")
			r_status = fmt.Errorf("err")
		default:
			r_chat, err := ChatInfo(db, id, true)
			if err != nil {
				logger.LogColor("DATABASE", "CHAT - JOIN Error getting chat info")
				r_status = err
			}
			logger.LogColor("DATABASE", fmt.Sprintf("User %v joined channel %v", user_id, chat_id))
			chat = r_chat
		}
	default:
		logger.LogColor("DATABASE", "Error in pg.ChatJoin")
		r_status = fmt.Errorf("err")
	}
	return chat, r_status
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
	// 	logger.LogColor("DATABASE", "JOIN - ERROR UPDATE ServerChannels")
	// 	return chat, err
	// }
	// chat, err = ChatInfo(db, chat_id, true)
	// if err != nil {
	// 	logger.LogColor("DATABASE", "JOIN - ERROR GET ChatInfo")
	// 	return chat, err
	// }
}
