package shared

import (
	"errors"
	"sort"
	"strconv"
	"time"
)

// var CHANNEL_ID_PREFIXES = []string{"NP", "BQ", "PL", "UH"}
// var DM_ID_PREFIXES = []string{"PC", "QH", "AV", "NT"}

type HTTP_Session struct {
	Token   string
	UserID  string
	IP      string
	Created time.Time
	Expiry  time.Time
}

func (s HTTP_Session) Is_Expired() bool {
	return s.Expiry.Before(time.Now())
}

// * Generic User and Pass with ID * //
type Credentials struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Server_Channels struct {
	Msg     chan Data_Payload
	Dm      chan Data_Payload
	Notify  chan Data_Payload
	Cmd     chan Data_Payload
	Conn    chan Data_Payload
	Disconn chan Data_Payload
}

type Conversation struct {
	ID           string                  `json:"id"`
	Type         string                  `json:"type"`
	Name         string                  `json:"name"`
	Users        []string                `json:"users"`
	Messages     map[string]Chat_Message `json:"messages"`
	Creator      string                  `json:"creator"`
	CreationDate time.Time               `json:"creation_date"`
}

type Chat_Message struct {
	ID        string    `json:"message_id"`
	Timestamp time.Time `json:"timestamp"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Type      string    `json:"type"`
	Text      string    `json:"text"`
}

type Server_Settings struct {
	ServerID       string `json:"server_id"`
	DefaultChannel string `json:"default_channel"`
	TestKey        string `json:"test_key"`
	ConnType       string `json:"conn_type"`
}

type User_Profile struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	IsAdmin     string    `json:"is_admin"`
	DisplayName string    `json:"display_name"`
	Color       string    `json:"color"`
	Status      string    `json:"status"`
	StatusFixed string    `json:"status_fixed"`
	JoinedChats []string  `json:"joined_chats"`
	LastLogin   time.Time `json:"last_login"`
	LastOffline time.Time `json:"last_offline"`
}

type Data_Payload struct {
	Success      bool     `json:"success"`
	Origin       string   `json:"origin"`
	Destinations []string `json:"destinations"`
	Event        string   `json:"event"`
	Data         any      `json:"data"`
}

func NewPayload(succ bool, orig string, dest []string, context string, context_type string, event string, data any) (payload Data_Payload) {
	payload = Data_Payload{
		Success:      succ,
		Origin:       orig,
		Destinations: dest,
		Event:        event,
		Data:         data,
	}
	return payload
}

func Min(values []int) (min int, e error) {
	if len(values) == 0 {
		return 0, errors.New("cannot detect a minimum value in an empty slice")
	}

	min = values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}

	return min, nil
}

func Atoi(s string) (converted_value int, err error) {
	if len(s) > 0 {
		converted_value, err = strconv.Atoi(s)
		if err != nil {
			return converted_value, err
		}
	}
	return converted_value, err
}

func Higher(num1 int, num2 int) int {
	nums := []int{num1, num2}
	sort.Ints(nums)
	// for _, _ := range nums {
	// 	// 		// _ids = append(_ids, strconv.Itoa(id))
	// 	// 	}
	// }
	return 0
}
