package errcodes

const (
	DatabaseConnectionInitialization = 100
	InvalidDatabaseState             = 101
	DatabaseInitialization           = 102
	ServerInitialization             = 103
	ExternalCheckerInitialization    = 104
	ServerShutdown                   = 200
	BadRequest                       = 400
	InternalServerError              = 500
	Miscellaneous                    = 999
)

func FromHttpStatusCode(code int) uint64 {
	switch {
	case code >= 400 && code < 500:
		return BadRequest
	case code >= 500:
		return InternalServerError
	default:
		return Miscellaneous
	}
}
