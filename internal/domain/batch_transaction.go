package domain

type BatchCreditItem struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

type BatchDebitItem struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

type BatchTransactionItemResult struct {
	Index         int64        `json:"index"`
	Transaction   *Transaction `json:"transaction,omitempty"`
	Error         string       `json:"error,omitempty"`
	WasSuccessful bool         `json:"was_successful"`
}

type BatchTransactionResult struct {
	TotalCount   int64                        `json:"total_count"`
	SuccessCount int64                        `json:"success_count"`
	FailureCount int64                        `json:"failure_count"`
	Items        []BatchTransactionItemResult `json:"items"`
}
