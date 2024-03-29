package e

var MsgFlags = map[int]string{
    OK:                              "ok",
    ERROR:                           "fail",
    INVALID_PARAMS:                  "请求参数错误",
    ERROR_AUTH_CHECK_TOKEN_FAIL:     "Token鉴权失败",
    ERROR_AUTH_CHECK_TOKEN_TIMEOUT:  "Token已超时",
    ERROR_AUTH_TOKEN:                "Token生成失败",
    ERROR_AUTH:                      "Token错误",
    ERROR_UPLOAD_SAVE_IMAGE_FAIL:    "保存镜像失败",
    ERROR_UPLOAD_CHECK_IMAGE_FAIL:   "检查镜像失败",
    ERROR_UPLOAD_CHECK_IMAGE_FORMAT: "校验镜像错误，镜像格式或大小有问题",

    ERR_CFG_NO_AVAIL_POOL:           "Please double check the available resource pool assigned to this user!",
}

// GetMsg get error information based on Code
func GetMsg(code int) string {
    msg, ok := MsgFlags[code]
    if ok {
        return msg
    }

    return MsgFlags[ERROR]
}
