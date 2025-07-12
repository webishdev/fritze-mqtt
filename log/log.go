package log

import (
	"encoding/xml"
	"log/slog"
	"os"
)

var textHandler = *slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	//Level: slog.LevelInfo,
})

var infoLogger = slog.NewLogLogger(&textHandler, slog.LevelInfo)
var warnLogger = slog.NewLogLogger(&textHandler, slog.LevelWarn)
var errorLogger = slog.NewLogLogger(&textHandler, slog.LevelError)
var debugLogger = slog.NewLogLogger(&textHandler, slog.LevelDebug)

func SetLogLevel(level slog.Level) {
	textHandler = *slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
}

func Info(format string, v ...any) {
	infoLogger.Printf(format, v...)
}

func Warn(format string, v ...any) {
	warnLogger.Printf(format, v...)
}

func Error(format string, v ...any) {
	errorLogger.Printf(format, v...)
}

func Debug(format string, v ...any) {
	debugLogger.Printf(format, v...)
}

func PrintXML(v interface{}) {
	xmlBytes, err := xml.Marshal(v)
	if err != nil {
		return
	}
	xmlString := string(xmlBytes)
	debugLogger.Println(xmlString)
}
