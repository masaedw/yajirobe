package yajirobe

// RebalancingBuy リバランス購入 購入金額を調整し売却せずに積み立てながらリバランスする場合の計算
func (a *AssetAllocation) RebalancingBuy(cost float64) map[AssetClass]float64 {
	// 追加資金を入れた後の割合から超えているものは無視して、
	// 超えていないものについて、超えていない割合に応じて追加投資する
	//
	// 現在 1000万の資産があるとして、追加資金が100万とし、以下の割合になっているとする。
	// 目標率    25%  30%  45%
	// 目標額    275  330  495  (1100)
	// 現在率    20%  20%  60%
	// 現在額    200  200  600  (1000)
	// 差分      -75 -130 +105
	// 追加額     37   63
	//                 ↑(70/(75+130))*100
	//           ↑(75/(75+130))*100

	// 追加資金300万のパターン
	// 目標率    25%  30%  45%
	// 目標額    325  390  585  (1300)
	// 現在率    20%  20%  60%
	// 現在額    200  200  600
	// 差分     -125 -190  +15
	// 追加額    119  181
	//                 ↑(190/(125+190))*300
	//           ↑(125/(125+190))*300

	// 追加資金400万のパターン
	// 目標率    25%  30%  45%
	// 目標額    350  420  630  (1400)
	// 現在率    20%  20%  60%
	// 現在額    200  200  600  (1000)
	// 差分     -150 -220  -30
	// 追加額    150  220   30

	// 必要なアセットクラス
	keys := a.keys()

	// 追加資金を入れた後の評価額
	total := a.cprice + cost

	// 不足分合計
	shortfail := 0.0

	// 差分
	diffs := make([]float64, len(keys))
	for i, c := range keys {
		d := a.details[c]
		tp := d.targetRatio * total
		sf := d.cprice - tp
		diffs[i] = sf
		if sf < 0 {
			shortfail += -sf
		}
	}

	adds := make(map[AssetClass]float64, len(keys))
	for i, c := range diffs {
		if c < 0 {
			adds[keys[i]] = -c / shortfail * cost
		}
	}

	return adds
}
