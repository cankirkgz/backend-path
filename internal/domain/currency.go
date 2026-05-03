package domain

type Currency string

const (
	CurrencyTRY Currency = "TRY"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)

func (c Currency) IsValid() bool {
	switch c {
	case CurrencyTRY, CurrencyUSD, CurrencyEUR:
		return true
	default:
		return false
	}
}
