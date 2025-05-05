package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type PrettyJSONWriter struct {
	writer io.Writer
}

func (w *PrettyJSONWriter) Write(p []byte) (n int, err error) {
	var buffer bytes.Buffer
	if err := json.Indent(&buffer, p, "", "  "); err != nil {
		return 0, err
	}
	buffer.WriteByte('\n')
	return w.writer.Write(buffer.Bytes())
}

func NewZapLogger() *zap.Logger {
	config := zap.NewProductionConfig()

	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.CallerKey = ""

	writer := &PrettyJSONWriter{writer: os.Stdout}

	return zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(config.EncoderConfig),
			zapcore.AddSync(writer),
			zapcore.InfoLevel,
		),
		zap.AddCaller(),
	)
}
