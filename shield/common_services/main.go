package common_services

import "encoding/json"

func BuildErrorResponse(err bool, msg string, data interface{}, status int) HandlerResponse {
	var handlerResp HandlerResponse

	byteSlice, _ := json.Marshal(data)
	stringBody := string(byteSlice)

	handlerResp.Err = err
	handlerResp.Status = status
	handlerResp.Data = stringBody

	return handlerResp

}

func BuildResponse(err bool, msg string, data interface{}, status int) HandlerResponse {
	var handlerResp HandlerResponse

	byteSlice, _ := json.Marshal(data)
	stringBody := string(byteSlice)

	handlerResp.Err = err
	handlerResp.Status = status
	handlerResp.Data = stringBody

	return handlerResp

}
