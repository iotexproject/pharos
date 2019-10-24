package handler

const (
	DelegationStatusActive  DelegationStatus = "active"
	DelegationStatusPending DelegationStatus = "pending"
)

type ValidatorPage []Validator
type DelegationsPage []Delegation

type DelegationStatus string

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

type MemberDelegates struct {
	Data struct {
		BPCandidates []struct {
			ID               string   `json:"id"`
			Rand             string   `json:"rank"`
			Logo             string   `json:"logo"`
			Name             string   `json:"name"`
			Status           string   `json:"status"`
			Category         string   `json:"category"`
			ServerStatus     string   `json:"serverStatus"`
			LiveVotes        int64    `json:"liveVotes"`
			LiveVotesDelta   int64    `json:"liveVotesDelta"`
			Percent          string   `json:"percent"`
			RegisteredName   string   `json:"registeredName"`
			SocialMedia      []string `json:"socialMedia"`
			Productivity     int      `json:"productivity"`
			ProductivityBase int      `json:"productivityBase"`
		} `json:"bpCandidates"`
	} `json:"data"`
}
