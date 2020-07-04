package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cockroachdb/cockroach-go/crdb"

	"github.com/shijuvar/go-distsys/eventdriven/pb"
)

// QueryStoreRepository syncs data model to be used for query operations
// Because it's repository for read model, denormalized data can be inserted

type QueryStoreRepository struct{}

func (store QueryStoreRepository) SyncOrderQueryModel(order pb.OrderCreateCommand) error {

	// Run a transaction to sync the query model.
	err := crdb.ExecuteTx(context.Background(), db, nil, func(tx *sql.Tx) error {
		return createOrderQueryModel(tx, order)
	})
	if err != nil {
		return fmt.Errorf("error on syncing query repository: %w", err)
	}
	return nil
}

func createOrderQueryModel(tx *sql.Tx, order pb.OrderCreateCommand) error {

	// Insert order into the "orders" table.
	sql := `
INSERT INTO orders (id, customerid, status, createdon, restaurantid, amount) 
VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := tx.Exec(sql, order.OrderId, order.CustomerId, order.Status, order.CreatedOn, order.RestaurantId, order.Amount)
	if err != nil {
		return fmt.Errorf("error on insert into orders: %w", err)
	}
	// Insert order items into the "orderitems" table.
	// Because it's repository for read model, we can insert denormalized data
	for _, v := range order.OrderItems {
		sql = `
INSERT INTO orderitems (orderid, customerid, code, name, unitprice, quantity) 
VALUES ($1,$2,$3,$4,$5,$6)`

		_, err := tx.Exec(sql, order.OrderId, order.CustomerId, v.Code, v.Name, v.UnitPrice, v.Quantity)
		if err != nil {
			return fmt.Errorf("error on insert into order items: %w", err)
		}
	}
	return nil
}

// Approve order
func (store QueryStoreRepository) ChangeOrderStatus(orderId string, status string) error {
	sql := `
UPDATE orders 
SET status=$2
WHERE id=$1`

	_, err := db.Exec(sql, orderId, status)
	if err != nil {
		return fmt.Errorf("error on updating status of order: %w", err)
	}
	return nil
}
