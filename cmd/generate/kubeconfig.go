package generate

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	"sort"
)

const (
	// Use 'Remote Browser' when a browser is not locally available for the local session (i.e while using runai cli via ssh)
	// In this case the cli cannot open a browser and cannot listen for the auth response on localhost:8000 (default) so the redirect must bounce the browser to some generally
	// available location like app.run.ai/auth or <airgapped-backencd-url>/auth for airgapped envs.
	AuthMethodRemoteBrowser           = "remote-browser"
	AuthMethodBrowser                 = "browser"
	AuthMethodPassword                = "password"
	AuthMethodLocalClusterIdpPassword = "local-cluster-password" //auth0 and keycloak handle 'password' grant types a bit differently. This flag is for keycloak. The difference is mainly in the redirect url.


	ParamClientId       = "client-id"
	ParamClientSecret   = "client-secret"
	ParamIssuerUrl      = "idp-issuer-url"
	ParamRedirectUrl    = "redirect-url"
	ParamAuthMethod     = "auth-method"
	ParamExtraScopes    = "auth-request-extra-scopes"
	ParamAuthRealm      = "auth-realm"

	AuthProviderName     = "oidc"

	DefaultKubeConfigUserName = "runai-oidc"
	DefaultIssuerUrl          = "https://runai-prod.auth0.com/"
	DefaultRedirectUrl        = "https://app.run.ai/auth"
)

var (
	AllowedAuthMethods = []string{AuthMethodBrowser, AuthMethodLocalClusterIdpPassword, AuthMethodPassword, AuthMethodRemoteBrowser}

	paramKubeConfigUser string
	paramAuthMethod     string
	paramAuthRealm      string
	paramClientId       string
	paramClientSecret   string
	paramIssuerUrl      string
	paramRedirectUrl    string
	paramExtraScopes    string
)

func init() {
	sort.Strings(AllowedAuthMethods)
}

func KubeConfigGenerateCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:           "kubeconfig [path to kubeconfig file]",
		Short:         "Generates a kubeconfig file with authentication parameters for the cluster contained in the given kubeconfig file.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			var (
				kubeConfig *clientapi.Config
				err error
			)

			if kubeConfig, err = clientcmd.LoadFromFile(args[0]); err != nil {
				return err
			} else if err = promptUserForAuthConfig(); err != nil {
				return err
			} else if err = validateParams(); err != nil {
				return err
			}

			setAuthProvider(kubeConfig)
			if err := setContext(kubeConfig); err != nil {
				return err
			}
			rawConfig, err := clientcmd.Write(*kubeConfig)
			if err == nil {
				fmt.Println()
				fmt.Println(string(rawConfig))
			}
			return err
		},
	}

	// Required
	command.Flags().StringVar(&paramClientId, ParamClientId, "", "OIDC Client ID")
	command.Flags().StringVar(&paramClientSecret, ParamClientSecret, "", "OIDC Client Secret")

	// Required, but has defaults
	command.Flags().StringVar(&paramIssuerUrl, ParamIssuerUrl, DefaultIssuerUrl, "OIDC Issuer URL")
	command.Flags().StringVar(&paramRedirectUrl, ParamRedirectUrl, DefaultRedirectUrl, "Auth Response Redirect URL")
	command.Flags().StringVar(&paramAuthMethod, ParamAuthMethod, AuthMethodBrowser, "The method to use for initial authentication. can be one of [browser,remote-browser,password,local-password]")
	command.Flags().StringVar(&paramKubeConfigUser, "kube-config-user", DefaultKubeConfigUserName, "The user defined in the kubeconfig file to operate on")

	// Optional
	command.Flags().StringVar(&paramAuthRealm, ParamAuthRealm, "", "[password only] Governs which realm will be used when authenticating the user with the IDP")
	command.Flags().StringVar(&paramExtraScopes, ParamExtraScopes, "", "comma-delimited list of extra scopes to request with the ID token")

	return command
}

// promptUserForAuthConfig prompts user for basic required config.
func promptUserForAuthConfig() (err error) {
	if paramClientId == "" {
		if clientId, err := util.ReadString("Client ID: "); err == nil {
			paramClientId = clientId
		}
	}
	if paramClientSecret == "" {
		if clientSecret, err := util.ReadPassword("Client Secret: "); err == nil {
			paramClientSecret = clientSecret
		}
	}
	return
}

// validateParams validates all required params are present.
func validateParams() (err error) {
	authMethodPosition := sort.SearchStrings(AllowedAuthMethods, paramAuthMethod)
	if authMethodPosition > len(AllowedAuthMethods) || AllowedAuthMethods[authMethodPosition] != paramAuthMethod {
		err = fmt.Errorf("unknown auth method '%s'", paramAuthMethod)
	} else if paramClientId == "" {
		err = fmt.Errorf("client id is required")
	} else if paramClientSecret == "" {
		err = fmt.Errorf("client secret is required")
	} else if paramIssuerUrl == "" {
		err = fmt.Errorf("issuer url is required")
	} else if paramRedirectUrl == "" {
		err = fmt.Errorf("redirect url is required")
	}
	return
}

func setAuthProvider(kubeConfig *clientapi.Config) {
	if kubeConfig.AuthInfos == nil {
		kubeConfig.AuthInfos = make(map[string]*clientapi.AuthInfo)
	}
	kubeConfig.AuthInfos[paramKubeConfigUser] = &clientapi.AuthInfo{AuthProvider: createAuthProviderConfig()}
}

func createAuthProviderConfig() (authProviderConfig *clientapi.AuthProviderConfig) {
	authProviderConfig = &clientapi.AuthProviderConfig{
		Config: make(map[string]string),
		Name:   AuthProviderName,
	}
	// Required
	authProviderConfig.Config[ParamClientId] = paramClientId
	authProviderConfig.Config[ParamClientSecret] = paramClientSecret
	authProviderConfig.Config[ParamAuthMethod] = paramAuthMethod
	authProviderConfig.Config[ParamIssuerUrl] = paramIssuerUrl
	authProviderConfig.Config[ParamRedirectUrl] = paramRedirectUrl

	// Optional
	if paramAuthRealm != "" {
		authProviderConfig.Config[ParamAuthRealm] = paramAuthRealm
	}
	if paramExtraScopes != "" {
		authProviderConfig.Config[ParamExtraScopes] = paramExtraScopes
	}

	return
}

func setContext(kubeConfig *clientapi.Config) error {
	if len(kubeConfig.Clusters) != 1 {
		fmt.Println("setting a context is supported only when a single cluster is observed in the kubeconfig file.")
		return nil
	}
	var clusterName string
	for clusterName, _ = range kubeConfig.Clusters {
		//We validate there's only one cluster above, so we're fine.
		break
	}
	if kubeConfig.Contexts == nil {
		kubeConfig.Contexts = make(map[string]*clientapi.Context)
	}
	contextName := fmt.Sprintf("%s@%s", paramKubeConfigUser, clusterName)
	kubeConfig.Contexts[contextName] = &clientapi.Context{
		Cluster:  clusterName,
		AuthInfo: paramKubeConfigUser,
	}
	kubeConfig.CurrentContext = contextName
	return nil
}
