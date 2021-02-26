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

func WriteJSONCommandResponse(ctx context.Context, resp CommandResponse, conn *websocket.Conn) error {
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

type JSONResponse json.RawMessage

func (jr JSONResponse) Unmarshal(into interface{}) error {
	if len(jr) == 0 {
		return nil
	}
	return json.Unmarshal([]byte(jr), into)
}

func ReadJSONResponse(ctx context.Context, conn *websocket.Conn) (JSONResponse, error) {
	_, data, err := conn.Read(ctx)
	if err != nil {
		return nil, err
	}
	temp := struct {
		Success bool            `json:"success"`
		Error   string          `json:"error,omitempty"`
		Result  json.RawMessage `json:"result,omitempty"`
	}{}

	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, err
	}
	if temp.Error != "" {
		return nil, errors.New(temp.Error)
	}
	return JSONResponse(temp.Result), nil
}

func ExpectResponse(ctx context.Context, conn *websocket.Conn) error {
	_, err := ReadJSONResponse(ctx, conn)
	return err
}

type CommandResponse struct {
	Success bool        `json:"success"`
	Error   error       `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

func (cr CommandResponse) MarshalJSON() ([]byte, error) {
	temp := struct {
		Success bool        `json:"success"`
		Error   string      `json:"error,omitempty"`
		Result  interface{} `json:"result,omitempty"`
	}{}
	temp.Success = cr.Success
	temp.Result = cr.Result
	if cr.Error != nil {
		temp.Error = cr.Error.Error()
	}
	return json.Marshal(temp)
}

func NewSuccessfulCommandResponse(result interface{}) CommandResponse {
	return CommandResponse{true, nil, result}
}

func NewErrorCommandResponse(err error) CommandResponse {
	return CommandResponse{false, err, nil}
}
