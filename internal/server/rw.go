package server

import "net/http"

// customResponseWriter implements http.ResponseWriter, http.Pusher, http.Flusher
type customResponseWriter struct {
	w           http.ResponseWriter
	wroteHeader bool
	headerCfg   HeaderCfg
}

func (s *customResponseWriter) Header() http.Header {
	return s.w.Header()
}

func (s *customResponseWriter) WriteHeader(statusCode int) {
	s.handleHeaders()
	s.w.WriteHeader(statusCode)
}

func (s *customResponseWriter) Write(b []byte) (int, error) {
	s.handleHeaders() // for when WriteHeader is not called
	return s.w.Write(b)
}

// Push implements the http.Pusher interface.
func (s *customResponseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := s.w.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

// Flush implements the http.Flusher interface.
func (s *customResponseWriter) Flush() {
	f, ok := s.w.(http.Flusher)
	if ok {
		f.Flush()
	}
}

func (s *customResponseWriter) handleHeaders() {
	if s.wroteHeader == false {
		s.w.Header().Set("Server", "Gargoyle 0.0.1")
		for k, v := range s.headerCfg.Add {
			s.w.Header().Set(k, v)
		}
		for _, v := range s.headerCfg.Remove {
			s.w.Header().Del(v)
		}
		s.wroteHeader = true
	}
}
