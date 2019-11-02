package setting

import (
    "log"
    "time"

    "github.com/go-ini/ini"
)

type App struct {
    JwtSecret string
    PageSize  int
    PrefixUrl string

    RuntimeRootPath string

    ImageSavePath  string
    ImageMaxSize   int
    ImageAllowExts []string

    ExportSavePath string
    QrCodeSavePath string
    FontSavePath   string

    LogSavePath string
    LogSaveName string
    LogFileExt  string
    TimeFormat  string

    DefaultRedirectHost string
    DefaultRedirectPort string
    DefaultRedirectTimeout time.Duration
}

var AppSetting = &App{}

type Server struct {
    RunMode      string
    HttpPort     string
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
}

var ServerSetting = &Server{}

type Database struct {
    Type        string
    User        string
    Password    string
    Host        string
    Name        string
    TablePrefix string
}

var DatabaseSetting = &Database{}

type Redis struct {
    Host        string
    Password    string
    Name        int
    MaxIdle     int
    MaxActive   int
    IdleTimeout time.Duration
}

var RedisSetting = &Redis{}

var cfg *ini.File

// mapTo map section
func mapTo(section string, v interface{}) {
    err := cfg.Section(section).MapTo(v)
    if err != nil {
        log.Fatalf("Cfg.MapTo RedisSetting err: %v", err)
    }
}

// Setup initialize the configuration instance
func Setup() {
    var err error
    log.Printf("[info] setting.Init")

    cfg, err = ini.Load("conf/cfg.ini")
    if err != nil {
        log.Fatalf("setting.Setup, fail to parse 'conf/cfg.ini': %v", err)
    }

    mapTo("app", AppSetting)
    mapTo("server", ServerSetting)
    mapTo("database", DatabaseSetting)
    mapTo("redis", RedisSetting)

    AppSetting.ImageMaxSize *= 1024 * 1024
    AppSetting.DefaultRedirectTimeout *= time.Second
    ServerSetting.ReadTimeout *= time.Second
    ServerSetting.WriteTimeout *= time.Second
    RedisSetting.IdleTimeout *= time.Second
}
