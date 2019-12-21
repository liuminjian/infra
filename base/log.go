package base

import (
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/mattn/go-colorable"
	"github.com/pkg/errors"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"github.com/tietang/go-utils"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"time"
)

var formatter *prefixed.TextFormatter

var lfh *utils.LineNumLogrusHook

func InitLog(baseLogPath string, level log.Level, maxAge time.Duration, rotationTime time.Duration) {
	formatter = &prefixed.TextFormatter{}
	formatter.ForceColors = true
	formatter.DisableColors = false
	formatter.ForceFormatting = true
	formatter.SetColorScheme(&prefixed.ColorScheme{
		InfoLevelStyle:  "green",
		WarnLevelStyle:  "yellow",
		ErrorLevelStyle: "red",
		FatalLevelStyle: "41",
		PanicLevelStyle: "41",
		DebugLevelStyle: "blue",
		PrefixStyle:     "cyan",
		TimestampStyle:  "37",
	})
	formatter.FullTimestamp = true
	formatter.TimestampFormat = "2006-01-02.15:04:05.000000"
	log.SetFormatter(formatter)
	log.SetOutput(colorable.NewColorableStdout())
	log.SetReportCaller(true)
	log.SetLevel(level)
	SetLineNumLogrusHook()
	SetRotateLogsHook(baseLogPath, maxAge, rotationTime, formatter)
}

func SetLineNumLogrusHook() {
	lfh = utils.NewLineNumLogrusHook()
	lfh.EnableFileNameLog = true
	lfh.EnableFuncNameLog = true
	log.AddHook(lfh)
}

func SetRotateLogsHook(baseLogPath string, maxAge time.Duration, rotationTime time.Duration,
	formatter *prefixed.TextFormatter) {
	writer, err := rotatelogs.New(
		baseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(baseLogPath),      // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(maxAge),             // 文件最大保存时间
		rotatelogs.WithRotationTime(rotationTime), // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config local file system logger error. %v", errors.WithStack(err))
	}
	lfHook := lfshook.NewHook(lfshook.WriterMap{
		log.DebugLevel: writer, // 为不同级别设置不同的输出目的
		log.InfoLevel:  writer,
		log.WarnLevel:  writer,
		log.ErrorLevel: writer,
		log.FatalLevel: writer,
		log.PanicLevel: writer,
	}, formatter)
	log.AddHook(lfHook)
}

type GormLogger struct{}

func (*GormLogger) Print(v ...interface{}) {
	switch v[0] {
	case "sql":
		log.WithFields(
			log.Fields{
				"module":  "gorm",
				"type":    "sql",
				"rows":    v[5],
				"src_ref": v[1],
				"values":  v[4],
			},
		).Debug(v[3])
	case "log":
		log.WithFields(log.Fields{"module": "gorm", "type": "log"}).Print(v[2])
	}
}
