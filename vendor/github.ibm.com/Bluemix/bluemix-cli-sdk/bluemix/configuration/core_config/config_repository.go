package core_config

// internal use only

import (
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/blang/semver"

	cfconfiguration "github.com/cloudfoundry/cli/cf/configuration"
	cfconfighelpers "github.com/cloudfoundry/cli/cf/configuration/confighelpers"
	cfcoreconfig "github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	cfmodels "github.com/cloudfoundry/cli/cf/models"

	"github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/configuration"
	"github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/configuration/config_helpers"
	"github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/models"
)

const (
	_DEFAULT_CLI_INFO_ENDPOINT = "https://clis.ng.bluemix.net/info"
	_DEFAULT_CLI_DOWNLOAD_PAGE = "https://clis.ng.bluemix.net"

	_DEFAULT_PLUGIN_REPO_NAME = "Bluemix"
	_DEFAULT_PLUGIN_REPO_URL  = "https://plugins.ng.bluemix.net"
)

type BluemixConfigData struct {
	CLIInfoEndpoint          string
	CLIDownloadPage          string
	CLILatestVersion         string
	MinCLIVersion            string
	MinRecommendedCLIVersion string
	PluginRepos              []models.PluginRepo
	Locale                   string
	Trace                    string
	ColorEnabled             string
	HTTPTimeout              int
	CheckCLIVersionDisabled  bool
	UsageStatsDisabled       bool
	Region                   string
}

type Reader interface {
	// cf
	APIVersion() string
	APIEndpoint() string
	HasAPIEndpoint() bool
	Domain() string

	AuthenticationEndpoint() string
	LoggregatorEndpoint() string
	DopplerEndpoint() string
	UaaEndpoint() string
	RoutingAPIEndpoint() string
	AccessToken() string
	SSHOAuthClient() string
	RefreshToken() string

	OrganizationFields() models.OrganizationFields
	HasOrganization() bool

	SpaceFields() models.SpaceFields
	HasSpace() bool

	Username() string
	UserGUID() string
	UserEmail() string
	IsLoggedIn() bool

	IsSSLDisabled() bool

	// bx
	IsMinCLIVersion(string) bool
	MinCLIVersion() string
	MinRecommendedCLIVersion() string

	CLIInfoEndpoint() string
	CLIDownloadPage() string
	CLILatestVersion() string

	PluginRepos() []models.PluginRepo
	PluginRepo(string) (models.PluginRepo, bool)

	HTTPTimeout() int
	Trace() string

	ColorEnabled() string

	Locale() string

	CheckCLIVersionDisabled() bool

	UsageStatsDisabled() bool

	Region() string
}

type ReadWriter interface {
	Reader

	ReloadCFConfig()
	SetAPIVersion(string)
	SetAPIEndpoint(string)
	SetMinRecommendedCLIVersion(string)
	SetAuthenticationEndpoint(string)
	SetLoggregatorEndpoint(string)
	SetDopplerEndpoint(string)
	SetUaaEndpoint(string)
	SetRoutingAPIEndpoint(string)
	SetSSHOauthClient(string)
	SetAccessToken(string)
	SetRefreshToken(string)
	SetOrganizationFields(models.OrganizationFields)
	SetSpaceFields(models.SpaceFields)
	SetSSLDisabled(bool)

	SetCLIInfoEndpoint(string)
	SetCLIDownloadPage(string)
	SetCLILatestVersion(string)
	SetMinCLIVersion(string)
	SetPluginRepo(models.PluginRepo)
	UnSetPluginRepo(string)
	SetLocale(string)
	SetTrace(string)
	SetColorEnabled(string)
	SetHTTPTimeout(int)
	SetCheckCLIVersionDisabled(bool)
	SetUsageStatsDisabled(bool)
	SetRegion(string)
}

type configRepository struct {
	bxConfigData *BluemixConfigData
	bxPersistor  configuration.Persistor
	initOnce     *sync.Once
	lock         sync.RWMutex
	onError      func(error)

	cfConfig    cfcoreconfig.Repository
	cfPersistor cfconfiguration.Persistor
}

func NewCoreConfig(errHandler func(error)) ReadWriter {
	cfConfigPath, err := cfconfighelpers.DefaultFilePath()
	if err != nil {
		errHandler(err)
	}
	return NewCoreConfigFromPath(cfConfigPath, config_helpers.ConfigFilePath(), errHandler)
}

func NewCoreConfigFromPath(cfConfigPath string, bxConfigPath string, errHandler func(error)) ReadWriter {
	cfPersistor := cfconfiguration.NewDiskPersistor(cfConfigPath)
	bxPersistor := configuration.NewDiskPersistor(bxConfigPath)
	return NewCoreConfigFromPersistor(cfPersistor, bxPersistor, errHandler)
}

func NewCoreConfigFromPersistor(cfPersistor cfconfiguration.Persistor, bxPersistor configuration.Persistor, errHandler func(error)) ReadWriter {
	return &configRepository{
		bxConfigData: &BluemixConfigData{},
		bxPersistor:  bxPersistor,
		initOnce:     new(sync.Once),
		onError:      errHandler,

		cfConfig:    cfcoreconfig.NewRepositoryFromPersistor(cfPersistor, errHandler),
		cfPersistor: cfPersistor,
	}
}

func (c *configRepository) initBXConfig() {
	c.initOnce.Do(func() {
		err := c.bxPersistor.Load(c.bxConfigData)

		if err != nil && os.IsNotExist(err) {
			c.bxConfigData = bxConfigDefaults()
			err = c.bxPersistor.Save(c.bxConfigData)
		}

		if err != nil {
			c.onError(err)
		}
	})
}

func bxConfigDefaults() *BluemixConfigData {
	return &BluemixConfigData{
		PluginRepos: []models.PluginRepo{
			{
				Name: _DEFAULT_PLUGIN_REPO_NAME,
				URL:  _DEFAULT_PLUGIN_REPO_URL,
			},
		},
	}
}

func (repo *configRepository) read(cb func()) {
	repo.lock.RLock()
	defer repo.lock.RUnlock()

	repo.initBXConfig()

	cb()
}

func (repo *configRepository) write(cb func()) {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	repo.initBXConfig()

	cb()

	err := repo.bxPersistor.Save(repo.bxConfigData)
	if err != nil {
		repo.onError(err)
	}
}

func (repo *configRepository) CLIInfoEndpoint() (endpoint string) {
	repo.read(func() {
		endpoint = repo.bxConfigData.CLIInfoEndpoint
	})
	if endpoint != "" {
		return endpoint
	}
	return _DEFAULT_CLI_INFO_ENDPOINT
}

func (repo *configRepository) CLIDownloadPage() (url string) {
	repo.read(func() {
		url = repo.bxConfigData.CLIDownloadPage
	})
	if url != "" {
		return url
	}
	return _DEFAULT_CLI_DOWNLOAD_PAGE
}

func (repo *configRepository) CLILatestVersion() (version string) {
	repo.read(func() {
		version = repo.bxConfigData.CLILatestVersion
	})
	return
}

func (repo *configRepository) MinCLIVersion() (minCLIVersion string) {
	repo.read(func() {
		minCLIVersion = repo.bxConfigData.MinCLIVersion
	})
	return
}

func (repo *configRepository) IsMinCLIVersion(version string) bool {
	minCLIVersion := repo.MinCLIVersion()
	if minCLIVersion == "" {
		return true
	}

	actualVersion, err := semver.Make(version)
	if err != nil {
		return false
	}

	minVersion, err := semver.Make(minCLIVersion)
	if err != nil {
		return false
	}

	return actualVersion.GTE(minVersion)
}

func (repo *configRepository) MinRecommendedCLIVersion() (minRecommendedCLIVersion string) {
	repo.read(func() {
		minRecommendedCLIVersion = repo.bxConfigData.MinRecommendedCLIVersion
	})
	return
}

func (repo *configRepository) PluginRepos() (repos []models.PluginRepo) {
	repo.read(func() {
		repos = repo.bxConfigData.PluginRepos
	})
	return
}

func (repo *configRepository) PluginRepo(name string) (models.PluginRepo, bool) {
	for _, r := range repo.PluginRepos() {
		if strings.EqualFold(r.Name, name) {
			return r, true
		}
	}
	return models.PluginRepo{}, false
}

func (repo *configRepository) Locale() (locale string) {
	repo.read(func() {
		locale = repo.bxConfigData.Locale
	})
	return
}

func (repo *configRepository) Trace() (trace string) {
	repo.read(func() {
		trace = repo.bxConfigData.Trace
	})
	return
}

func (repo *configRepository) ColorEnabled() (colorEnabled string) {
	repo.read(func() {
		colorEnabled = repo.bxConfigData.ColorEnabled
	})
	return
}

func (repo *configRepository) HTTPTimeout() (timeout int) {
	repo.read(func() {
		timeout = repo.bxConfigData.HTTPTimeout
	})
	return
}

func (repo *configRepository) CheckCLIVersionDisabled() (disabled bool) {
	repo.read(func() {
		disabled = repo.bxConfigData.CheckCLIVersionDisabled
	})
	return
}

func (repo *configRepository) UsageStatsDisabled() (disabled bool) {
	repo.read(func() {
		disabled = repo.bxConfigData.UsageStatsDisabled
	})
	return
}

func (repo *configRepository) Region() (region string) {
	repo.read(func() {
		region = repo.bxConfigData.Region
	})
	return
}

func (repo *configRepository) SetCLIInfoEndpoint(endpoint string) {
	repo.write(func() {
		repo.bxConfigData.CLIInfoEndpoint = endpoint
	})
}

func (repo *configRepository) SetCLIDownloadPage(url string) {
	repo.write(func() {
		repo.bxConfigData.CLIDownloadPage = url
	})
}

func (repo *configRepository) SetCLILatestVersion(version string) {
	repo.write(func() {
		repo.bxConfigData.CLILatestVersion = version
	})
}

func (repo *configRepository) SetMinCLIVersion(minCLIVersion string) {
	repo.write(func() {
		repo.bxConfigData.MinCLIVersion = minCLIVersion
	})
}

func (repo *configRepository) SetMinRecommendedCLIVersion(minRecommendedCLIVersion string) {
	repo.write(func() {
		repo.bxConfigData.MinRecommendedCLIVersion = minRecommendedCLIVersion
	})
}

func (repo *configRepository) SetPluginRepo(pluginRepo models.PluginRepo) {
	repo.write(func() {
		repo.bxConfigData.PluginRepos = append(repo.bxConfigData.PluginRepos, pluginRepo)
	})
}

func (repo *configRepository) UnSetPluginRepo(repoName string) {
	repo.write(func() {
		i := 0
		for ; i < len(repo.bxConfigData.PluginRepos); i++ {
			if strings.ToLower(repo.bxConfigData.PluginRepos[i].Name) == strings.ToLower(repoName) {
				break
			}
		}
		if i != len(repo.bxConfigData.PluginRepos) {
			repo.bxConfigData.PluginRepos = append(repo.bxConfigData.PluginRepos[:i], repo.bxConfigData.PluginRepos[i+1:]...)
		}
	})
}

func (repo *configRepository) SetLocale(locale string) {
	repo.write(func() {
		repo.bxConfigData.Locale = locale
	})
}

func (repo *configRepository) SetTrace(trace string) {
	repo.write(func() {
		repo.bxConfigData.Trace = trace
	})
}

func (repo *configRepository) SetColorEnabled(colorEnabled string) {
	repo.write(func() {
		repo.bxConfigData.ColorEnabled = colorEnabled
	})
}

func (repo *configRepository) SetHTTPTimeout(timeout int) {
	repo.write(func() {
		repo.bxConfigData.HTTPTimeout = timeout
	})
}

func (repo *configRepository) SetCheckCLIVersionDisabled(disabled bool) {
	repo.write(func() {
		repo.bxConfigData.CheckCLIVersionDisabled = disabled
	})
}

func (repo *configRepository) SetUsageStatsDisabled(disabled bool) {
	repo.write(func() {
		repo.bxConfigData.UsageStatsDisabled = disabled
	})
}

func (repo *configRepository) SetRegion(region string) {
	repo.write(func() {
		repo.bxConfigData.Region = region
	})
}

func (repo *configRepository) cfconfig() cfcoreconfig.Repository {
	repo.lock.RLock()
	defer repo.lock.RUnlock()

	return repo.cfConfig
}

func (repo *configRepository) ReloadCFConfig() {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	repo.cfConfig = cfcoreconfig.NewRepositoryFromPersistor(repo.cfPersistor, repo.onError)
}

func (repo *configRepository) APIVersion() string {
	return repo.cfconfig().APIVersion()
}

func (repo *configRepository) APIEndpoint() string {
	return repo.cfconfig().APIEndpoint()
}

func (repo *configRepository) HasAPIEndpoint() bool {
	return repo.cfconfig().HasAPIEndpoint()
}

func (repo *configRepository) Domain() string {
	endpoint := repo.APIEndpoint()
	if endpoint == "" {
		return ""
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}

	host := u.Host
	idx := strings.Index(host, ".")
	if idx == -1 || idx == len(host)-1 {
		return host
	}
	return string(host[idx+1:])
}

func (repo *configRepository) AuthenticationEndpoint() string {
	return repo.cfconfig().AuthenticationEndpoint()
}

func (repo *configRepository) LoggregatorEndpoint() string {
	return repo.cfconfig().LoggregatorEndpoint()
}

func (repo *configRepository) DopplerEndpoint() string {
	return repo.cfconfig().DopplerEndpoint()
}

func (repo *configRepository) UaaEndpoint() string {
	return repo.cfconfig().UaaEndpoint()
}

func (repo *configRepository) RoutingAPIEndpoint() string {
	return repo.cfconfig().RoutingAPIEndpoint()
}

func (repo *configRepository) SSHOAuthClient() string {
	return repo.cfconfig().SSHOAuthClient()
}

func (repo *configRepository) AccessToken() string {
	return repo.cfconfig().AccessToken()
}

func (repo *configRepository) RefreshToken() string {
	return repo.cfconfig().RefreshToken()
}

func (repo *configRepository) OrganizationFields() models.OrganizationFields {
	cforg := repo.cfconfig().OrganizationFields()

	var org models.OrganizationFields
	org.Name = cforg.Name
	org.GUID = cforg.GUID
	org.QuotaDefinition.GUID = cforg.QuotaDefinition.GUID
	org.QuotaDefinition.Name = cforg.QuotaDefinition.Name
	org.QuotaDefinition.MemoryLimitInMB = cforg.QuotaDefinition.MemoryLimit
	org.QuotaDefinition.InstanceMemoryLimitInMB = cforg.QuotaDefinition.InstanceMemoryLimit
	org.QuotaDefinition.RoutesLimit = cforg.QuotaDefinition.RoutesLimit
	org.QuotaDefinition.ServicesLimit = cforg.QuotaDefinition.ServicesLimit
	org.QuotaDefinition.NonBasicServicesAllowed = cforg.QuotaDefinition.NonBasicServicesAllowed
	org.QuotaDefinition.AppInstanceLimit = cforg.QuotaDefinition.AppInstanceLimit
	return org
}

func (repo *configRepository) HasOrganization() bool {
	return repo.cfconfig().HasOrganization()
}

func (repo *configRepository) SpaceFields() models.SpaceFields {
	cfspace := repo.cfconfig().SpaceFields()

	var space models.SpaceFields
	space.Name = cfspace.Name
	space.GUID = cfspace.GUID
	space.AllowSSH = cfspace.AllowSSH
	return space
}

func (repo *configRepository) HasSpace() bool {
	return repo.cfconfig().HasSpace()
}

func (repo *configRepository) Username() string {
	return repo.cfconfig().Username()
}

func (repo *configRepository) UserGUID() string {
	return repo.cfconfig().UserGUID()
}

func (repo *configRepository) UserEmail() string {
	return repo.cfconfig().UserEmail()
}

func (repo *configRepository) IsLoggedIn() bool {
	return repo.cfconfig().IsLoggedIn()
}

func (repo *configRepository) IsSSLDisabled() bool {
	return repo.cfconfig().IsSSLDisabled()
}

func (repo *configRepository) SetAPIVersion(version string) {
	repo.cfconfig().SetAPIVersion(version)
}

func (repo *configRepository) SetAPIEndpoint(endpoint string) {
	repo.cfconfig().SetAPIEndpoint(endpoint)
}

func (repo *configRepository) SetAuthenticationEndpoint(endpoint string) {
	repo.cfconfig().SetAuthenticationEndpoint(endpoint)
}

func (repo *configRepository) SetLoggregatorEndpoint(endpoint string) {
	repo.cfconfig().SetLoggregatorEndpoint(endpoint)
}

func (repo *configRepository) SetDopplerEndpoint(endpoint string) {
	repo.cfconfig().SetDopplerEndpoint(endpoint)
}

func (repo *configRepository) SetUaaEndpoint(endpoint string) {
	repo.cfconfig().SetUaaEndpoint(endpoint)
}

func (repo *configRepository) SetRoutingAPIEndpoint(endpoint string) {
	repo.cfconfig().SetRoutingAPIEndpoint(endpoint)
}

func (repo *configRepository) SetSSHOauthClient(client string) {
	repo.cfconfig().SetSSHOAuthClient(client)
}

func (repo *configRepository) SetAccessToken(token string) {
	repo.cfconfig().SetAccessToken(token)
}

func (repo *configRepository) SetRefreshToken(token string) {
	repo.cfconfig().SetRefreshToken(token)
}

func (repo *configRepository) SetOrganizationFields(org models.OrganizationFields) {
	var cfOrg cfmodels.OrganizationFields

	cfOrg.GUID = org.GUID
	cfOrg.Name = org.Name
	cfOrg.QuotaDefinition.GUID = org.QuotaDefinition.GUID
	cfOrg.QuotaDefinition.Name = org.QuotaDefinition.Name
	cfOrg.QuotaDefinition.MemoryLimit = org.QuotaDefinition.MemoryLimitInMB
	cfOrg.QuotaDefinition.InstanceMemoryLimit = org.QuotaDefinition.InstanceMemoryLimitInMB
	cfOrg.QuotaDefinition.RoutesLimit = org.QuotaDefinition.RoutesLimit
	cfOrg.QuotaDefinition.ServicesLimit = org.QuotaDefinition.ServicesLimit
	cfOrg.QuotaDefinition.NonBasicServicesAllowed = org.QuotaDefinition.NonBasicServicesAllowed
	cfOrg.QuotaDefinition.AppInstanceLimit = org.QuotaDefinition.AppInstanceLimit

	repo.cfconfig().SetOrganizationFields(cfOrg)
}

func (repo *configRepository) SetSpaceFields(space models.SpaceFields) {
	var cfSpace cfmodels.SpaceFields

	cfSpace.GUID = space.GUID
	cfSpace.Name = space.Name
	cfSpace.AllowSSH = space.AllowSSH

	repo.cfconfig().SetSpaceFields(cfSpace)
}

func (repo *configRepository) SetSSLDisabled(disabled bool) {
	repo.cfconfig().SetSSLDisabled(disabled)
}
