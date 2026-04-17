package cloudfail

import (
	"fmt"

	"cloud-platform-api/app/Services/ai_robot/internal/deps"
)

func HandleCloudCallFailure(r deps.Responder, d *deps.Deps, errMsg string, status int, err error) bool {
	if err != nil {
		r.FrontFailed(d.Gin, errMsg, err)
		return true
	}
	if status >= 400 {
		r.FrontFailed(d.Gin, fmt.Sprintf("云平台返回 HTTP %d", status), nil)
		return true
	}
	return false
}
