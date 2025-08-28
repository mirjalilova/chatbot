package entity

type Restriction struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	RequestLimit   *int   `json:"request_limit"`
	CharacterLimit *int   `json:"character_limit,omitempty"`
	ChatLimit      *int   `json:"chat_limit,omitempty"`
	// TimeLimit      *int   `json:"time_limit,omitempty"`
}

type UpdateRestrictionBody struct {
	RequestLimit   *int `json:"request_limit" binding:"required"`
	CharacterLimit *int `json:"character_limit,omitempty"`
	ChatLimit      *int `json:"chat_limit,omitempty"`
	// TimeLimit      *int `json:"time_limit,omitempty"`
}

type UpdateRestriction struct {
	ID             string `json:"-"`
	RequestLimit   *int   `json:"request_limit,omitempty"`
	CharacterLimit *int   `json:"character_limit,omitempty"`
	ChatLimit      *int   `json:"chat_limit,omitempty"`
	// TimeLimit      *int   `json:"time_limit,omitempty"`
}

type ListRestriction struct {
	Restrictions []Restriction `json:"restrictions"`
}
