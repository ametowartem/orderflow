package client

import (
	"context"

	inventoryv1 "github.com/ametowartem/orderflow/proto/inventory/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InventoryClient struct {
	client inventoryv1.InventoryServiceClient
}

func NewInventoryClient(addr string) (*InventoryClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}
	return &InventoryClient{client: inventoryv1.NewInventoryServiceClient(conn)}, nil
}

func (c *InventoryClient) CheckStock(ctx context.Context, productID string, quantity int32) (bool, error) {
	resp, err := c.client.CheckStock(ctx, &inventoryv1.CheckStockRequest{
		ProductId: productID,
		Quantity:  quantity,
	})

	if err != nil {
		return false, err
	}

	return resp.Available, nil
}
