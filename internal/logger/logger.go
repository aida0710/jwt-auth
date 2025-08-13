package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// Logger ロギングのインターフェース
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, err error, fields ...Field)
	Fatal(ctx context.Context, msg string, err error, fields ...Field)
	With(fields ...Field) Logger
}

// Field ログフィールドを表す構造体
type Field struct {
	Key   string
	Value interface{}
}

// F Fieldを作成するヘルパー関数
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Level ログレベル
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String ログレベルを文字列に変換
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel 文字列からログレベルを解析
func ParseLevel(s string) Level {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return DebugLevel
	case "INFO":
		return InfoLevel
	case "WARN", "WARNING":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	case "FATAL":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// logger Loggerインターフェースの実装
type logger struct {
	level  Level
	format string
	output io.Writer
	fields []Field
}

// NewLogger 新しいロガーを作成
func NewLogger(level, format string) Logger {
	return &logger{
		level:  ParseLevel(level),
		format: format,
		output: os.Stdout,
		fields: []Field{},
	}
}

// With フィールドを追加した新しいロガーを返す
func (l *logger) With(fields ...Field) Logger {
	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)

	return &logger{
		level:  l.level,
		format: l.format,
		output: l.output,
		fields: newFields,
	}
}

// Debug デバッグログを出力
func (l *logger) Debug(ctx context.Context, msg string, fields ...Field) {
	if l.level <= DebugLevel {
		l.log(ctx, DebugLevel, msg, nil, fields...)
	}
}

// Info 情報ログを出力
func (l *logger) Info(ctx context.Context, msg string, fields ...Field) {
	if l.level <= InfoLevel {
		l.log(ctx, InfoLevel, msg, nil, fields...)
	}
}

// Warn 警告ログを出力
func (l *logger) Warn(ctx context.Context, msg string, fields ...Field) {
	if l.level <= WarnLevel {
		l.log(ctx, WarnLevel, msg, nil, fields...)
	}
}

// Error エラーログを出力
func (l *logger) Error(ctx context.Context, msg string, err error, fields ...Field) {
	if l.level <= ErrorLevel {
		l.log(ctx, ErrorLevel, msg, err, fields...)
	}
}

// Fatal 致命的エラーログを出力してプログラムを終了
func (l *logger) Fatal(ctx context.Context, msg string, err error, fields ...Field) {
	l.log(ctx, FatalLevel, msg, err, fields...)
	os.Exit(1)
}

// log ログを出力する内部メソッド
func (l *logger) log(ctx context.Context, level Level, msg string, err error, fields ...Field) {
	// 呼び出し元の情報を取得
	_, file, line, _ := runtime.Caller(2)

	// ファイル名を短縮
	parts := strings.Split(file, "/")
	if len(parts) > 2 {
		file = strings.Join(parts[len(parts)-2:], "/")
	}

	// タイムスタンプ
	timestamp := time.Now().Format(time.RFC3339)

	// すべてのフィールドを結合
	allFields := make([]Field, 0, len(l.fields)+len(fields)+6)
	allFields = append(allFields, l.fields...)
	allFields = append(allFields, fields...)
	allFields = append(allFields,
		F("timestamp", timestamp),
		F("level", level.String()),
		F("message", msg),
		F("file", file),
		F("line", line),
	)

	// リクエストIDがあれば追加
	if requestID := getRequestID(ctx); requestID != "" {
		allFields = append(allFields, F("request_id", requestID))
	}

	// エラーがあれば追加
	if err != nil {
		allFields = append(allFields, F("error", err.Error()))
	}

	// フォーマットに応じて出力
	if l.format == "json" {
		l.outputJSON(allFields)
	} else {
		l.outputText(allFields)
	}
}

// outputJSON JSON形式でログを出力
func (l *logger) outputJSON(fields []Field) {
	logEntry := make(map[string]interface{})
	for _, field := range fields {
		logEntry[field.Key] = field.Value
	}

	data, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}

	_, err = fmt.Fprintln(l.output, string(data))
	if err != nil {
		log.Printf("Failed to write log entry: %v", err)
		return
	}
}

// outputText テキスト形式でログを出力
func (l *logger) outputText(fields []Field) {
	var sb strings.Builder

	// 基本情報を先に出力
	for _, field := range fields {
		switch field.Key {
		case "timestamp":
			sb.WriteString(fmt.Sprintf("[%s] ", field.Value))
		case "level":
			sb.WriteString(fmt.Sprintf("%-5s ", field.Value))
		case "message":
			sb.WriteString(fmt.Sprintf("%s ", field.Value))
		}
	}

	// その他のフィールドを出力
	var extras []string
	for _, field := range fields {
		switch field.Key {
		case "timestamp", "level", "message":
			// 既に出力済み
		default:
			extras = append(extras, fmt.Sprintf("%s=%v", field.Key, field.Value))
		}
	}

	if len(extras) > 0 {
		sb.WriteString("| ")
		sb.WriteString(strings.Join(extras, " "))
	}

	_, err := fmt.Fprintln(l.output, sb.String())
	if err != nil {
		log.Printf("Failed to write log entry: %v", err)
		return
	}
}

// getRequestID コンテキストからリクエストIDを取得
func getRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if reqID, ok := ctx.Value("request_id").(string); ok {
		return reqID
	}

	// EchoフレームワークのリクエストIDも試す
	if reqID, ok := ctx.Value("echo_request_id").(string); ok {
		return reqID
	}

	return ""
}
