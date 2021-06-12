package services

import (
	"net/http"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/johnnyipcom/polyartbot/internal/cdn/config"
	"go.uber.org/zap"
)

type OAuth2Service interface {
	Enabled() bool
	HandleAuthorizeRequest(w http.ResponseWriter, r *http.Request) error
	HandleTokenRequest(w http.ResponseWriter, r *http.Request) error
	ValidationBearerToken(r *http.Request) (oauth2.TokenInfo, error)
}

type oauth2Service struct {
	cfg    config.Server
	server *server.Server
}

func NewOAuth2Service(cfg config.Config, log *zap.Logger) OAuth2Service {
	manager := manage.NewDefaultManager()
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	clientStore := store.NewClientStore()
	for _, clientCfg := range cfg.Server.OAuth2.Clients {
		clientStore.Set(clientCfg.ID, &models.Client{
			ID:     clientCfg.ID,
			Secret: clientCfg.Secret,
			Domain: clientCfg.Domain,
		})
	}
	manager.MapClientStorage(clientStore)

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)

	return &oauth2Service{
		cfg:    cfg.Server,
		server: server.NewDefaultServer(manager),
	}
}

func (o oauth2Service) Enabled() bool {
	return o.cfg.OAuth2.Enabled
}

func (o *oauth2Service) HandleAuthorizeRequest(w http.ResponseWriter, r *http.Request) error {
	return o.server.HandleAuthorizeRequest(w, r)
}

func (o *oauth2Service) HandleTokenRequest(w http.ResponseWriter, r *http.Request) error {
	return o.server.HandleTokenRequest(w, r)
}

func (o *oauth2Service) ValidationBearerToken(r *http.Request) (oauth2.TokenInfo, error) {
	return o.server.ValidationBearerToken(r)
}
