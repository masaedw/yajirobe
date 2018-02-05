package yajirobe

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf/agent"
	"github.com/headzoo/surf/browser"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/headzoo/surf.v1"
)

type sbiClient struct {
	browser *browser.Browser
	cache   FundInfoCache
	Logger  *zap.Logger
}

// SbiOption NewSbiScannerの引数
type SbiOption struct {
	UserID   string
	Password string
	Cache    FundInfoCache
	Logger   *zap.Logger
}

// NewSbiScanner SBI証券用Scannerを作る
func NewSbiScanner(option SbiOption) (Scanner, error) {
	if option.Logger == nil {
		option.Logger = zap.NewNop()
	}

	client := &sbiClient{
		browser: surf.NewBrowser(),
		cache:   option.Cache,
		Logger:  option.Logger,
	}

	if err := client.login(option.UserID, option.Password); err != nil {
		return nil, errors.Wrap(err, "can't login")
	}
	client.Logger.Sugar().Debugf("sbi: login")

	return client, nil
}

func (c *sbiClient) login(userID, password string) error {
	bow := c.browser
	bow.SetUserAgent(agent.Chrome())

	if err := bow.Open("https://www.sbisec.co.jp/ETGate"); err != nil {
		return errors.Wrap(err, "SBI: Can't open sbi top page")
	}
	c.Logger.Debug("sbi: open top-page")

	loginForm, err := bow.Form("[name='form_login']")
	if err != nil {
		return errors.Wrap(err, "SBI: Can't detect login form")
	}
	c.Logger.Debug("sbi: detect login page")

	err = setForms(loginForm, map[string]string{
		"JS_FLG":          "0",
		"BW_FLG":          "0",
		"_ControlID":      "WPLETlgR001Control",
		"_DataStoreID":    "DSWPLETlgR001Control",
		"_PageID":         "WPLETlgR001Rlgn20",
		"_ActionID":       "login",
		"getFlg":          "on",
		"allPrmFlg":       "on",
		"_ReturnPageInfo": "WPLEThmR001Control/DefaultPID/DefaultAID/DSWPLEThmR001Control",
		"user_id":         userID,
		"user_password":   password,
	})
	if err != nil {
		return errors.Wrap(err, "SBI: Can't set login form")
	}

	if err := loginForm.Submit(); err != nil {
		return errors.Wrap(err, "SBI: Can't submit login form")
	}
	c.Logger.Debug("sbi: submit login")

	text := toUtf8(bow.Find("font").Text())
	if strings.Contains(text, "WBLE") {
		// ログイン失敗画面
		return errors.Errorf("SBI: login failed: %s", text)
	}

	nextForm := bow.Find("form").First()
	if nextForm == nil {
		return errors.Errorf("SBI: formSwitch not found")
	}
	c.Logger.Debug("sbi: detect formswitch")

	// 2回目のPOST
	if err := bow.PostForm(nextForm.AttrOr("action", "url not found"), exportValues(nextForm)); err != nil {
		return errors.Wrap(err, "SBI: formSwitch post failed")
	}
	c.Logger.Debug("sbi: post formswitch")

	if !strings.Contains(toUtf8(bow.Body()), "最終ログイン:") {
		// ログイン成功時のメッセージが出てなければログイン失敗してる
		c.Logger.Debug("Can't detect login message")
		return errors.New("SBI: the SBI User ID or Password failed")
	}
	c.Logger.Debug("sbi: succeeded login")

	return nil
}

func (c *sbiClient) accountPage() error {
	bow := c.browser
	s := filterAttrContains(bow.Dom().Find("img"), "alt", toSjis("口座管理"))
	if s == nil || s.Length() == 0 {
		return errors.New("SBI: Can't find 口座管理")
	}
	url := s.Parent().AttrOr("href", "can't find url")
	url, _ = bow.ResolveStringUrl(url)
	e := bow.Open(url)
	if e != nil {
		return errors.Wrap(e, "SBI: Can't open 口座管理")
	}

	s = filterAttrContains(bow.Dom().Find("area"), "alt", toSjis("保有証券"))
	url, _ = bow.ResolveStringUrl(s.AttrOr("href", "can't find url"))
	e = bow.Open(url)
	if e != nil {
		return errors.Wrap(e, "SBI: Can't open 保有証券")
	}

	return nil
}

func (c *sbiClient) scanStock(row *goquery.Selection) *Stock {
	cells := iterate(row.Find("td"))

	nameCode := toUtf8(cells[0].Text())
	re := regexp.MustCompile(`(\S+?)(\d{4})`)
	m := re.FindStringSubmatch(nameCode)

	name := m[1]
	code, _ := strconv.Atoi(m[2])
	amount := int(parseSeparatedInt(toUtf8(cells[1].Text())))
	units := iterateText(cells[2])
	acquisitionUnitPrice := parseSeparatedInt(units[0])
	currentUnitPrice := parseSeparatedInt(units[1])
	acquisitionPrice := acquisitionUnitPrice * int64(amount)
	currentPrice := currentUnitPrice * int64(amount)

	return &Stock{
		Name:                 name,
		Code:                 code,
		Amount:               amount,
		AcquisitionUnitPrice: acquisitionUnitPrice,
		CurrentUnitPrice:     currentUnitPrice,
		AcquisitionPrice:     acquisitionPrice,
		CurrentPrice:         currentPrice,
	}
}

func (c *sbiClient) getFundInfo(code FundCode) (*FundInfo, error) {
	bow := c.browser
	url, _ := url.Parse("https://site0.sbisec.co.jp/marble/fund/detail/achievement.do")
	query := url.Query()
	query.Set("Param6", string(code))
	url.RawQuery = query.Encode()
	if err := bow.Open(url.String()); err != nil {
		return nil, errors.Wrapf(err, "SBI: Can't open fund's page of %v", code)
	}
	categoryHeader := filterTextContains(bow.Find("tr th div p"), toSjis("商品分類"))
	category := categoryHeader.Parent().Parent().Parent().First().Next()
	nameHeader := strings.TrimSpace(toUtf8(bow.Find("h3").Text()))
	names := strings.Split(nameHeader, "－")
	name := names[0]
	if len(names) > 1 {
		name = names[1]
	}
	assetClass := parseAssetClass(strings.TrimSpace(toUtf8(category.Text())))

	return &FundInfo{
		Name:  name,
		Class: assetClass,
		Code:  code,
	}, nil
}

func (c *sbiClient) scanFund(row *goquery.Selection) (*Fund, error) {
	cells := iterate(row.Find("td"))

	href := cells[0].Find("a").AttrOr("href", "Fund name not found")
	url, _ := url.Parse(href)
	query := url.Query()
	code := FundCode(query.Get("fund_sec_code"))
	amount := parseSeparatedInt(cells[1].Text())
	units := iterateText(cells[2])
	acquisitionUnitPrice := parseSeparatedInt(units[0])
	currentUnitPrice := parseSeparatedInt(units[1])

	fi, err := c.cache.GetOrFind(code, FundInfoFinder(func(code FundCode) (*FundInfo, error) {
		return c.getFundInfo(code)
	}))

	if err != nil {
		return nil, errors.WithStack(err)
	}

	name := fi.Name
	class := fi.Class

	return &Fund{
		Name:                 name,
		Code:                 code,
		AssetClass:           class,
		Amount:               int(amount),
		AcquisitionUnitPrice: float64(acquisitionUnitPrice),
		CurrentUnitPrice:     float64(currentUnitPrice),
		AcquisitionPrice:     float64(acquisitionUnitPrice) * float64(amount) / 10000,
		CurrentPrice:         float64(currentUnitPrice) * float64(amount) / 10000,
	}, nil
}

func (c *sbiClient) Scan() ([]*Stock, []*Fund, error) {
	if e := c.accountPage(); e != nil {
		return nil, nil, e
	}

	bow := c.browser

	stockFont := filterTextContains(bow.Find("font"), toSjis("銘柄"))
	stockTable := stockFont.ParentsFiltered("table").First()

	stocks := []*Stock{}

	for _, tr := range iterate(stockTable.Find("tr"))[1:] {
		stocks = append(stocks, c.scanStock(tr))
	}

	fundFont := filterTextContains(bow.Find("font"), toSjis("ファンド名"))
	fundTables := iterate(fundFont.Parent().Parent().Parent().Parent())

	funds := []*Fund{}

	for _, table := range fundTables {
		for i, tr := range iterate(table.Find("tr")) {
			if i%2 == 0 {
				continue
			}
			f, e := c.scanFund(tr)
			if e != nil {
				return nil, nil, e
			}
			funds = append(funds, f)
		}
	}

	return stocks, funds, nil
}
