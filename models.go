package main

// Stock 銘柄
type Stock struct {
	Name                 string // 名称
	Code                 int    // 銘柄コード
	Amount               int    // 保有株数
	AcquisitionUnitPrice int64  // 取得単価
	CurrentUnitPrice     int64  // 現在値
	AcquisitionPrice     int64  // 取得金額
	CurrentPrice         int64  // 評価額
}
