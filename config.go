package slack

type Config struct {
	Ok       bool
	Error    string
	Self     User
	Channels []Channel
	Url      string
	Users    []User
}

type User struct {
	Id             string
	Name           string
	RealName       string `json:"real_name"`
	Deleted        bool
	IsAdmin        bool `json:"is_admin"`
	IsOwner        bool `json:"is_owner"`
	IsPrimaryOwner bool `json:"is_primary_owner"`
	IsBot          bool `json:"is_bot"`
	Profile        UserProfile
}

type Channel struct {
	Id         string
	Name       string
	IsChannel  bool `json:"is_channel"`
	IsIm       bool `json:"is_im"`
	User       string
	Created    int
	Creator    string
	IsArchived bool `json:"is_archived"`
	IsGeneral  bool `json:"is_general"`
	IsMember   bool `json:"is_member"`
	Members    []string
}

type UserProfile struct {
	FirstName          string `json:"first_name"`
	LastName           string `json:"last_name"`
	RealName           string `json:"real_name"`
	Title              string
	RealNameNormalized string `json:"real_name_normalized"`
	Email              string `json:"email"`
}
