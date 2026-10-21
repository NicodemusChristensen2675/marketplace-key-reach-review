package main

import "testing"

func TestMarketplaceReach(t *testing.T) {
	in := ReviewInput{SellerAssets: []Activity{{"catalog", []string{"assets:read"}}}, BuyerUpdates: []Activity{{"notify", []string{"buyers:write"}}}, OrderHandoff: []Activity{{"transfer", []string{"orders:read", "orders:write"}}}}
	keys := []Key{{ID: "seller", Scopes: []string{"assets:read", "orders:read"}}, {ID: "fulfillment", Scopes: []string{"orders:read", "orders:write"}}}
	cases := []struct {
		key, activity string
		allowed       bool
	}{
		{"seller", "catalog", true}, {"seller", "notify", false}, {"seller", "transfer", false},
		{"fulfillment", "catalog", false}, {"fulfillment", "notify", false}, {"fulfillment", "transfer", true},
	}
	rows := evaluate(in, keys)
	for i, tc := range cases {
		row := rows[i]
		if row.KeyID != tc.key || row.Activity != tc.activity || row.Allowed != tc.allowed {
			t.Errorf("row %d: got %+v, want %+v", i, row, tc)
		}
	}
}
