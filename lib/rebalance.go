package yajirobe

import (
	"math"
)

// RebalancingBuy リバランス購入 購入金額を調整し売却せずに積み立てながらリバランスする場合の計算
func (a *AssetAllocation) RebalancingBuy(cost float64) map[AssetClass]float64 {
	// 1, 追加資金を入れた後のアセットアロケーションの目標金額を計算し、現在の評価額と差分をとる。
	// 2, 目標の金額に不足している資産クラスについて、追加投資する。
	//    追加投資額は、不足している資産クラスの差分を
	//
	// 現在 1000万の資産があるとして、追加資金が100万とすると、以下のようになる。
	// 目標率    25%  30%  45%          目標とするアセットアロケーション
	// 目標額    275  330  495  (1100)  追加投資後の目標の金額
	// 現在率    20%  20%  60%          現在の保有率
	// 現在額    200  200  600  (1000)  現在の評価額
	// 差分      -75 -130 +105          現在額の目標額からの差分
	// 追加額     37   63               追加投資の金額
	//                 ↑(130/(75+130))*100   超えていない分の割合 = 130/(75+130)
	//           ↑(75/(75+130))*100

	// 追加資金300万のパターン
	// 目標率    25%  30%  45%
	// 目標額    325  390  585  (1300)
	// 現在率    20%  20%  60%
	// 現在額    200  200  600  (1000)
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

	sum := 0.0
	adds := make(map[AssetClass]float64, len(keys))
	for i, c := range diffs {
		if c < 0 {
			// 端数丸め
			x := round(-c / shortfail * cost)
			adds[keys[i]] = x
			sum += x
		}
	}

	// 丸め誤差を足しておく
	// 足す対象は、追加投資をするクラスのうち、AssetClasses順にみて先頭に出現するものと決めておく
	if sum != cost {
		for _, c := range AssetClasses {
			if v, e := adds[c]; e && v != 0 {
				adds[c] += cost - sum
				break
			}
		}
	}

	return adds
}

func round(n float64) float64 {
	// Round half to even, aka banker's rounding
	// https://en.wikipedia.org/wiki/Rounding#Round_half_to_even
	// https://en.wikipedia.org/wiki/Nearest_integer_function

	return math.Ceil((n-0.5)/2) + math.Floor((n+0.5)/2)
}
