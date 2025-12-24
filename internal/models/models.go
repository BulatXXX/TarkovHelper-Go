package models

type Mode string

const (
	ModePVP Mode = "pvp"
	ModePVE Mode = "pve"
)

func (m Mode) Valid() bool {
	return m == ModePVP || m == ModePVE
}

type User struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatarUrl"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type MeResponse struct {
	User User `json:"user"`
}

type TrackedItem struct {
	ID        string  `json:"id"`
	IconLink  *string `json:"iconLink"`
	UpdatedAt int64   `json:"updatedAt"`
}

type GetTrackedResponse struct {
	Items []TrackedItem `json:"items"`
}

type PutTrackedRequest struct {
	Items []TrackedItem `json:"items"`
}

type PutTrackedResponse struct {
	Items []TrackedItem `json:"items"`
}

type ErrorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
