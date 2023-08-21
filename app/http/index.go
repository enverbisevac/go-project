package http

import (
	"net/http"

	"github.com/enverbisevac/go-project/assets"
	"github.com/rs/zerolog/log"
)

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	index := "index.html"
	t, ok := assets.Templates[index]
	if !ok {
		log.Warn().Str("template", index).Msg("not found")
		return
	}

	data := make(map[string]interface{})

	if err := t.Execute(w, data); err != nil {
		log.Err(err).Stack().Send()
	}
}

func (s *Server) protected(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is a protected handler"))
}
