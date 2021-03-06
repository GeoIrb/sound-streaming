package player

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Client rpc player
type Client struct {
	hostLayout  string
	controlPort string
}

// State return all busy ports, devices on player and existing storage
func (c *Client) State(ctx context.Context, ip string) (ports, storages, devices []string, err error) {
	conn, err := grpc.Dial(
		fmt.Sprintf(c.hostLayout, ip, c.controlPort),
		// todo
		grpc.WithInsecure(),
	)
	if err != nil {
		return
	}
	defer conn.Close()

	if res, err := NewPlayerClient(conn).
		State(
			ctx,
			&StateRequest{},
		); err == nil {
		ports, storages, devices = res.Ports, res.Storages, res.Devices
	}
	return
}

// ReceiveStart rpc request to player with ip for start receive signal from server on port.
// UUID of the storage existing on the player
// if the storage with UUID does not exist or the UUID is zero, a new storage will be created on the player
// The signal will be stored in the storage sUUID
func (c *Client) ReceiveStart(ctx context.Context, ip, port string, uuid *string) (sUUID string, err error) {
	conn, err := grpc.Dial(
		fmt.Sprintf(c.hostLayout, ip, c.controlPort),
		// todo
		grpc.WithInsecure(),
	)
	if err != nil {
		return
	}
	defer conn.Close()
	req := &StartReceiveRequest{
		Port: port,
	}
	if uuid != nil {
		req.StorageUUID = &wrapperspb.StringValue{
			Value: *uuid,
		}
	}

	if res, err := NewPlayerClient(conn).
		ReceiveStart(
			ctx,
			req,
		); err == nil {
		sUUID = res.StorageUUID
	}
	return
}

// ReceiveStop rpc request to player with ip for stop receive signal from server on port.
func (c *Client) ReceiveStop(ctx context.Context, ip, port string) (err error) {
	conn, err := grpc.Dial(
		fmt.Sprintf(c.hostLayout, ip, c.controlPort),
		// todo
		grpc.WithInsecure(),
	)
	if err != nil {
		return
	}
	defer conn.Close()

	_, err = NewPlayerClient(conn).
		ReceiveStop(
			ctx,
			&StopReceiveRequest{
				Port: port,
			})
	return
}

// Play rpc request to player with ip for play audio signal from storage with UUID on deviceName
// channels, rate - playback options
func (c *Client) Play(ctx context.Context, ip, UUID, deviceName string, channels, rate, bitsPerSample uint32) (err error) {
	conn, err := grpc.Dial(
		fmt.Sprintf(c.hostLayout, ip, c.controlPort),
		// todo
		grpc.WithInsecure(),
	)
	if err != nil {
		return
	}
	defer conn.Close()

	_, err = NewPlayerClient(conn).
		Play(
			ctx,
			&StartPlayRequest{
				DeviceName:    deviceName,
				Channels:      channels,
				Rate:          rate,
				BitsPerSample: bitsPerSample,
				StorageUUID:   UUID,
			})
	return
}

// Stop rpc request to player with ip for stop audio
func (c *Client) Stop(ctx context.Context, playerIP, deviceName string) (err error) {
	conn, err := grpc.Dial(
		fmt.Sprintf(c.hostLayout, playerIP, c.controlPort),
		// todo
		grpc.WithInsecure(),
	)
	if err != nil {
		return
	}
	defer conn.Close()

	_, err = NewPlayerClient(conn).
		Stop(
			ctx,
			&StopPlayRequest{
				DeviceName: deviceName,
			},
		)
	return
}

// ClearStorage rpc request to player with ip for clear audio storage with UUID
func (c *Client) ClearStorage(ctx context.Context, ip, UUID string) (err error) {
	conn, err := grpc.Dial(
		fmt.Sprintf(c.hostLayout, ip, c.controlPort),
		// todo
		grpc.WithInsecure(),
	)
	if err != nil {
		return
	}
	defer conn.Close()

	_, err = NewPlayerClient(conn).
		ClearStorage(
			ctx,
			&ClearStorageRequest{
				StorageUUID: UUID,
			},
		)
	return
}

// NewClient ...
func NewClient(hostLayout, controlPort string) *Client {
	return &Client{
		hostLayout:  hostLayout,
		controlPort: controlPort,
	}
}
