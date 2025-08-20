package models

type Order struct {
	ID    string   `json:"id"`
	Items []string `json:"items"`
}
