// Copyright 2024 Syntio Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/dataphos/aquarium-janitor-standalone-sr/internal/errcodes"
	"github.com/dataphos/lib-logger/logger"
)

func RequestLogger(log logger.Log) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				fields := logger.F{
					"method":         r.Method,
					"path":           r.URL.Path,
					"remote_adrr":    r.RemoteAddr,
					"status":         strconv.Itoa(ww.Status()),
					"content_length": strconv.FormatInt(r.ContentLength, 10),
					"bytes":          strconv.Itoa(ww.BytesWritten()),
					"response_time":  strconv.FormatInt(time.Since(t1).Milliseconds(), 10),
				}

				status := ww.Status()
				if status >= 100 && status < 400 {
					log.Infow("request completed", fields)
				} else {
					log.Errorw("request not completed successfully", errcodes.FromHttpStatusCode(status), fields)
				}
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
