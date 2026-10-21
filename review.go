package main

import (
	"fmt"
	"strings"
)

type Activity struct {
	Name           string   `json:"name"`
	RequiredScopes []string `json:"required_scopes"`
}

type ReviewInput struct {
	RequestID    string     `json:"request_id"`
	SellerAssets []Activity `json:"seller_assets"`
	BuyerUpdates []Activity `json:"buyer_updates"`
	OrderHandoff []Activity `json:"order_handoff"`
}

type Key struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

type Reach struct {
	KeyID    string `json:"key_id"`
	KeyName  string `json:"key_name"`
	Area     string `json:"area"`
	Activity string `json:"activity"`
	Allowed  bool   `json:"allowed"`
}

func evaluate(in ReviewInput, keys []Key) []Reach {
	var rows []Reach
	for _, key := range keys {
		granted := make(map[string]bool, len(key.Scopes))
		for _, scope := range key.Scopes {
			granted[scope] = true
		}
		for _, group := range []struct {
			area       string
			activities []Activity
		}{
			{"seller_assets", in.SellerAssets}, {"buyer_updates", in.BuyerUpdates}, {"order_handoff", in.OrderHandoff},
		} {
			for _, activity := range group.activities {
				allowed := len(activity.RequiredScopes) > 0
				for _, scope := range activity.RequiredScopes {
					allowed = allowed && granted[scope]
				}
				rows = append(rows, Reach{key.ID, key.Name, group.area, activity.Name, allowed})
			}
		}
	}
	return rows
}

func document(date string, rows []Reach) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Marketplace access review: %s\n\n| Key ID | Key name | Area | Activity | Reach |\n|---|---|---|---|---|\n", date)
	for _, row := range rows {
		status := "no"
		if row.Allowed {
			status = "yes"
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", cell(row.KeyID), cell(row.KeyName), cell(row.Area), cell(row.Activity), status)
	}
	return b.String()
}

func cell(s string) string { return strings.NewReplacer("|", "\\|", "\n", " ", "\r", " ").Replace(s) }
