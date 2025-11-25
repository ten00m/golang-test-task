package response

type Response struct {
	Status string `json:"status"`
	Error  Error  `json:"error,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const (
	StatusOK    = "OK"
	StatusError = "ERROR"

	CodeTeamExists  = "TEAM_EXISTS"
	CodePRExists    = "PR_EXISTS"
	CodePRMerged    = "PR_MERGED"
	CodeNotAssigned = "NOT_ASSIGNED"
	CodeNoCandidate = "NO_CANDIDATE"
	CodeNotFound    = "NOT_FOUND"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func ErrorResponse(errMsg string, code string) Response {
	return Response{
		Status: StatusError,
		Error:  Error{Code: code, Message: errMsg},
	}
}
