package validations

import (
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

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
	// Common field paths
	upstreamFieldPath := fieldPathSpec("upstream")
	remoteURLFieldPath := fieldPathSpec("remoteURL")
	volumeSizeFieldPath := fieldPathSpec("volume", "size")
	garbageCollectionTTLFieldPath := fieldPathSpec("garbageCollection", "ttl")
	httpProxyFieldPath := fieldPathSpec("proxy", "httpProxy")
	httpsProxyFieldPath := fieldPathSpec("proxy", "httpsProxy")

	// Shared test fixtures
	secretWithIncorrectStructure := buildSecret(
		"invalid-secret", "default", true,
		map[string][]byte{"invalid-key": []byte("invalid-value")},
	)
	mutableSecret := buildSecret(
		"mutable-secret", "default", false,
		map[string][]byte{"username": []byte("user"), "password": []byte("password")},
	)

	// DNS mocks
	dnsResolverAlwaysTrue := &mocks.DNSValidator{}
	dnsResolverAlwaysTrue.On("IsResolvable", mock.Anything).Return(true)

	dnsResolver := &mocks.DNSValidator{}
	dnsResolver.On("IsResolvable", "docker.io").Return(true)
	dnsResolver.On("IsResolvable", "registry-not-existing.not-exists.io").Return(false)
	dnsResolver.On("IsResolvable", "some.incorrect.repo.io").Return(false)

	t.Run("happy path", func(t *testing.T) {
		cfg := buildConfig("", "", registrycache.RegistryCacheConfigSpec{
			Upstream:  "docker.io",
			RemoteURL: ptr.To("https://registry-1.docker.io"),
		})
		errs := NewValidator(nil, nil, dnsResolverAlwaysTrue).Do(&cfg)
		validateResult(t, field.ErrorList{}, errs)
	})

	t.Run("spec emptiness", func(t *testing.T) {
		cfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{})
		errs := NewValidator(nil, nil, dnsResolverAlwaysTrue).Do(&cfg)
		validateResult(t, field.ErrorList{
			field.Required(field.NewPath("spec"), "spec must not be empty"),
		}, errs)
	})

	t.Run("field validation (invalid values)", func(t *testing.T) {
		cfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
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
		})
		errs := NewValidator(nil, nil, dnsResolverAlwaysTrue).Do(&cfg)
		validateResult(t, field.ErrorList{
			field.Invalid(upstreamFieldPath, InvalidUpstreamPort, "valid port must be in the range [1, 65535]"),
			field.Invalid(remoteURLFieldPath, InvalidRemoteURL, "url must start with 'http://' or 'https://'"),
			field.Invalid(volumeSizeFieldPath, InvalidVolumeSize, "must be greater than 0"),
			field.Invalid(garbageCollectionTTLFieldPath, "-1ns", "ttl must be a non-negative duration"),
			field.Invalid(httpProxyFieldPath, InvalidHttpProxyUrl, "url must start with 'http://' or 'https://'"),
			field.Invalid(httpProxyFieldPath, InvalidHttpProxyUrl, "subdomain must consist of lower case alphanumeric characters"),
			field.Invalid(httpsProxyFieldPath, InvalidHttpsProxyUrl, "url must start with 'http://' or 'https://'"),
			field.Invalid(httpsProxyFieldPath, InvalidHttpsProxyUrl, "subdomain must consist of lower case alphanumeric characters"),
		}, errs)
	})

	t.Run("uniqueness", func(t *testing.T) {
		cfg := buildConfig("config2", "default", registrycache.RegistryCacheConfigSpec{
			Upstream: "docker.io",
		})
		existing := []registrycache.RegistryCacheConfig{
			buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{Upstream: "docker.io"}),
		}
		errs := NewValidator(nil, existing, dnsResolverAlwaysTrue).Do(&cfg)
		validateResult(t, field.ErrorList{
			field.Duplicate(upstreamFieldPath, "docker.io"),
		}, errs)
	})

	t.Run("dns upstream", func(t *testing.T) {
		cfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream: "some.incorrect.repo.io",
		})
		errs := NewValidator(nil, nil, dnsResolver).Do(&cfg)
		validateResult(t, field.ErrorList{
			field.Invalid(upstreamFieldPath, "some.incorrect.repo.io", "upstream is not DNS resolvable"),
		}, errs)
	})

	t.Run("dns remoteURL", func(t *testing.T) {
		cfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream:  "docker.io",
			RemoteURL: ptr.To("https://registry-not-existing.not-exists.io"),
		})
		errs := NewValidator(nil, nil, dnsResolver).Do(&cfg)
		validateResult(t, field.ErrorList{
			field.Invalid(remoteURLFieldPath, ptr.To("https://registry-not-existing.not-exists.io"), "remoteURL is not DNS resolvable"),
		}, errs)
	})

	t.Run("secret validity", func(t *testing.T) {
		t.Run("non existent", func(t *testing.T) {
			cfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To("non-existent-secret"),
			})
			errs := NewValidator(nil, nil, dnsResolverAlwaysTrue).Do(&cfg)
			validateResult(t, field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), "non-existent-secret", "secret does not exist"),
			}, errs)
		})
		t.Run("invalid structure", func(t *testing.T) {
			cfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(secretWithIncorrectStructure.Name),
			})
			errs := NewValidator([]v1.Secret{secretWithIncorrectStructure}, nil, dnsResolverAlwaysTrue).Do(&cfg)
			validateResult(t, field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), secretWithIncorrectStructure.Name, "two data entries"),
				field.Invalid(fieldPathSpec("secretReferenceName"), secretWithIncorrectStructure.Name, "missing \"username\" data entry"),
				field.Invalid(fieldPathSpec("secretReferenceName"), secretWithIncorrectStructure.Name, "missing \"password\" data entry"),
			}, errs)
		})
		t.Run("mutable secret", func(t *testing.T) {
			cfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(mutableSecret.Name),
			})
			errs := NewValidator([]v1.Secret{mutableSecret}, nil, dnsResolverAlwaysTrue).Do(&cfg)
			validateResult(t, field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), mutableSecret.Name, "should be immutable"),
			}, errs)
		})
	})
}

func TestDoOnUpdate(t *testing.T) {
	// Field paths
	volumeSizeFieldPath := fieldPathSpec("volume", "size")
	volumeStorageClassNameFieldPath := fieldPathSpec("volume", "storageClassName")
	garbageCollectionTTLFieldPath := fieldPathSpec("garbageCollection", "ttl")

	// Secrets
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

	// DNS mocks
	dnsResolverAlwaysTrue := &mocks.DNSValidator{}
	dnsResolverAlwaysTrue.On("IsResolvable", mock.Anything).Return(true)

	dnsResolver := &mocks.DNSValidator{}
	dnsResolver.On("IsResolvable", "docker.io").Return(true)
	dnsResolver.On("IsResolvable", "registry-not-existing.not-exists.io").Return(false)
	dnsResolver.On("IsResolvable", "some.incorrect.repo.io").Return(false)

	t.Run("happy path", func(t *testing.T) {
		oldCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream: "docker.io",
		})
		newCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream:            "docker.io",
			SecretReferenceName: ptr.To(validSecret.Name),
		})
		errs := NewValidator([]v1.Secret{validSecret}, []registrycache.RegistryCacheConfig{
			buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{Upstream: "docker.io"}),
		}, dnsResolverAlwaysTrue).DoOnUpdate(&newCfg, &oldCfg)
		validateResult(t, field.ErrorList{}, errs)
	})

	t.Run("spec emptiness", func(t *testing.T) {
		oldCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream: "quay.io",
		})
		newCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{})
		errs := NewValidator(nil, nil, dnsResolverAlwaysTrue).DoOnUpdate(&newCfg, &oldCfg)
		validateResult(t, field.ErrorList{
			field.Required(field.NewPath("spec"), "spec must not be empty"),
		}, errs)
	})

	t.Run("immutability and garbage collection enabling", func(t *testing.T) {
		oldCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Volume: &registrycache.Volume{
				Size:             ptr.To(resource.MustParse("10Gi")),
				StorageClassName: ptr.To("standard"),
			},
			GarbageCollection: &registrycache.GarbageCollection{
				TTL: metav1.Duration{Duration: 0},
			},
		})
		newCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Volume: &registrycache.Volume{
				Size:             ptr.To(resource.MustParse(NewVolumeSize)),
				StorageClassName: ptr.To(NewStorageClassName),
			},
			GarbageCollection: &registrycache.GarbageCollection{
				TTL: metav1.Duration{Duration: 1 * time.Hour},
			},
		})
		errs := NewValidator(nil, nil, dnsResolverAlwaysTrue).DoOnUpdate(&newCfg, &oldCfg)
		validateResult(t, field.ErrorList{
			field.Invalid(volumeSizeFieldPath, NewVolumeSize, "field is immutable"),
			field.Invalid(volumeStorageClassNameFieldPath, ptr.To(NewStorageClassName), "field is immutable"),
			field.Invalid(garbageCollectionTTLFieldPath, &registrycacheext.GarbageCollection{
				TTL: metav1.Duration{Duration: 1 * time.Hour},
			}, "garbage collection cannot be enabled"),
		}, errs)
	})

	t.Run("uniqueness", func(t *testing.T) {
		oldCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream: "quay.io",
		})
		newCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream: "docker.io",
		})
		existing := []registrycache.RegistryCacheConfig{
			buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{Upstream: "quay.io"}),
			buildConfig("config2", "default", registrycache.RegistryCacheConfigSpec{Upstream: "docker.io"}),
		}
		errs := NewValidator(nil, existing, dnsResolverAlwaysTrue).DoOnUpdate(&newCfg, &oldCfg)
		validateResult(t, field.ErrorList{
			field.Duplicate(fieldPathSpec("upstream"), "docker.io"),
		}, errs)
	})

	t.Run("dns upstream", func(t *testing.T) {
		oldCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream: "quay.io",
		})
		newCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream: "some.incorrect.repo.io",
		})
		errs := NewValidator([]v1.Secret{}, []registrycache.RegistryCacheConfig{
			buildConfig("", "", registrycache.RegistryCacheConfigSpec{Upstream: "docker.io"}),
		}, dnsResolver).DoOnUpdate(&newCfg, &oldCfg)
		validateResult(t, field.ErrorList{
			field.Invalid(fieldPathSpec("upstream"), "some.incorrect.repo.io", "upstream is not DNS resolvable"),
		}, errs)
	})

	t.Run("dns remoteURL", func(t *testing.T) {
		oldCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream: "quay.io",
		})
		newCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream:  "docker.io",
			RemoteURL: ptr.To("https://registry-not-existing.not-exists.io"),
		})
		errs := NewValidator(nil, []registrycache.RegistryCacheConfig{
			buildConfig("", "", registrycache.RegistryCacheConfigSpec{Upstream: "quay.io"}),
		}, dnsResolver).DoOnUpdate(&newCfg, &oldCfg)
		validateResult(t, field.ErrorList{
			field.Invalid(fieldPathSpec("remoteURL"), ptr.To("https://registry-not-existing.not-exists.io"), "remoteURL is not DNS resolvable"),
		}, errs)
	})

	t.Run("secret validity", func(t *testing.T) {
		t.Run("invalid structure", func(t *testing.T) {
			oldCfg := buildConfig("config-with-invalid-secret", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(validSecret.Name),
			})
			newCfg := buildConfig("config-with-invalid-secret", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(secretWithIncorrectStructure.Name),
			})
			errs := NewValidator([]v1.Secret{secretWithIncorrectStructure}, nil, dnsResolverAlwaysTrue).DoOnUpdate(&newCfg, &oldCfg)
			validateResult(t, field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), secretWithIncorrectStructure.Name, "two data entries"),
				field.Invalid(fieldPathSpec("secretReferenceName"), secretWithIncorrectStructure.Name, "missing \"username\" data entry"),
				field.Invalid(fieldPathSpec("secretReferenceName"), secretWithIncorrectStructure.Name, "missing \"password\" data entry"),
			}, errs)
		})
		t.Run("mutable secret", func(t *testing.T) {
			oldCfg := buildConfig("config-with-invalid-secret", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(validSecret.Name),
			})
			newCfg := buildConfig("config-with-invalid-secret", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(mutableSecret.Name),
			})
			errs := NewValidator([]v1.Secret{validSecret, mutableSecret}, nil, dnsResolverAlwaysTrue).DoOnUpdate(&newCfg, &oldCfg)
			validateResult(t, field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), mutableSecret.Name, "should be immutable"),
			}, errs)
		})
	})
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
