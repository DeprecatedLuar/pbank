package commands

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
)

// API configuration
const (
	coinGeckoAPIURL   = "https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=%s"
	awesomeAPIURL     = "https://economia.awesomeapi.com.br/json/last/%s"
	frankfurterAPIURL = "https://api.frankfurter.app/latest?from=%s&to=%s"
	httpStatusOK      = 200
)

// Display formatting
const (
	valueTabPad     = 2
	decimalFormat   = "%.2f"
	valueNotAvail   = "N/A"
)

var cryptoIDMap = map[string]string{
	"BTC":  "bitcoin",
	"ETH":  "ethereum",
	"BNB":  "binancecoin",
	"SOL":  "solana",
	"XRP":  "ripple",
	"ADA":  "cardano",
	"AVAX": "avalanche-2",
	"DOGE": "dogecoin",
	"DOT":  "polkadot",
	"MATIC": "matic-network",
	"LINK": "chainlink",
	"UNI":  "uniswap",
	"ATOM": "cosmos",
	"LTC":  "litecoin",
	"BCH":  "bitcoin-cash",
	"XLM":  "stellar",
	"ALGO": "algorand",
	"VET":  "vechain",
	"ICP":  "internet-computer",
	"FIL":  "filecoin",
	"HBAR": "hedera-hashgraph",
	"NEAR": "near",
	"APT":  "aptos",
	"QNT":  "quant-network",
	"ARB":  "arbitrum",
	"OP":   "optimism",
	"INJ":  "injective-protocol",
	"RUNE": "thorchain",
	"AAVE": "aave",
	"GRT":  "the-graph",
	"MKR":  "maker",
	"SNX":  "synthetix-network-token",
	"ENJ":  "enjincoin",
	"ETHW": "ethereum-pow-iou",
	"MBOX": "mobox",
}

func HandleNetworth(db *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("currency required. Usage: pbank networth <CURRENCY>\nExample: pbank networth BRL")
	}

	targetCurrency := strings.ToUpper(args[0])
	rows, err := db.Query(`
		SELECT f.label, fb.currency, fb.amount
		FROM fund_balances fb
		JOIN funds f ON fb.fund_id = f.id
		WHERE fb.amount != 0
		ORDER BY f.label, fb.currency
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type Balance struct {
		Fund     string
		Currency string
		Amount   float64
	}

	var balances []Balance
	cryptoCurrencies := make(map[string]bool)
	fiatCurrencies := make(map[string]bool)

	for rows.Next() {
		var b Balance
		if err := rows.Scan(&b.Fund, &b.Currency, &b.Amount); err != nil {
			return err
		}
		balances = append(balances, b)

		if _, isCrypto := cryptoIDMap[b.Currency]; isCrypto {
			cryptoCurrencies[b.Currency] = true
		} else if b.Currency != targetCurrency {
			fiatCurrencies[b.Currency] = true
		}
	}

	if len(balances) == 0 {
		fmt.Println("No balances found")
		return nil
	}

	cryptoRates := make(map[string]float64)
	if len(cryptoCurrencies) > 0 {
		var err error
		cryptoRates, err = fetchCryptoRates(cryptoCurrencies, targetCurrency)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: crypto rates fetch failed: %v\n", err)
		}
	}

	fiatRates := make(map[string]float64)
	usedFallback := false
	if len(fiatCurrencies) > 0 {
		var err error
		fiatRates, err = fetchFiatRates(fiatCurrencies, targetCurrency)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: awesomeapi failed (%v), trying fallback...\n", err)
			fiatRates, err = fetchFiatRatesFrankfurter(fiatCurrencies, targetCurrency)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: frankfurter fallback also failed: %v\n", err)
			} else {
				usedFallback = true
			}
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, valueTabPad, ' ', 0)
	fmt.Fprintf(w, "Fund\tCurrency\tAmount\t%s Value\n", targetCurrency)
	fmt.Fprintln(w, "----\t--------\t------\t---------")

	total := 0.0
	for _, b := range balances {
		var convertedValue string
		var numericValue float64

		if b.Currency == targetCurrency {
			convertedValue = fmt.Sprintf(decimalFormat, b.Amount)
			numericValue = b.Amount
		} else if rate, ok := cryptoRates[b.Currency]; ok {
			numericValue = b.Amount * rate
			convertedValue = fmt.Sprintf(decimalFormat, numericValue)
		} else if rate, ok := fiatRates[b.Currency]; ok {
			numericValue = b.Amount * rate
			convertedValue = fmt.Sprintf(decimalFormat, numericValue)
		} else {
			convertedValue = valueNotAvail
		}

		if convertedValue != valueNotAvail {
			total += numericValue
		}

		fmt.Fprintf(w, "%s\t%s\t%v\t%s\n", b.Fund, b.Currency, b.Amount, convertedValue)
	}

	fmt.Fprintln(w, "----\t--------\t------\t---------")
	totalPrefix := ""
	if usedFallback {
		totalPrefix = "~"
	}
	fmt.Fprintf(w, "Total\t\t\t"+totalPrefix+decimalFormat+" %s\n", total, targetCurrency)
	w.Flush()

	return nil
}

func fetchCryptoRates(currencies map[string]bool, targetCurrency string) (map[string]float64, error) {
	ids := make([]string, 0, len(currencies))
	for ticker := range currencies {
		if id, ok := cryptoIDMap[ticker]; ok {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		return nil, nil
	}

	url := fmt.Sprintf(coinGeckoAPIURL, strings.Join(ids, ","), strings.ToLower(targetCurrency))

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != httpStatusOK {
		return nil, fmt.Errorf("coingecko returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]map[string]float64
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	targetCurrencyLower := strings.ToLower(targetCurrency)
	rates := make(map[string]float64)
	for ticker, coinID := range cryptoIDMap {
		if currencies[ticker] {
			if priceData, ok := data[coinID]; ok {
				if rate, ok := priceData[targetCurrencyLower]; ok {
					rates[ticker] = rate
				}
			}
		}
	}

	return rates, nil
}

func fetchFiatRates(currencies map[string]bool, targetCurrency string) (map[string]float64, error) {
	pairs := make([]string, 0, len(currencies))
	for curr := range currencies {
		pairs = append(pairs, curr+"-"+targetCurrency)
	}

	url := fmt.Sprintf(awesomeAPIURL, strings.Join(pairs, ","))

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != httpStatusOK {
		return nil, fmt.Errorf("awesomeapi returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]struct {
		Bid string `json:"bid"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	rates := make(map[string]float64)
	for curr := range currencies {
		key := strings.ReplaceAll(curr+targetCurrency, "-", "")
		if rateData, ok := data[key]; ok {
			var rate float64
			if _, err := fmt.Sscanf(rateData.Bid, "%f", &rate); err == nil {
				rates[curr] = rate
			}
		}
	}

	return rates, nil
}

func fetchFiatRatesFrankfurter(currencies map[string]bool, targetCurrency string) (map[string]float64, error) {
	rates := make(map[string]float64)

	for curr := range currencies {
		url := fmt.Sprintf(frankfurterAPIURL, curr, targetCurrency)

		resp, err := http.Get(url)
		if err != nil {
			continue
		}

		if resp.StatusCode != httpStatusOK {
			resp.Body.Close()
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		var data struct {
			Rates map[string]float64 `json:"rates"`
		}
		if err := json.Unmarshal(body, &data); err != nil {
			continue
		}

		if rate, ok := data.Rates[targetCurrency]; ok {
			rates[curr] = rate
		}
	}

	if len(rates) == 0 {
		return nil, fmt.Errorf("no rates fetched")
	}

	return rates, nil
}
