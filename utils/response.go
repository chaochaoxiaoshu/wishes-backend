package utils

import (
	"time"

	"github.com/gin-gonic/gin"
)

func CreateResponse(result any, errMsg ...string) gin.H {
	response := gin.H{
		"errcode":   0,
		"errmsg":    "ok",
		"result":    result,
		"timestamp": time.Now().Unix(),
	}

	if len(errMsg) > 0 && errMsg[0] != "" {
		response["errcode"] = 1
		response["errmsg"] = errMsg[0]
	}

	return response
}
