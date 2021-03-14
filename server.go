package wsapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edaniels/golog"
	"nhooyr.io/websocket"
)

type Server interface {
	HTTPHandler() http.Handler
	RegisterCommand(name string, handler CommandHandler)
	SetLogger(logger golog.Logger)
}

func NewServer() Server {
	return &server{commands: map[string]CommandHandler{}, logger: golog.Global}
}

type CommandHandler interface {
	Handle(ctx context.Context, cmd *Command) (interface{}, error)
}

type CommandHandlerFunc func(ctx context.Context, cmd *Command) (interface{}, error)

func (chf CommandHandlerFunc) Handle(ctx context.Context, cmd *Command) (interface{}, error) {
	return chf(ctx, cmd)
}

type server struct {
	commands map[string]CommandHandler
	logger   golog.Logger
}

func (s *server) SetLogger(logger golog.Logger) {
	s.logger = logger
}

func (s *server) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			s.logger.Error("error making websocket connection", "error", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		for {
			select {
			case <-r.Context().Done():
				return
			default:
			}

			cmd, err := ReadCommand(r.Context(), conn)
			if err != nil {
				s.logger.Errorw("error reading command", "error", err)
				return
			}
			result, err := s.handleCommand(r.Context(), cmd)
			if err != nil {
				resp := NewErrorCommandResponse(err)
				if err := WriteJSONCommandResponse(r.Context(), resp, conn); err != nil {
					s.logger.Errorw("error writing", "error", err)
					continue
				}
				continue
			}
			if err := WriteJSONCommandResponse(r.Context(), NewSuccessfulCommandResponse(result), conn); err != nil {
				s.logger.Errorw("error writing", "error", err)
				continue
			}
		}
	})
}

func (s *server) RegisterCommand(name string, handler CommandHandler) {
	s.commands[name] = handler
}

func (s *server) handleCommand(ctx context.Context, cmd *Command) (interface{}, error) {
	handler, ok := s.commands[cmd.Name]
	if !ok {
		return nil, fmt.Errorf("unknown command %s", cmd.Name)
	}
	return handler.Handle(ctx, cmd)
}
