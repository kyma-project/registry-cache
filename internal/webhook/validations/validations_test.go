package validations

import (
	registrycacheext "github.com/gardener/gardener-extension-registry-cache/pkg/apis/registry"
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
	"time"
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
	upstreamFieldPath := fieldPathSpec("upstream")
	remoteURLFieldPath := fieldPathSpec("remoteURL")
	volumeSizeFieldPath := fieldPathSpec("volume", "size")
	garbageCollectionTTLFieldPath := fieldPathSpec("garbageCollection", "ttl")
	httpProxyFieldPath := fieldPathSpec("proxy", "httpProxy")
	httpsProxyFieldPath := fieldPathSpec("proxy", "httpsProxy")

	secretWithIncorrectStructure := buildSecret(
		"invalid-secret", "default", true,
		map[string][]byte{"invalid-key": []byte("invalid-value")},
	)
	mutableSecret := buildSecret(
		"mutable-secret", "default", false,
		map[string][]byte{"username": []byte("user"), "password": []byte("password")},
	)

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
			RegistryCacheConfig: buildConfig("", "", registrycache.RegistryCacheConfigSpec{
				Upstream:  "docker.io",
				RemoteURL: ptr.To("https://registry-1.docker.io"),
			}),
			errorsList:   field.ErrorList{},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name:                "empty spec",
			RegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{}),
			errorsList: field.ErrorList{
				field.Required(field.NewPath("spec"), "spec must not be empty"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name: "invalid spec",
			RegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
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
			}),
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
			RegistryCacheConfig: buildConfig("config2", "default", registrycache.RegistryCacheConfigSpec{
				Upstream: "docker.io",
			}),
			existingConfigs: []registrycache.RegistryCacheConfig{
				buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
					Upstream: "docker.io",
				}),
			},
			errorsList: field.ErrorList{
				field.Duplicate(upstreamFieldPath, "docker.io"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name: "upstream non-resolvable",
			RegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream: "some.incorrect.repo.io",
			}),
			existingConfigs: []registrycache.RegistryCacheConfig{},
			errorsList: field.ErrorList{
				field.Invalid(upstreamFieldPath, "some.incorrect.repo.io", "upstream is not DNS resolvable"),
			},
			dnsValidator: dnsResolver,
		},
		{
			name: "remote url non-resolvable",
			RegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:  "docker.io",
				RemoteURL: ptr.To("https://registry-not-existing.not-exists.io"),
			}),
			existingConfigs: []registrycache.RegistryCacheConfig{},
			errorsList: field.ErrorList{
				field.Invalid(remoteURLFieldPath, ptr.To("https://registry-not-existing.not-exists.io"), "remoteURL is not DNS resolvable"),
			},
			dnsValidator: dnsResolver,
		},
		{
			name: "non existent secret reference name",
			RegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To("non-existent-secret"),
			}),
			errorsList: field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), "non-existent-secret", "secret does not exist"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name:    "secret with invalid structure",
			secrets: []v1.Secret{secretWithIncorrectStructure},
			RegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(secretWithIncorrectStructure.Name),
			}),
			errorsList: field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), secretWithIncorrectStructure.Name, "two data entries"),
				field.Invalid(fieldPathSpec("secretReferenceName"), secretWithIncorrectStructure.Name, "missing \"username\" data entry"),
				field.Invalid(fieldPathSpec("secretReferenceName"), secretWithIncorrectStructure.Name, "missing \"password\" data entry"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name:    "mutable secret",
			secrets: []v1.Secret{mutableSecret},
			RegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(mutableSecret.Name),
			}),
			errorsList: field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), mutableSecret.Name, "should be immutable"),
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
	volumeSizeFieldPath := fieldPathSpec("volume", "size")
	volumeStorageClassNameFieldPath := fieldPathSpec("volume", "storageClassName")
	garbageCollectionTTLFieldPath := fieldPathSpec("garbageCollection", "ttl")

	validSecret := buildSecret(
		"valid-secret", "default", true,
		map[string][]byte{"username": []byte("user"), "password": []byte("password")},
	)
	secretWithIncorrectStructure := buildSecret(
		"invalid-secret", "default", true,
		map[string][]byte{"invalid-key": []byte("invalid-value")},
	)
	mutableSecret := buildSecret(
		"mutable-secret", "default", false,
		map[string][]byte{"username": []byte("user"), "password": []byte("password")},
	)

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
		{
			name: "valid spec",
			oldRegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream: "docker.io",
			}),
			newRegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(validSecret.Name),
			}),
			errorsList:   field.ErrorList{},
			dnsValidator: dnsResolverAlwaysTrue,
			existingConfigs: []registrycache.RegistryCacheConfig{
				buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
					Upstream: "docker.io",
				}),
			},
			secrets: []v1.Secret{validSecret},
		},
		{
			name: "empty spec",
			oldRegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream: "quay.io",
			}),
			newRegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{}),
			errorsList: field.ErrorList{
				field.Required(field.NewPath("spec"), "spec must not be empty"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name: "invalid spec",
			oldRegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Volume: &registrycache.Volume{
					Size:             ptr.To(resource.MustParse("10Gi")),
					StorageClassName: ptr.To("standard"),
				},
				GarbageCollection: &registrycache.GarbageCollection{
					TTL: metav1.Duration{Duration: 0},
				},
			}),
			newRegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Volume: &registrycache.Volume{
					Size:             ptr.To(resource.MustParse(NewVolumeSize)),
					StorageClassName: ptr.To(NewStorageClassName),
				},
				GarbageCollection: &registrycache.GarbageCollection{
					TTL: metav1.Duration{Duration: 1 * time.Hour},
				},
			}),
			errorsList: field.ErrorList{
				field.Invalid(volumeSizeFieldPath, NewVolumeSize, "field is immutable"),
				field.Invalid(volumeStorageClassNameFieldPath, ptr.To(NewStorageClassName), "field is immutable"),
				field.Invalid(garbageCollectionTTLFieldPath, &registrycacheext.GarbageCollection{
					TTL: metav1.Duration{Duration: 1 * time.Hour},
				}, "garbage collection cannot be enabled"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name: "non existent secret reference name",
			oldRegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(validSecret.Name),
			}),
			newRegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To("non-existent-secret"),
			}),
			errorsList: field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), "non-existent-secret", "secret does not exist"),
			},
			secrets:      []v1.Secret{validSecret},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name: "duplicated upstream",
			oldRegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream: "quay.io",
			}),
			newRegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream: "docker.io",
			}),
			existingConfigs: []registrycache.RegistryCacheConfig{
				buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
					Upstream: "quay.io",
				}),
				buildConfig("config2", "default", registrycache.RegistryCacheConfigSpec{
					Upstream: "docker.io",
				}),
			},
			errorsList: field.ErrorList{
				field.Duplicate(fieldPathSpec("upstream"), "docker.io"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name: "upstream non-resolvable",
			oldRegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream: "quay.io",
			}),
			newRegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream: "some.incorrect.repo.io",
			}),
			existingConfigs: []registrycache.RegistryCacheConfig{
				buildConfig("", "", registrycache.RegistryCacheConfigSpec{
					Upstream: "docker.io",
				}),
			},
			errorsList: field.ErrorList{
				field.Invalid(fieldPathSpec("upstream"), "some.incorrect.repo.io", "upstream is not DNS resolvable"),
			},
			dnsValidator: dnsResolver,
		},
		{
			name: "remoteURL non-resolvable",
			oldRegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream: "quay.io",
			}),
			newRegistryCacheConfig: buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:  "docker.io",
				RemoteURL: ptr.To("https://registry-not-existing.not-exists.io"),
			}),
			existingConfigs: []registrycache.RegistryCacheConfig{
				buildConfig("", "", registrycache.RegistryCacheConfigSpec{
					Upstream: "quay.io",
				}),
			},
			errorsList: field.ErrorList{
				field.Invalid(fieldPathSpec("remoteURL"), ptr.To("https://registry-not-existing.not-exists.io"), "remoteURL is not DNS resolvable"),
			},
			dnsValidator: dnsResolver,
		},
		{
			name: "non existent secret reference name",
			oldRegistryCacheConfig: buildConfig("config", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(validSecret.Name),
			}),
			newRegistryCacheConfig: buildConfig("config", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To("non-existent-secret"),
			}),
			errorsList: field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), "non-existent-secret", "secret does not exist"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name:    "secret with invalid structure",
			secrets: []v1.Secret{secretWithIncorrectStructure},
			oldRegistryCacheConfig: buildConfig("config-with-invalid-secret", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(validSecret.Name),
			}),
			newRegistryCacheConfig: buildConfig("config-with-invalid-secret", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(secretWithIncorrectStructure.Name),
			}),
			errorsList: field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), secretWithIncorrectStructure.Name, "two data entries"),
				field.Invalid(fieldPathSpec("secretReferenceName"), secretWithIncorrectStructure.Name, "missing \"username\" data entry"),
				field.Invalid(fieldPathSpec("secretReferenceName"), secretWithIncorrectStructure.Name, "missing \"password\" data entry"),
			},
			dnsValidator: dnsResolverAlwaysTrue,
		},
		{
			name:    "mutable secret",
			secrets: []v1.Secret{validSecret, mutableSecret},
			oldRegistryCacheConfig: buildConfig("config-with-invalid-secret", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(validSecret.Name),
			}),
			newRegistryCacheConfig: buildConfig("config-with-invalid-secret", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(mutableSecret.Name),
			}),
			errorsList: field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), mutableSecret.Name, "should be immutable"),
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

func fieldPathSpec(parts ...string) *field.Path {
	p := field.NewPath("spec")
	for _, part := range parts {
		p = p.Child(part)
	}
	return p
}

func buildSecret(name, ns string, immutable bool, data map[string][]byte) v1.Secret {
	return v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Data:      data,
		Immutable: ptr.To(immutable),
	}
}

func buildConfig(name, ns string, spec registrycache.RegistryCacheConfigSpec) registrycache.RegistryCacheConfig {
	return registrycache.RegistryCacheConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: spec,
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
