package pg

import (
	"chat/shared"
	"database/sql"
)

func ServerSettingsRead(db *sql.DB) (serverSettings shared.Server_Settings) {
	sqlStatement := `SELECT * FROM server_settings;`

	rows, err := db.Query(sqlStatement)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var settings = make(map[string]string)

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

	serverSettings = shared.Server_Settings{
		ServerID:       settings["server_id"],
		DefaultChannel: settings["default_channel"],
		TestKey:        settings["test_key"],
		ConnType:       settings["conn_type"],
	}
	return
}
