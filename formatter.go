package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gookit/color"
	"go.uber.org/zap"
)

type customJSONFormatter struct {
	attr []zap.Field
}

func newCustomJSONFormatter(params Params) *customJSONFormatter {
	res := &customJSONFormatter{}

	if params.AppName != "" {
		res.attr = append(res.attr, zap.String("app", params.AppName))
	}

	hostname, err := os.Hostname()
	if err == nil && hostname != "" {
		res.attr = append(res.attr, zap.String("host", hostname))
	}

	return res
}

var levelColors = map[string]func(...any) string{
	"DEBUG": color.Magenta.Render,
	"INFO":  color.Blue.Render,
	"WARN":  color.Yellow.Render,
	"ERROR": color.Red.Render,
}

func renderCrashLevel(a ...any) string {
	return color.Bold.Render(color.Red.Render(a...))
}

func (f *customJSONFormatter) getRuntimeData() (funcName, filename string, line int) {
	pc, filePath, line, ok := runtime.Caller(9)
	if !ok {
		return "unknown", "unknown", 0
	}
	filename = filepath.Base(filePath)
	funcPath := runtime.FuncForPC(pc).Name()
	funcName = funcPath[strings.LastIndex(funcPath, ".")+1:]

	if strings.Contains(funcName, "func") {
		funcBefore := funcPath[:strings.LastIndex(funcPath, ".")]
		funcName = funcPath[strings.LastIndex(funcBefore, ".")+1:]
	}
	return
}

func (f *customJSONFormatter) getDefaultFields() []zap.Field {
	return f.attr
}

func (f *customJSONFormatter) Format(logEntry string) string {
	var entry map[string]any
	if err := json.Unmarshal([]byte(logEntry), &entry); err != nil {
		return logEntry
	}

	entry = f.applyColorToLevel(entry)
	entry = f.formatRuntimeDetails(entry)
	return f.reorderFields(entry)
}

func (f *customJSONFormatter) applyColorToLevel(entry map[string]any) map[string]any {
	level, ok := entry["level"].(string)
	if !ok {
		return entry
	}

	// Crash entries stay visually distinct on a terminal while their level stays standard.
	if entry[fieldFatal] == true || entry[fieldCritical] == true {
		entry["level"] = renderCrashLevel(level)
		return entry
	}

	if colorFunc, exists := levelColors[level]; exists {
		entry["level"] = colorFunc(level)
	}
	return entry
}

func (f *customJSONFormatter) formatRuntimeDetails(entry map[string]any) map[string]any {
	if msg, ok := entry["msg"].(string); ok {
		funcName, _, _ := f.getRuntimeData()
		entry["msg"] = fmt.Sprintf("%s: %s", funcName, msg)
	}

	return entry
}

func (f *customJSONFormatter) reorderFields(entry map[string]any) string {
	orderedFields := []string{"time", "level", "msg", "file", "app", "host"}
	line := newJSONLine(len(entry) * 10)

	line.buf.WriteString("{")
	firstField := true

	for _, field := range orderedFields {
		if value, exists := entry[field]; exists {
			if !firstField {
				line.buf.WriteString(",")
			}
			firstField = false
			if field == "level" {
				line.writeLevel(value)
			} else {
				line.writeField(field, value)
			}
			delete(entry, field)
		}
	}

	for key, value := range entry {
		if !firstField {
			line.buf.WriteString(",")
		}
		firstField = false
		line.writeField(key, value)
	}

	line.buf.WriteString("}\n")
	return line.buf.String()
}

type jsonLine struct {
	buf     *bytes.Buffer
	encoder *json.Encoder
}

func newJSONLine(size int) jsonLine {
	buf := bytes.NewBuffer(make([]byte, 0, size))
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	return jsonLine{buf: buf, encoder: encoder}
}

// The level is ours and may carry ANSI color codes, which a terminal must receive unescaped.
func (l jsonLine) writeLevel(value any) {
	l.buf.WriteString(`"level":"`)
	fmt.Fprint(l.buf, value)
	l.buf.WriteString(`"`)
}

func (l jsonLine) writeField(key string, value any) {
	l.writeJSON(key)
	l.buf.WriteString(":")
	l.writeJSON(value)
}

// So a quote or a newline in a message never breaks the line.
func (l jsonLine) writeJSON(value any) {
	if err := l.encoder.Encode(value); err != nil {
		_ = l.encoder.Encode(fmt.Sprint(value))
	}
	l.buf.Truncate(l.buf.Len() - 1)
}
