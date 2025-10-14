package validations

import (
	registrycache "github.com/kyma-project/registry-cache/api/v1beta1"
	"github.com/kyma-project/registry-cache/internal/webhook/validations/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
	"reflect"
	"slices"
	"strings"
	"testing"
)

const (
	InvalidUpstreamPort           = "docker.io:77777"
	InvalidRemoteURL              = "docker.io"
	InvalidVolumeSize             = "-1"
	InvalidGarbageCollectionValue = -1
	InvalidHttpProxyUrl           = "http//invalid-url"
	InvalidHttpsProxyUrl          = "https//invalid-url"

	NewVolumeSize       = "20Gi"
	NewStorageClassName = "nonstandard"
)

func TestDo(t *testing.T) {

	upstreamFieldPath := field.NewPath("spec").Child("upstream")
	remoteURLFieldPath := field.NewPath("spec").Child("remoteURL")
	volumeSizeFieldPath := field.NewPath("spec").Child("volume").Child("size")
	garbageCollectionTTLFieldPath := field.NewPath("spec").Child("garbageCollection").Child("ttl")
	httpProxyFieldPath := field.NewPath("spec").Child("proxy").Child("httpProxy")
	httpsProxyFieldPath := field.NewPath("spec").Child("proxy").Child("httpsProxy")

	secretWithIncorrectStructure := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "invalid-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"invalid-key": []byte("invalid-value"),
		},
		Immutable: ptr.To(true),
	}

	mutableSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mutable-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"username": []byte("user"),
			"password": []byte("password"),
		},
		Immutable: ptr.To(false),
	}

	dnsResolverAlwaysTrue := &mocks.DNSValidator{}
	dnsResolverAlwaysTrue.On("IsResolvable", mock.Anything).Return(true)

	dnsResolver := &mocks.DNSValidator{}
	dnsResolver.On("IsResolvable", "docker.io").Return(true)
	dnsResolver.On("IsResolvable", "registry-not-existing.not-exists.io").Return(false)
	dnsResolver.On("IsResolvable", "some.incorrect.repo.io").Return(false)

	for _, tt := range []struct {
		name string
		registrycache.RegistryCacheConfig
		existingConfigs []registrycache.RegistryCacheConfig
		errorsList      field.ErrorList
		secrets         []v1.Secret
		dnsValidator    DNSValidator
	}{
		{
			name: "valid spec",
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:  "docker.io",
					RemoteURL: ptr.To("https://registry-1.docker.io"),
				},
			},
			errorsList:   field.ErrorList{},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name: "empty spec",
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{},
			},
			errorsList: field.ErrorList{
				field.Required(field.NewPath("spec"), "spec must not be empty"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name: "invalid spec",
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:  InvalidUpstreamPort,
					RemoteURL: ptr.To(InvalidRemoteURL),
					Volume: &registrycache.Volume{
						Size: ptr.To(resource.MustParse(InvalidVolumeSize)),
					},
					GarbageCollection: &registrycache.GarbageCollection{
						TTL: metav1.Duration{Duration: InvalidGarbageCollectionValue},
					},
					Proxy: &registrycache.Proxy{
						HTTPProxy:  ptr.To(InvalidHttpProxyUrl),
						HTTPSProxy: ptr.To(InvalidHttpsProxyUrl),
					},
				},
			},
			errorsList: field.ErrorList{
				field.Invalid(upstreamFieldPath, InvalidUpstreamPort, "valid port must be in the range [1, 65535]"),
				field.Invalid(remoteURLFieldPath, InvalidRemoteURL, "url must start with 'http://' or 'https://'"),
				field.Invalid(volumeSizeFieldPath, InvalidVolumeSize, "must be greater than 0"),
				field.Invalid(garbageCollectionTTLFieldPath, "-1ns", "ttl must be a non-negative duration"),
				field.Invalid(httpProxyFieldPath, InvalidHttpProxyUrl, "url must start with 'http://' or 'https://"),
				field.Invalid(httpProxyFieldPath, InvalidHttpProxyUrl, "subdomain must consist of lower case alphanumeric characters"),
				field.Invalid(httpsProxyFieldPath, InvalidHttpsProxyUrl, "url must start with 'http://' or 'https://"),
				field.Invalid(httpsProxyFieldPath, InvalidHttpsProxyUrl, "subdomain must consist of lower case alphanumeric characters"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name: "duplicated upstream",
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream: "docker.io",
				},
			},
			existingConfigs: []registrycache.RegistryCacheConfig{
				{
					Spec: registrycache.RegistryCacheConfigSpec{
						Upstream: "docker.io",
					},
				},
			},
			errorsList: field.ErrorList{
				field.Duplicate(field.NewPath("spec").Child("upstream"), "docker.io"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name: "upstream non-resolvable",
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream: "some.incorrect.repo.io",
				},
			},
			existingConfigs: []registrycache.RegistryCacheConfig{},
			errorsList: field.ErrorList{
				field.Invalid(field.NewPath("spec").Child("upstream"), "some.incorrect.repo.io", "upstream is not DNS resolvable"),
			},
			dnsValidator: dnsResolver,
		},
		{
			name: "remote url non-resolvable",
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:  "docker.io",
					RemoteURL: ptr.To("https://registry-not-existing.not-exists.io"),
				},
			},
			existingConfigs: []registrycache.RegistryCacheConfig{},
			errorsList: field.ErrorList{
				field.Invalid(field.NewPath("spec").Child("remoteURL"), ptr.To("https://registry-not-existing.not-exists.io"), "remoteURL is not DNS resolvable"),
			},
			dnsValidator: dnsResolver,
		},
		{
			name: "non existent secret reference name",
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:            "docker.io",
					SecretReferenceName: ptr.To("non-existent-secret"),
				},
			},
			errorsList: field.ErrorList{
				field.Invalid(field.NewPath("spec").Child("secretReferenceName"), "non-existent-secret", "secret does not exist"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name:    "secret with invalid structure",
			secrets: []v1.Secret{secretWithIncorrectStructure},
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "config-with-invalid-secret",
					Namespace: "default",
				},
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:            "docker.io",
					SecretReferenceName: ptr.To(secretWithIncorrectStructure.Name),
				},
			},
			errorsList: field.ErrorList{
				field.Invalid(field.NewPath("spec").Child("secretReferenceName"), secretWithIncorrectStructure.Name, "two data entries"),
				field.Invalid(field.NewPath("spec").Child("secretReferenceName"), secretWithIncorrectStructure.Name, "missing \"username\" data entry"),
				field.Invalid(field.NewPath("spec").Child("secretReferenceName"), secretWithIncorrectStructure.Name, "missing \"password\" data entry"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name:    "mutable secret",
			secrets: []v1.Secret{mutableSecret},
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "config-with-invalid-secret",
					Namespace: "default",
				},
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:            "docker.io",
					SecretReferenceName: ptr.To(mutableSecret.Name),
				},
			},
			errorsList: field.ErrorList{
				field.Invalid(field.NewPath("spec").Child("secretReferenceName"), mutableSecret.Name, "should be immutable"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			errs := NewValidator(tt.secrets, tt.existingConfigs, tt.dnsValidator).Do(&tt.RegistryCacheConfig)

			validateResult(t, tt.errorsList, errs)
		})
	}
}

func TestDoOnUpdate(t *testing.T) {

	//volumeSizeFieldPath := field.NewPath("spec").Child("volume").Child("size")
	//volumeStorageClassNameFieldPath := field.NewPath("spec").Child("volume").Child("storageClassName")
	//garbageCollectionTTLFieldPath := field.NewPath("spec").Child("garbageCollection").Child("ttl")

	validSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "valid-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"username": []byte("user"),
			"password": []byte("password"),
		},
		Immutable: ptr.To(true),
	}

	secretWithIncorrectStructure := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "invalid-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"invalid-key": []byte("invalid-value"),
		},
		Immutable: ptr.To(true),
	}

	mutableSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mutable-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"username": []byte("user"),
			"password": []byte("password"),
		},
		Immutable: ptr.To(false),
	}

	dnsResolverAlwaysTrue := &mocks.DNSValidator{}
	dnsResolverAlwaysTrue.On("IsResolvable", mock.Anything).Return(true)

	dnsResolver := &mocks.DNSValidator{}
	dnsResolver.On("IsResolvable", "docker.io").Return(true)
	dnsResolver.On("IsResolvable", "registry-not-existing.not-exists.io").Return(false)
	dnsResolver.On("IsResolvable", "some.incorrect.repo.io").Return(false)

	for _, tt := range []struct {
		name                   string
		newRegistryCacheConfig registrycache.RegistryCacheConfig
		oldRegistryCacheConfig registrycache.RegistryCacheConfig
		existingConfigs        []registrycache.RegistryCacheConfig
		errorsList             field.ErrorList
		secrets                []v1.Secret
		dnsValidator           DNSValidator
	}{
		//{
		//	name: "valid spec",
		//	oldRegistryCacheConfig: registrycache.RegistryCacheConfig{
		//		Spec: registrycache.RegistryCacheConfigSpec{
		//			Upstream: "quay.io",
		//		},
		//	},
		//	newRegistryCacheConfig: registrycache.RegistryCacheConfig{
		//		Spec: registrycache.RegistryCacheConfigSpec{
		//			Upstream: "docker.io",
		//		},
		//	},
		//	errorsList:   field.ErrorList{},
		//	dnsValidator: dnsResolverAlwaysTrue,
		//},
		//{
		//	name: "empty spec",
		//	oldRegistryCacheConfig: registrycache.RegistryCacheConfig{
		//		Spec: registrycache.RegistryCacheConfigSpec{
		//			Upstream: "quay.io",
		//		},
		//	},
		//	newRegistryCacheConfig: registrycache.RegistryCacheConfig{
		//		Spec: registrycache.RegistryCacheConfigSpec{},
		//	},
		//	errorsList: field.ErrorList{
		//		field.Required(field.NewPath("spec"), "spec must not be empty"),
		//	},
		//	dnsValidator: dnsResolverAlwaysTrue,
		//},
		//{
		//	name: "invalid spec",
		//	oldRegistryCacheConfig: registrycache.RegistryCacheConfig{
		//		Spec: registrycache.RegistryCacheConfigSpec{
		//			Volume: &registrycache.Volume{
		//				Size:             ptr.To(resource.MustParse("10Gi")),
		//				StorageClassName: ptr.To("standard"),
		//			},
		//			GarbageCollection: &registrycache.GarbageCollection{
		//				TTL: metav1.Duration{Duration: 0},
		//			},
		//		},
		//	},
		//	newRegistryCacheConfig: registrycache.RegistryCacheConfig{
		//		Spec: registrycache.RegistryCacheConfigSpec{
		//			Volume: &registrycache.Volume{
		//				Size:             ptr.To(resource.MustParse(NewVolumeSize)),
		//				StorageClassName: ptr.To(NewStorageClassName),
		//			},
		//			GarbageCollection: &registrycache.GarbageCollection{
		//				TTL: metav1.Duration{Duration: 1 * time.Hour},
		//			},
		//		},
		//	},
		//	errorsList: field.ErrorList{
		//		field.Invalid(volumeSizeFieldPath, NewVolumeSize, "field is immutable"),
		//		field.Invalid(volumeStorageClassNameFieldPath, ptr.To(NewStorageClassName), "field is immutable"),
		//		field.Invalid(garbageCollectionTTLFieldPath, &registrycacheext.GarbageCollection{
		//			TTL: metav1.Duration{Duration: 1 * time.Hour},
		//		}, "garbage collection cannot be enabled"),
		//	},
		//	dnsValidator: dnsResolverAlwaysTrue,
		//},
		//{
		//	name: "non existent secret reference name",
		//	oldRegistryCacheConfig: registrycache.RegistryCacheConfig{
		//		Spec: registrycache.RegistryCacheConfigSpec{
		//			Upstream:            "docker.io",
		//			SecretReferenceName: ptr.To(validSecret.Name),
		//		},
		//	},
		//	newRegistryCacheConfig: registrycache.RegistryCacheConfig{
		//		Spec: registrycache.RegistryCacheConfigSpec{
		//			Upstream:            "docker.io",
		//			SecretReferenceName: ptr.To("non-existent-secret"),
		//		},
		//	},
		//	errorsList: field.ErrorList{
		//		field.Invalid(field.NewPath("spec").Child("secretReferenceName"), "non-existent-secret", "secret does not exist"),
		//	},
		//	secrets: []v1.Secret{
		//		validSecret,
		//	},
		//	dnsValidator: dnsResolverAlwaysTrue,
		//},
		//{
		//	name: "duplicated upstream",
		//	oldRegistryCacheConfig: registrycache.RegistryCacheConfig{
		//		Spec: registrycache.RegistryCacheConfigSpec{
		//			Upstream: "quay.io",
		//		},
		//	},
		//	newRegistryCacheConfig: registrycache.RegistryCacheConfig{
		//		Spec: registrycache.RegistryCacheConfigSpec{
		//			Upstream: "docker.io",
		//		},
		//	},
		//	existingConfigs: []registrycache.RegistryCacheConfig{
		//		{
		//			Spec: registrycache.RegistryCacheConfigSpec{
		//				Upstream: "docker.io",
		//			},
		//		},
		//	},
		//	errorsList: field.ErrorList{
		//		field.Duplicate(field.NewPath("spec").Child("upstream"), "docker.io"),
		//	},
		//	dnsValidator: dnsResolverAlwaysTrue,
		//},
		{
			name: "upstream non-resolvable",
			oldRegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream: "quay.io",
				},
			},
			newRegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream: "some.incorrect.repo.io",
				},
			},
			existingConfigs: []registrycache.RegistryCacheConfig{
				{
					Spec: registrycache.RegistryCacheConfigSpec{
						Upstream: "docker.io",
					},
				},
			},
			errorsList: field.ErrorList{
				field.Invalid(field.NewPath("spec").Child("upstream"), "some.incorrect.repo.io", "upstream is not DNS resolvable"),
			},
			dnsValidator: dnsResolver,
		},
		{
			name: "remoteURL non-resolvable",
			oldRegistryCacheConfig: registrycache.RegistryCacheConfig{

				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream: "quay.io",
				},
			},
			newRegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:  "docker.io",
					RemoteURL: ptr.To("https://registry-not-existing.not-exists.io"),
				},
			},
			existingConfigs: []registrycache.RegistryCacheConfig{
				{
					Spec: registrycache.RegistryCacheConfigSpec{
						Upstream: "quay.io",
					},
				},
			},
			errorsList: field.ErrorList{
				field.Invalid(field.NewPath("spec").Child("remoteURL"), ptr.To("https://registry-not-existing.not-exists.io"), "remoteURL is not DNS resolvable"),
			},
			dnsValidator: dnsResolver,
		},
		{
			name: "non existent secret reference name",
			oldRegistryCacheConfig: registrycache.RegistryCacheConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "config",
					Namespace: "default",
				},
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:            "docker.io",
					SecretReferenceName: ptr.To(validSecret.Name),
				},
			},
			newRegistryCacheConfig: registrycache.RegistryCacheConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "config",
					Namespace: "default",
				},
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:            "docker.io",
					SecretReferenceName: ptr.To("non-existent-secret"),
				},
			},
			errorsList: field.ErrorList{
				field.Invalid(field.NewPath("spec").Child("secretReferenceName"), "non-existent-secret", "secret does not exist"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name:    "secret with invalid structure",
			secrets: []v1.Secret{secretWithIncorrectStructure},
			oldRegistryCacheConfig: registrycache.RegistryCacheConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "config-with-invalid-secret",
					Namespace: "default",
				},
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:            "docker.io",
					SecretReferenceName: ptr.To(validSecret.Name),
				},
			},
			newRegistryCacheConfig: registrycache.RegistryCacheConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "config-with-invalid-secret",
					Namespace: "default",
				},
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:            "docker.io",
					SecretReferenceName: ptr.To(secretWithIncorrectStructure.Name),
				},
			},
			errorsList: field.ErrorList{
				field.Invalid(field.NewPath("spec").Child("secretReferenceName"), secretWithIncorrectStructure.Name, "two data entries"),
				field.Invalid(field.NewPath("spec").Child("secretReferenceName"), secretWithIncorrectStructure.Name, "missing \"username\" data entry"),
				field.Invalid(field.NewPath("spec").Child("secretReferenceName"), secretWithIncorrectStructure.Name, "missing \"password\" data entry"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name:    "mutable secret",
			secrets: []v1.Secret{validSecret, mutableSecret},
			oldRegistryCacheConfig: registrycache.RegistryCacheConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "config-with-invalid-secret",
					Namespace: "default",
				},
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:            "docker.io",
					SecretReferenceName: ptr.To(validSecret.Name),
				},
			},
			newRegistryCacheConfig: registrycache.RegistryCacheConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "config-with-invalid-secret",
					Namespace: "default",
				},
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:            "docker.io",
					SecretReferenceName: ptr.To(mutableSecret.Name),
				},
			},
			errorsList: field.ErrorList{
				field.Invalid(field.NewPath("spec").Child("secretReferenceName"), mutableSecret.Name, "should be immutable"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			errs := NewValidator(tt.secrets, tt.existingConfigs, tt.dnsValidator).DoOnUpdate(&tt.newRegistryCacheConfig, &tt.oldRegistryCacheConfig)

			validateResult(t, tt.errorsList, errs)
		})
	}
}

func validateResult(t *testing.T, expectedErrors field.ErrorList, actualErrors field.ErrorList) {
	require.Equal(t, len(expectedErrors), len(actualErrors))

	for _, expectedErr := range expectedErrors {
		actualErrIndex := slices.IndexFunc(actualErrors, func(err *field.Error) bool {
			return err.Type == expectedErr.Type &&
				expectedErr.Field == err.Field &&
				reflect.DeepEqual(expectedErr.BadValue, err.BadValue) &&
				strings.Contains(err.Detail, expectedErr.Detail)
		})
		require.NotEqual(t, -1, actualErrIndex, "actual error not found: %v", expectedErr)
	}
}
