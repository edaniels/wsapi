package wsapi

import (
	"context"
	"encoding/json"
	"errors"

	"nhooyr.io/websocket"
)

type Command struct {
	Name string `json:"name"`
}

func NewCommand(name string) *Command {
	return &Command{Name: name}
}

func WriteCommand(ctx context.Context, cmd *Command, conn *websocket.Conn) error {
	return WriteJSON(ctx, cmd, conn)
}

func WriteJSONResponse(ctx context.Context, resp Response, conn *websocket.Conn) error {
	return WriteJSON(ctx, resp, conn)
}

func WriteJSON(ctx context.Context, val interface{}, conn *websocket.Conn) error {
	md, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, md)
}

func ReadCommand(ctx context.Context, conn *websocket.Conn) (*Command, error) {
	_, data, err := conn.Read(ctx)
	if err != nil {
		return nil, err
	}
	var cmd Command
	if err := json.Unmarshal(data, &cmd); err != nil {
		return nil, err
	}
	return &cmd, nil
}

func ReadJSONResponse(ctx context.Context, conn *websocket.Conn, result interface{}) error {
	_, data, err := conn.Read(ctx)
	if err != nil {
		return err
	}
	temp := struct {
		Success bool            `json:"success"`
		Error   string          `json:"error,omitempty"`
		Result  json.RawMessage `json:"result,omitempty"`
	}{}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	if temp.Error != "" {
		return errors.New(temp.Error)
	}
	return json.Unmarshal([]byte(temp.Result), result)
}

type Response struct {
	Success bool        `json:"success"`
	Error   error       `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

func (r Response) MarshalJSON() ([]byte, error) {
	temp := struct {
		Success bool        `json:"success"`
		Error   string      `json:"error,omitempty"`
		Result  interface{} `json:"result,omitempty"`
	}{}
	temp.Success = r.Success
	temp.Result = r.Result
	if r.Error != nil {
		temp.Error = r.Error.Error()
	}
	return json.Marshal(temp)
}

func NewSuccessfulResponse(result interface{}) Response {
	return Response{true, nil, result}
}

func NewErrorResponse(err error) Response {
	return Response{false, err, nil}
}
