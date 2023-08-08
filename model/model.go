package model

import "time"

// I should use this format when returning json from server (nest the data json inside this json?)
type ResponseBody struct {
	Data    string `json:"data"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type FormAuth struct {
	Identity   string `json:"identity" xml:"identity" form:"identity"`
	Password   string `json:"password" xml:"password" form:"password"`
	RememberMe bool   `json:"remember_me" xml:"remember_me" form:"remember_me"`
}

type FormNewSale struct {
	Sale_TimeDate  string `json:"sale_time_date" xml:"sale_time_date" form:"sale_time_date"`
	PaymentMethod  string `json:"payment_type" xml:"payment_type" form:"payment_type"`
	Qty_FreshJuice int    `json:"fresh_juice_qty" xml:"fresh_juice_qty" form:"fresh_juice_qty"`
	Qty_CutFruit   int    `json:"cut_fruit_qty" xml:"cut_fruit_qty" form:"cut_fruit_qty"`
	Qty_RawFruit   int    `json:"raw_fruit_qty" xml:"raw_fruit_qty" form:"raw_fruit_qty"`
}

type JsonSale struct {
	ID          int       `json:"ID"`
	Amount      float32   `json:"amount"`
	Qty         float32   `json:"qty"` //this is float and not int bcos in case we plan to sell by weight, then it wouldnt make sense to use int
	PaymentType int       `json:"payment_type"`
	OperationID int       `json:"operation_id"`
	ItemID      int       `json:"item_id"`
	GroupSaleID int       `json:"group_sale_id"`
	CreatedAt   time.Time `json:"CreatedAt"`
	UpdatedAt   time.Time `json:"UpdatedAt"`
}

type ViewSale struct {
	ID          int    `json:"id"`
	Amount      string `json:"amount"`
	Qty         string `json:"quantity"` //this is float and not int bcos in case we plan to sell by weight, then it wouldnt make sense to use int
	PaymentType string `json:"payment_type"`
	Operation   string `json:"operation"`
	Item        string `json:"item"`
	Time        string `json:"time"`
	Date        string `json:"date"`
}

type ViewSalesReport struct {
	TotalGrossRevenue float64 `json:"total_gross_revenue"`
	TotalExpenses     float64 `json:"total_expenses"`
	TotalNetRevenue   float64 `json:"total_net_revenue"`
	IncomeTax         float64 `json:"income_tax"`
	GrantLoan         float64 `json:"grant_loan"`
	ProfitLoss        float64 `json:"profit_loss"`
}
