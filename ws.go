package wsapi

import (
	"context"
	"encoding/json"

	"nhooyr.io/websocket"
)

func WriteJSON(ctx context.Context, val interface{}, conn *websocket.Conn) error {
	md, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, md)
}

type JSONResponse json.RawMessage

func (jr JSONResponse) Unmarshal(into interface{}) error {
	if len(jr) == 0 {
		return nil
	}
	return json.Unmarshal([]byte(jr), into)
}
