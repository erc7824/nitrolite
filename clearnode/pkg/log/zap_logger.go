package log

import (
	"os"
	"path/filepath"
	"time"

	zaplogfmt "github.com/jsternberg/zap-logfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ Logger = &ZapLogger{}

// ZapLogger is a logger implementation backed by Uber's zap logger.
// It provides structured logging with high performance and supports
// various output formats and destinations.
type ZapLogger struct {
	lg            *zap.SugaredLogger
	keysAndValues []any
}

// Config is used to configure the ZapLogger.
// It supports environment variable configuration with default values.
type Config struct {
	Format string `env:"LOG_FORMAT" env-default:"console"` // console, logfmt or json
	Level  Level  `env:"LOG_LEVEL" env-default:"info"`     // debug, info, warn, error, fatal, trace
	Output string `env:"LOG_OUTPUT" env-default:"stderr"`  // stderr, stdout or file path
}

// NewZapLogger creates a new ZapLogger with the given configuration.
// It supports multiple output formats (console, logfmt, json) and destinations (stderr, stdout, file).
// Additional write syncers can be provided to write logs to multiple destinations.
func NewZapLogger(conf Config, extraWriters ...zapcore.WriteSyncer) Logger {
	// Create a production encoder config and customize time format.
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(ts.UTC().Format(time.RFC3339))
	}

	// Choose the encoder based on the config.
	var encoder zapcore.Encoder
	switch conf.Format {
	case "logfmt":
		encoder = zaplogfmt.NewEncoder(encCfg)
	case "json":
		encoder = zapcore.NewJSONEncoder(encCfg)
	default:
		encoder = zapcore.NewConsoleEncoder(encCfg)
	}

	var ws zapcore.WriteSyncer
	if conf.Output == "" || conf.Output == "stderr" {
		ws = zapcore.Lock(os.Stderr)
	} else if conf.Output == "stdout" {
		ws = zapcore.Lock(os.Stdout)
	} else {
		dir := filepath.Dir(conf.Output)
		err1 := os.MkdirAll(dir, 0755) // 0755 gives read/write/execute permissions to the owner, and read/execute permissions to others

		// Open the specified file; fallback to stderr on error.
		file, err2 := os.OpenFile(conf.Output, os.O_RDWR|os.O_CREATE, 0666)
		if err1 != nil || err2 != nil {
			ws = zapcore.Lock(os.Stdout)
		} else {
			ws = zapcore.AddSync(file)
		}
	}
	wss := zapcore.NewMultiWriteSyncer(append(extraWriters, ws)...)

	// Build the core.
	core := zapcore.NewCore(encoder, wss, toZapLogLevel(conf.Level))
	// Create a SugaredLogger; AddCallerSkip(2) skips wrapper methods in the call stack.
	zl := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2)).Sugar()

	return &ZapLogger{
		lg: zl,
	}
}

// Debug logs a message at debug level.
func (l *ZapLogger) Debug(msg string, keysAndValues ...any) {
	l.log(LevelDebug, msg, keysAndValues...)
}

// Info logs a message at info level.
func (l *ZapLogger) Info(msg string, keysAndValues ...any) {
	l.log(LevelInfo, msg, keysAndValues...)
}

// Warn logs a message at warn level.
func (l *ZapLogger) Warn(msg string, keysAndValues ...any) {
	l.log(LevelWarn, msg, keysAndValues...)
}

// Error logs a message at error level.
func (l *ZapLogger) Error(msg string, keysAndValues ...any) {
	l.log(LevelError, msg, keysAndValues...)
}

// Fatal logs a message at fatal level.
func (l *ZapLogger) Fatal(msg string, keysAndValues ...any) {
	l.log(LevelFatal, msg, keysAndValues...)
}

func (l *ZapLogger) log(level Level, msg string, keysAndValues ...any) {
	l.lg.Logw(toZapLogLevel(level), msg, keysAndValues...)
}

// WithKV returns a new ZapLogger with the key-value pair added to all future log messages.
func (l *ZapLogger) WithKV(key string, value any) Logger {
	return &ZapLogger{
		lg:            l.lg.With(key, value),
		keysAndValues: append(l.keysAndValues, key, value),
	}
}

// GetAllKV returns all key-value pairs that have been added to this logger instance.
func (l *ZapLogger) GetAllKV() []any {
	return l.keysAndValues
}

// WithName returns a new ZapLogger with the given name.
// The name is added to the logger hierarchy separated by dots.
func (l *ZapLogger) WithName(name string) Logger {
	return &ZapLogger{
		lg:            l.lg.Named(name),
		keysAndValues: l.keysAndValues,
	}
}

// Name returns the current name of the logger.
func (l *ZapLogger) Name() string {
	return l.lg.Desugar().Name()
}

// AddCallerSkip returns a new ZapLogger that skips additional stack frames when determining the caller.
func (l *ZapLogger) AddCallerSkip(skip int) Logger {
	return &ZapLogger{
		lg:            l.lg.WithOptions(zap.AddCallerSkip(skip)),
		keysAndValues: l.keysAndValues,
	}
}

func toZapLogLevel(logLevel Level) zapcore.Level {
	var zapLevel zapcore.Level
	switch logLevel {
	case LevelDebug:
		zapLevel = zapcore.DebugLevel
	case LevelInfo:
		zapLevel = zapcore.InfoLevel
	case LevelWarn:
		zapLevel = zapcore.WarnLevel
	case LevelError:
		zapLevel = zapcore.ErrorLevel
	case LevelFatal:
		zapLevel = zapcore.FatalLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	return zapLevel
}
