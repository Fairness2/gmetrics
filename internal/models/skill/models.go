package skill

const (
	TypeSimpleUtterance = "SimpleUtterance"
)

// Request описывает запрос пользователя.
// см. https://yandex.ru/dev/dialogs/alice/doc/request.html
type Request struct {
	Request SimpleUtterance `json:"request"`
	Version string          `json:"version"`
	// тут будет, например, строка "Europe/Moscow" для часового пояса Москвы
	Timezone string  `json:"timezone"`
	Session  Session `json:"session"`
}

type Session struct {
	// New сообщает, является ли сессия новой
	New bool `json:"new"`
	// User пользователь запроса
	User User `json:"user"`
}

type User struct {
	// UserId идентификатор пользователя
	UserID string `json:"user_id"`
	// UserId токен доступа пользователя к апи
	AccessToken string `json:"access_token"`
}

// SimpleUtterance описывает команду, полученную в запросе типа SimpleUtterance.
type SimpleUtterance struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// Response описывает ответ сервера.
// см. https://yandex.ru/dev/dialogs/alice/doc/response.html
type Response struct {
	Response ResponsePayload `json:"response"`
	Version  string          `json:"version"`
}

// ResponsePayload описывает ответ, который нужно озвучить.
type ResponsePayload struct {
	Text string `json:"text"`
}
