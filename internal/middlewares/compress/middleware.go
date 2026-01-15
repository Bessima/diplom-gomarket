package compress

import (
	"log"
	"net/http"
	"strings"
)

var compressContentTypes = []string{"application/json", "text/html"}

func isAllowCompressContentType(contentType string) bool {
	for _, item := range compressContentTypes {
		if item == contentType {
			return true
		}
	}
	return false
}

func GZIPMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")

		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		handler.ServeHTTP(ow, r)
	})
}
