package yajirobe

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

func TestParseForeignStock(t *testing.T) {
	test := `
	<html><head></head><body><table>
	<tr class="even">
		<td>
			<img src="https://a248.e.akamai.net/f/248/29350/7d/pict.sbisec.co.jp/sbisec/images/base/foreign/i_flag01_usa_01.gif" width="22px" height="15px">
				<a href="#here" onclick="javascript:moveProduct('/czk/secCashBalance/','DIS','US','NYSE');">ウォルト ディズニー</a>
			<br>
			<div class="wfit75">
				DIS&nbsp;NYSE
			</div>
		</td>

		<td class="alC">
			<img width="36" height="26" src="https://a248.e.akamai.net/f/248/29350/7d/pict.sbisec.co.jp/sbisec/images/base/foreign/i_time01_03.gif">
		</td>
		
		<td class="alR" style="word-wrap:break-word;word-break:break-all;">
			<div class="wfit70">
				<font class="fl00">98.76USD</font><br>
				<font class="fl00">10,849円</font></div>
		</td>
		<td class="alR" style="word-wrap:break-word;word-break:break-all;">
			10<br>(0)
		</td>
		<td class="alR" style="word-wrap:break-word;word-break:break-all;">
			<div class="wfit70">
				<font class="fl00">102.87USD</font><br>
				<font class="fl00">11,264円</font></div>
		</td>
		<td class="alR" style="word-wrap:break-word;word-break:break-all;"><div class="wfit75">
			<font class="fl00">1,028.70USD</font><br>
			<font class="fl00">112,640円</font></div>
		</td>
		<td class="alR" style="word-wrap:break-word;word-break:break-all;"><div class="wfit75">
			<font class="fl00">987.60USD</font><br>
			<font class="fl00">108,497円</font></div>
		</td>
		<td class="alR" style="word-wrap:break-word;word-break:break-all;"><div class="wfit75">
			<font class="fl00">-41.10USD</font><br>
			<font class="fl00">-4,143円</font></div>
		</td>
		<td class="alC">
				<a href="#here" onclick="javascript:moveFbOrderEntryPrice('/czk/secCashBalance/','DIS','2','US','NYSE','1');"><img width="33" height="17" class="rollover" alt="買付" src="https://a248.e.akamai.net/f/248/29350/7d/pict.sbisec.co.jp/sbisec/images/base/foreign/i_buy_01.gif"></a>
				<p class="mgt3"></p>
				<a href="#here" onclick="javascript:moveFbOrderEntryPrice('/czk/secCashBalance/','DIS','3','US','NYSE','1');"><img width="33" height="17" class="rollover" alt="売却" src="https://a248.e.akamai.net/f/248/29350/7d/pict.sbisec.co.jp/sbisec/images/base/foreign/i_sell_01.gif"></a>
				<p class="mgt3"></p>
				<a href="#here" onclick="if(!chkDoubleTrn()){return false;};document.forms[0].action='/Fpts/rsv/reserveSettingAgreement/DIS/US/';document.forms[0].submit();"><img width="33" height="17" class="rollover" alt="定期" src="https://a248.e.akamai.net/f/248/29350/7d/pict.sbisec.co.jp/sbisec/images/base/foreign/i_regularly_01.gif"></a>
		</td>
	</tr>
	</table></body></html>`

	node, err := html.Parse(strings.NewReader(test))
	if err != nil {
		t.Fatal("can't create doc")
	}

	doc := goquery.NewDocumentFromNode(node)

	client := &sbiClient{
		browser: surf.NewBrowser(),
		cache:   NewNilFundInfoCache(),
		Logger:  zap.NewNop().Sugar(),
	}

	f, err := client.parseForegnStock(doc.Find("tr"))
	if err != nil {
		t.Fatal("can't parse stock")
	}

	if "ウォルト ディズニー" != f.Name {
		t.Errorf("Name: expected ウォルト ディズニー but got %s", f.Name)
	}

	if InternationalStocks != f.AssetClass {
		t.Errorf("AssetClass: expected %v but got %v", InternationalStocks, f.AssetClass)
	}

	if 112640 != f.AcquisitionPrice {
		t.Errorf("AcquisitionPrice: expected %d but got %v", 112640, f.AcquisitionPrice)
	}

	if 108497 != f.CurrentPrice {
		t.Errorf("CurrentPrice: expected %d but got %v", 108497, f.CurrentPrice)
	}
}
