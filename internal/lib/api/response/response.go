package httpresp

import (
	"net/http"

	"github.com/go-chi/render"
)


type ErrorCode string

const (
	CodeTeamExists ErrorCode = "TEAM_EXISTS"
	CodePRExists ErrorCode = "PR_EXISTS"
	CodePRMerged ErrorCode = "PR_MERGED"
	CodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	CodeNoCandidate ErrorCode = "NO_CANDIDATE"
	CodeNotFound ErrorCode = "NOT_FOUND"
)

type SuccessResponse struct {
	Status string `json:"status"`
	Data interface{} `json:"data,omitempty"`
}

func OK(data interface{}) SuccessResponse {
	return SuccessResponse{
		Status: "OK",
		Data: data,
	}
}

type ErrorResponse struct {
	Status string `json:"status"` 
	Error struct {
		Code ErrorCode `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func Error(code ErrorCode, msg string) ErrorResponse {
	resp := ErrorResponse{}
	resp.Status = "ERROR"
	resp.Error.Code = code
	resp.Error.Message = msg
	return resp
}


func WriteOK(w http.ResponseWriter, r *http.Request, data interface{}) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, OK(data))
}

func WriteError(w http.ResponseWriter, r *http.Request, httpStatus int, code ErrorCode, msg string) {
	render.Status(r, httpStatus)
	render.JSON(w, r, Error(code, msg))
}
