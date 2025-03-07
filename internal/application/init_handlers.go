package application

import (
	"context"
	"net"

	"github.com/neekrasov/kvdb/internal/database"
	"github.com/neekrasov/kvdb/internal/delivery/tcp"
)

func initOnConnectHandler(bufferSize int, db *database.Database) tcp.ConnectionHandler {
	return func(ctx context.Context, sessionID string, conn net.Conn) error {
		buffer := make([]byte, bufferSize)
		n, err := tcp.Read(conn, buffer, bufferSize)
		if err != nil {
			return err
		}

		_, err = db.Login(ctx, sessionID, string(buffer[:n]))
		if err != nil {
			_, err = conn.Write([]byte(database.WrapError(err)))
			if err != nil {
				return err
			}

			return nil
		}

		_, err = conn.Write([]byte(database.WrapOK("authentication successful")))
		if err != nil {
			return err
		}

		return nil
	}
}
func initOnDisconnectHandler(db *database.Database) tcp.ConnectionHandler {
	return func(ctx context.Context, sessionID string, conn net.Conn) error {
		db.Logout(ctx, sessionID)
		return nil
	}
}

func initQueryHandler(db *database.Database) tcp.Handler {
	return func(ctx context.Context, sessionID string, request []byte) []byte {
		return []byte(db.HandleQuery(ctx, sessionID, string(request)))
	}
}
