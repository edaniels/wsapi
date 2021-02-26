package wsapi

import (
	"context"

	"nhooyr.io/websocket"
)

type Conn interface {
	Close()
	SendCommand(ctx context.Context, name string) (JSONResponse, error)
}

type conn struct {
	wsConn *websocket.Conn
}

func Dial(ctx context.Context, address string) (Conn, error) {
	wsConn, _, err := websocket.Dial(ctx, address, nil)
	if err != nil {
		return nil, err
	}
	wsConn.SetReadLimit(10 * (1 << 24))
	return &conn{wsConn: wsConn}, nil
}

func (c *conn) SendCommand(ctx context.Context, name string) (JSONResponse, error) {
	if err := WriteCommand(ctx, NewCommand(name), c.wsConn); err != nil {
		return nil, err
	}
	return ReadJSONCommandResponse(ctx, c.wsConn)
}

func (c *conn) Close() {
	c.wsConn.Close(websocket.StatusNormalClosure, "")
}
