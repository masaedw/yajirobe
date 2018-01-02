package main

// Stock 銘柄
type Stock struct {
	Name                 string // 名称
	Code                 string // 銘柄コード
	Amount               int    // 保有株数
	AcquisitionUnitPrice int64  // 取得単価
	CurrentUnitPrice     int64  // 現在地
}

// AcquisitionPrice 取得金額
func (s Stock) AcquisitionPrice() int64 {
	return s.AcquisitionUnitPrice * int64(s.Amount)
}

// CurrentPrice 評価額
func (s Stock) CurrentPrice() int64 {
	return s.CurrentUnitPrice * int64(s.Amount)
}
