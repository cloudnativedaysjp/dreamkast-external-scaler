package dreamkast

type ListConferencesResp []GetConferenceResp

type GetConferenceResp struct {
	ID                      int             `json:"id"`
	Name                    string          `json:"name"`
	Abbr                    string          `json:"abbr"`
	Status                  string          `json:"status"`
	Theme                   string          `json:"theme"`
	About                   string          `json:"about"`
	PrivacyPolicy           string          `json:"privacy_policy"`
	PrivacyPolicyForSpeaker string          `json:"privacy_policy_for_speaker"`
	Copyright               string          `json:"copyright"`
	Coc                     string          `json:"coc"`
	ConferenceDays          []ConferenceDay `json:"conferenceDays"`
}

type ConferenceDay struct {
	ID       int    `json:"id"`
	Date     string `json:"date"`
	Internal bool   `json:"internal"`
}
