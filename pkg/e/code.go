package e

import (
    "fmt"
    "os"
    "runtime/debug"
)

const (
    OK             = 200
    ERROR          = 500
    INVALID_PARAMS = 400

    ERROR_EXIST_TAG       = 10001
    ERROR_EXIST_TAG_FAIL  = 10002
    ERROR_NOT_EXIST_TAG   = 10003
    ERROR_GET_TAGS_FAIL   = 10004
    ERROR_COUNT_TAG_FAIL  = 10005
    ERROR_ADD_TAG_FAIL    = 10006
    ERROR_EDIT_TAG_FAIL   = 10007
    ERROR_DELETE_TAG_FAIL = 10008
    ERROR_EXPORT_TAG_FAIL = 10009
    ERROR_IMPORT_TAG_FAIL = 10010

    ERROR_NOT_EXIST_ARTICLE        = 10011
    ERROR_CHECK_EXIST_ARTICLE_FAIL = 10012
    ERROR_ADD_ARTICLE_FAIL         = 10013
    ERROR_DELETE_ARTICLE_FAIL      = 10014
    ERROR_EDIT_ARTICLE_FAIL        = 10015
    ERROR_COUNT_ARTICLE_FAIL       = 10016
    ERROR_GET_ARTICLES_FAIL        = 10017
    ERROR_GET_ARTICLE_FAIL         = 10018
    ERROR_GEN_ARTICLE_POSTER_FAIL  = 10019

    ERROR_AUTH_CHECK_TOKEN_FAIL    = 20001
    ERROR_AUTH_CHECK_TOKEN_TIMEOUT = 20002
    ERROR_AUTH_TOKEN               = 20003
    ERROR_AUTH                     = 20004

    ERROR_UPLOAD_SAVE_IMAGE_FAIL    = 30001
    ERROR_UPLOAD_CHECK_IMAGE_FAIL   = 30002
    ERROR_UPLOAD_CHECK_IMAGE_FORMAT = 30003




    //*********** for broker server ******************
	SUCC        = 0
	ERR_UNKNOWN = -1


	//msg, strictly is not errors
	ERR_MSG                 = -4000

    //user
    ERR_USR                 = -4100
    ERR_USR_ZONE_MISMATCH   = -4101    //Client zone name not match, please get the latest zone list
    ERR_USR_LOGIN_UNFIN     = -4102    //Current login has not finished, please wait
    ERR_USR_ABNORMAL        = -4103    //Current user is abnormal, please double check
    ERR_USR_CC              = -4104    //CC error
    ERR_USR_MUL_VM          = -4105    //multi auto login server
    ERR_USR_SFTP            = -4105

    //config
    ERR_CFG                 = -4200
    ERR_CFG_NO_AVAIL_POOL   = -4201    //Please double check the available resource pool assigned to this user!
    ERR_CFG_SHARED_VM       = -4202    //Please double check the Shared-VM client login configuration

    //cluster
    ERR_CLUSTER             = -4300
    ERR_CLUSTER_NOT_STABLE  = -4301    //Cluster is not stable, please try later

    //system error
    ERR_SYS                 = -4400    //System error, please contact service
    ERR_SYS_MYSQL           = -4401    //MySQL operation error, please call the service
    ERR_SYS_HW              = -4402    //system hardware issue
    ERR_SYS_NW              = -4403    //system net work
    ERR_SYS_NW_TIMEOUT      = -4404    //Request timeout exceeded, please wait for a while
    ERR_SYS_NW_CONN_REFUSED = -4405    //Connection refused, please double check the NetWork and Services

    //vm manager error
    ERR_VMM                 = -4500    //vm manager error
    ERR_VMM_ABNORMAL        = -4501    //vm manager abnormal

    //license
    ERR_LIC                 = -4600
    ERR_LIC_NOT_INIT        = -4601    //System License not inited
    ERR_LIC_NOT_VALID       = -4602    //License is not valid or expired
    ERR_LIC_CHECK_ERR       = -4603    //License check error, please try later
    ERR_LIC_REACH_LIMIT     = -4604    //License reach limit, please double confirm the license of current user pool
)

func Debug(err error) {
    defer func() {
        if p := recover(); p != nil {
            fmt.Printf("panic recover! p: %v", p)
            debug.PrintStack()
            os.Exit(-1)
        }
    }()
    fmt.Printf(err.Error())
    panic("error fucker")
}
