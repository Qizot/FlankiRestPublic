package errors

type ApiError struct {
	Message string
	HttpCode int
}

func New(message string, code int) error {
	return &ApiError{message, code}
}

func (error *ApiError) Error() string {
	return error.Message
}


var (
	BadJsonRequestFormat         = &ApiError{"Invalid request, check your json", 400}
	UnauthorizedAccount          = &ApiError{"Invalid credentials or user doesn't exist", 401}
	UnauthorizedLobbyJoinRequest = &ApiError{"Invalid lobby password", 401}
	LobbyIsFull                  = &ApiError{"Lobby is already full", 403}
	PlayerNotFoundInAnyTeam      = &ApiError{"Player was not a member of any team", 404}
	PlayerNotActive              = &ApiError{"Player was not present in any active lobby", 401}
	CryptoError					 = &ApiError{"Cryptography error", 500}
	InvalidToken				 = &ApiError{"Invalid access token", 401}
	PlayerNotFound				 = &ApiError{ "Player has not been found", 404}
)

func DatabaseError(err error) error {
	return New("Database error: " + err.Error(),500)
}