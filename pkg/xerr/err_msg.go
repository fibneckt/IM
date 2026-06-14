package xerr

var codeText = map[int]string{
	SERVER_COMMON_ERROR: "服务器异常，请稍候再试",
	REQUEST_PARAM_ERROR: "请求参数有误",
	DB_ERROR:            "数据库繁忙，请稍候再试",
}

func ErrMsg(errcode int) string {
	if msg, ok := codeText[errcode]; ok {
		return msg
	}
	return codeText[SERVER_COMMON_ERROR]
}
