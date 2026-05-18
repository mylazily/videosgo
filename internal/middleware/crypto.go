package middleware

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"videosgo/pkg/crypto"
)

// CryptoResponse 需要加密的响应结构
type CryptoResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Crypto XOR 加密中间件（对响应中的 m3u8 链接进行加密）
func Crypto(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 拦截响应
		blw := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// 检查是否需要加密
		contentType := c.Writer.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			return
		}

		// 解析响应
		var resp CryptoResponse
		if err := json.Unmarshal(blw.body.Bytes(), &resp); err != nil {
			return
		}

		// 加密 m3u8 链接
		encryptM3u8Links(&resp, key)

		// 重新写入响应
		encrypted, err := json.Marshal(resp)
		if err != nil {
			return
		}
		blw.body.Reset()
		blw.body.Write(encrypted)
	}
}

// responseBodyWriter 响应体写入器
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 写入响应
func (w *responseBodyWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

// encryptM3u8Links 递归加密响应中的 m3u8 链接
func encryptM3u8Links(data interface{}, key string) {
	switch v := data.(type) {
	case map[string]interface{}:
		for k, val := range v {
			if k == "url" || k == "play_url" || k == "m3u8" {
				if str, ok := val.(string); ok && strings.HasSuffix(str, ".m3u8") {
					if encrypted, err := crypto.XOREncrypt(str, key); err == nil {
						v[k] = encrypted
					}
				}
			} else {
				encryptM3u8Links(val, key)
			}
		}
	case []interface{}:
		for i := range v {
			encryptM3u8Links(v[i], key)
		}
	}
}
