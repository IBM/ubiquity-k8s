package plugin

import (
	"os"
	"path/filepath"

	"github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/configuration/config_helpers"
	"github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/configuration/core_config"
	"github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/consts"
	"github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/token_refresher"
	"github.ibm.com/Bluemix/bluemix-cli-sdk/plugin/models"
)

type pluginContext struct {
	coreConfig     core_config.ReadWriter
	pluginConfig   PluginConfig
	pluginPath     string
	tokenRefresher token_refresher.TokenRefresher
}

func NewPluginContext(pluginName string, coreConfig core_config.ReadWriter) *pluginContext {
	pluginPath := config_helpers.PluginDir(pluginName)
	c := &pluginContext{
		pluginPath:   pluginPath,
		pluginConfig: NewPluginConfig(filepath.Join(pluginPath, "config.json")),
	}
	c.coreConfig = coreConfig
	c.tokenRefresher = token_refresher.NewTokenRefresher(coreConfig.UaaEndpoint())
	return c
}

func (c *pluginContext) PluginDirectory() string {
	return c.pluginPath
}

func (c *pluginContext) PluginConfig() PluginConfig {
	return c.pluginConfig
}

func (c *pluginContext) AuthenticationEndpoint() string {
	return c.coreConfig.AuthenticationEndpoint()
}

func (c *pluginContext) DopplerEndpoint() string {
	return c.coreConfig.DopplerEndpoint()
}

func (c *pluginContext) LoggregatorEndpoint() string {
	return c.coreConfig.LoggregatorEndpoint()
}

func (c *pluginContext) UaaEndpoint() string {
	return c.coreConfig.UaaEndpoint()
}

func (c *pluginContext) APIEndpoint() string {
	return c.coreConfig.APIEndpoint()
}

func (c *pluginContext) APIVersion() string {
	return c.coreConfig.APIVersion()
}

func (c *pluginContext) HasAPIEndpoint() bool {
	return c.coreConfig.HasAPIEndpoint()
}

func (c *pluginContext) IsLoggedIn() bool {
	return c.coreConfig.IsLoggedIn()
}

func (c *pluginContext) UserEmail() string {
	return c.coreConfig.UserEmail()
}

func (c *pluginContext) UserGUID() string {
	return c.coreConfig.UserGUID()
}

func (c *pluginContext) Username() string {
	return c.coreConfig.Username()
}

func (c *pluginContext) HasOrganization() bool {
	return c.coreConfig.HasOrganization()
}

func (c *pluginContext) HasSpace() bool {
	return c.coreConfig.HasSpace()
}

func (c *pluginContext) AccessToken() string {
	return c.coreConfig.AccessToken()
}

func (c *pluginContext) TokenRefresh() (string, error) {
	newToken, newRefreshToken, err := c.tokenRefresher.Refresh(c.coreConfig.RefreshToken())
	if err != nil {
		return "", err
	}
	c.coreConfig.SetAccessToken(newToken)
	c.coreConfig.SetRefreshToken(newRefreshToken)
	return newToken, nil
}

func (c *pluginContext) CurrentOrg() models.Organization {
	return models.Organization{
		OrganizationFields: c.coreConfig.OrganizationFields(),
	}
}

func (c *pluginContext) CurrentSpace() models.Space {
	return models.Space{
		SpaceFields: c.coreConfig.SpaceFields(),
	}
}

func (c *pluginContext) Locale() string {
	return c.coreConfig.Locale()
}

func (c *pluginContext) IsSSLDisabled() bool {
	return c.coreConfig.IsSSLDisabled()
}

func (c *pluginContext) Trace() string {
	return getFromEnvOrConfig(consts.ENV_BLUEMIX_TRACE, c.coreConfig.Trace())
}

func (c *pluginContext) ColorEnabled() string {
	return getFromEnvOrConfig(consts.ENV_BLUEMIX_COLOR, c.coreConfig.ColorEnabled())
}

func (c *pluginContext) HTTPTimeout() int {
	return c.coreConfig.HTTPTimeout()
}

func getFromEnvOrConfig(envKey string, config string) string {
	if envVal := os.Getenv(envKey); envVal != "" {
		return envVal
	}
	return config
}

func (c *pluginContext) CommandNamespace() string {
	return os.Getenv(consts.ENV_BLUEMIX_PLUGIN_NAMESPACE)
}
