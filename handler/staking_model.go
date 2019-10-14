package handler

type ValidatorPage []Validator
type DelegationsPage []Delegation

type DelegationStatus string

const (
	DelegationStatusActive  DelegationStatus = "active"
	DelegationStatusPending DelegationStatus = "pending"
)

type StakingReward struct {
	Annual float64 `json:"annual"`
}

type StakingDetails struct {
	Reward        StakingReward `json:"reward"`
	LockTime      int           `json:"locktime"`
	MinimumAmount string        `json:"minimum_amount"`
}

type Validator struct {
	ID      string         `json:"id"`
	Status  bool           `json:"status"`
	Details StakingDetails `json:"details"`
}

type Delegation struct {
	Delegator StakeValidator   `json:"delegator"`
	Value     string           `json:"value"`
	Status    DelegationStatus `json:"status"`
	Metadata  interface{}      `json:"metadata,omitempty"`
}

type StakeValidatorInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Website     string `json:"website"`
}

type StakeValidator struct {
	ID      string             `json:"id"`
	Status  bool               `json:"status,omitempty"`
	Info    StakeValidatorInfo `json:"info,omitempty"`
	Details StakingDetails     `json:"details,omitempty"`
}
