package handlers

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type LogEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	RequestBody  string    `json:"request_body"`
	Status       int       `json:"status"`
	ResponseBody string    `json:"response_body"`
}

var (
	requestLogs []LogEntry
	logMutex    sync.RWMutex
)

func StatsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Read Request Body
		var reqBodyBytes []byte
		if c.Request.Body != nil {
			reqBodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
		}

		// Custom Response Writer to capture response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// Record Log
		logMutex.Lock()
		requestLogs = append(requestLogs, LogEntry{
			Timestamp:    start,
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			RequestBody:  string(reqBodyBytes),
			Status:       c.Writer.Status(),
			ResponseBody: blw.body.String(),
		})
		// Keep log size manageable? Let's keep last 100 for now to avoid memory leak
		if len(requestLogs) > 100 {
			requestLogs = requestLogs[len(requestLogs)-100:]
		}
		logMutex.Unlock()
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func GetStats(c *gin.Context) {
	logMutex.RLock()
	defer logMutex.RUnlock()

	// Return logs in reverse order (newest first)
	reversedLogs := make([]LogEntry, len(requestLogs))
	for i, entry := range requestLogs {
		reversedLogs[len(requestLogs)-1-i] = entry
	}

	c.IndentedJSON(200, reversedLogs)
}
