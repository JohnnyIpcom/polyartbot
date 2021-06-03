package config

type OAuth2 struct {
	Enabled      bool     `yaml:"enabled" default:"true"`
	ClientID     string   `yaml:"clientID"`
	ClientSecret string   `yaml:"clientSecret"`
	Scopes       []string `yaml:"scopes" default:"[\"all\"]"`
	TokenURL     string   `yaml:"tokenURL" default:"\\oauth2\\token"`
}

type Client struct {
	URL    string `yaml:"url"`
	OAuth2 OAuth2 `yaml:"oauth2"`
}
