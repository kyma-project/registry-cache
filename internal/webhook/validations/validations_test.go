package validations

import (
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

type testEnv struct {
	upstreamFieldPath               *field.Path
	remoteURLFieldPath              *field.Path
	volumeSizeFieldPath             *field.Path
	volumeStorageClassNameFieldPath *field.Path
	garbageCollectionTTLFieldPath   *field.Path
	httpProxyFieldPath              *field.Path
	httpsProxyFieldPath             *field.Path
	validSecret                     v1.Secret
	invalidSecret                   v1.Secret
	mutableSecret                   v1.Secret
	dnsResolverAllOK                *mocks.DNSValidator
	dnsResolver                     *mocks.DNSValidator
}

func newTestEnv() testEnv {
	// field paths
	env := testEnv{
		upstreamFieldPath:               fieldPathSpec("upstream"),
		remoteURLFieldPath:              fieldPathSpec("remoteURL"),
		volumeSizeFieldPath:             fieldPathSpec("volume", "size"),
		volumeStorageClassNameFieldPath: fieldPathSpec("volume", "storageClassName"),
		garbageCollectionTTLFieldPath:   fieldPathSpec("garbageCollection", "ttl"),
		httpProxyFieldPath:              fieldPathSpec("proxy", "httpProxy"),
		httpsProxyFieldPath:             fieldPathSpec("proxy", "httpsProxy"),
		validSecret: buildSecret("valid-secret", "default", true, map[string][]byte{
			"username": []byte("user"),
			"password": []byte("password"),
		}),
		invalidSecret: buildSecret("invalid-secret", "default", true, map[string][]byte{
			"invalid-key": []byte("invalid-value"),
		}),
		mutableSecret: buildSecret("mutable-secret", "default", false, map[string][]byte{
			"username": []byte("user"),
			"password": []byte("password"),
		}),
		dnsResolverAllOK: &mocks.DNSValidator{},
		dnsResolver:      &mocks.DNSValidator{},
	}

	env.dnsResolverAllOK.On("IsResolvable", mock.Anything).Return(true)
	env.dnsResolver.On("IsResolvable", "docker.io").Return(true)
	env.dnsResolver.On("IsResolvable", "registry-not-existing.not-exists.io").Return(false)
	env.dnsResolver.On("IsResolvable", "some.incorrect.repo.io").Return(false)

	return env
}

func TestDo(t *testing.T) {
	env := newTestEnv()

	t.Run("happy path", func(t *testing.T) {
		cfg := buildConfig("", "", registrycache.RegistryCacheConfigSpec{
			Upstream:  "docker.io",
			RemoteURL: ptr.To("https://registry-1.docker.io"),
		})

		errs := NewValidator(env.dnsResolverAllOK, fixFakeClient()).Do(&cfg)
		validateResult(t, field.ErrorList{}, errs)
	})

	t.Run("spec emptiness", func(t *testing.T) {
		cfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{})

		errs := NewValidator(env.dnsResolverAllOK, fixFakeClient()).Do(&cfg)
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

		errs := NewValidator(env.dnsResolverAllOK, fixFakeClient()).Do(&cfg)
		validateResult(t, field.ErrorList{
			field.Invalid(env.upstreamFieldPath, InvalidUpstreamPort, "valid port must be in the range [1, 65535]"),
			field.Invalid(env.remoteURLFieldPath, InvalidRemoteURL, "url must start with 'http://' or 'https://'"),
			field.Invalid(env.volumeSizeFieldPath, InvalidVolumeSize, "must be greater than 0"),
			field.Invalid(env.garbageCollectionTTLFieldPath, "-1ns", "ttl must be a non-negative duration"),
			field.Invalid(env.httpProxyFieldPath, InvalidHttpProxyUrl, "url must start with 'http://' or 'https://'"),
			field.Invalid(env.httpProxyFieldPath, InvalidHttpProxyUrl, "subdomain must consist of lower case alphanumeric characters"),
			field.Invalid(env.httpsProxyFieldPath, InvalidHttpsProxyUrl, "url must start with 'http://' or 'https://'"),
			field.Invalid(env.httpsProxyFieldPath, InvalidHttpsProxyUrl, "subdomain must consist of lower case alphanumeric characters"),
		}, errs)
	})

	t.Run("uniqueness", func(t *testing.T) {
		cfg := buildConfig("config2", "default", registrycache.RegistryCacheConfigSpec{
			Upstream: "docker.io",
		})
		existing := buildConfig("config1", "test", registrycache.RegistryCacheConfigSpec{Upstream: "docker.io"})

		errs := NewValidator(env.dnsResolverAllOK, fixFakeClient(&existing)).Do(&cfg)
		validateResult(t, field.ErrorList{
			field.Duplicate(env.upstreamFieldPath, "docker.io"),
		}, errs)
	})

	t.Run("upstream not resolvable", func(t *testing.T) {
		cfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream: "some.incorrect.repo.io",
		})
		errs := NewValidator(env.dnsResolver, fixFakeClient()).Do(&cfg)
		validateResult(t, field.ErrorList{
			field.Invalid(env.upstreamFieldPath, "some.incorrect.repo.io", "upstream is not DNS resolvable"),
		}, errs)
	})

	t.Run("remoteURL not resolvable", func(t *testing.T) {
		cfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream:  "docker.io",
			RemoteURL: ptr.To("https://registry-not-existing.not-exists.io"),
		})
		errs := NewValidator(env.dnsResolver, fixFakeClient()).Do(&cfg)
		validateResult(t, field.ErrorList{
			field.Invalid(env.remoteURLFieldPath, ptr.To("https://registry-not-existing.not-exists.io"), "remoteURL is not DNS resolvable"),
		}, errs)
	})

	t.Run("secret validity", func(t *testing.T) {
		t.Run("non existent", func(t *testing.T) {
			cfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To("non-existent-secret"),
			})
			errs := NewValidator(env.dnsResolverAllOK, fixFakeClient()).Do(&cfg)
			validateResult(t, field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), "non-existent-secret", "secret does not exist"),
			}, errs)
		})
		t.Run("invalid structure", func(t *testing.T) {
			cfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(env.invalidSecret.Name),
			})
			errs := NewValidator(env.dnsResolverAllOK, fixFakeClient(&env.invalidSecret)).Do(&cfg)
			validateResult(t, field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), env.invalidSecret.Name, "two data entries"),
				field.Invalid(fieldPathSpec("secretReferenceName"), env.invalidSecret.Name, "missing \"username\" data entry"),
				field.Invalid(fieldPathSpec("secretReferenceName"), env.invalidSecret.Name, "missing \"password\" data entry"),
			}, errs)
		})
		t.Run("mutable secret", func(t *testing.T) {
			cfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(env.mutableSecret.Name),
			})
			errs := NewValidator(env.dnsResolverAllOK, fixFakeClient(&env.mutableSecret)).Do(&cfg)
			validateResult(t, field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), env.mutableSecret.Name, "should be immutable"),
			}, errs)
		})
	})
}

func TestDoOnUpdate(t *testing.T) {
	env := newTestEnv()

	t.Run("happy path", func(t *testing.T) {
		oldCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream:  "docker.io",
			RemoteURL: ptr.To("https://registry-1.docker.io"),
		})
		newCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream:            "docker.io",
			SecretReferenceName: ptr.To(env.validSecret.Name),
		})
		errs := NewValidator(env.dnsResolverAllOK, fixFakeClient(&env.validSecret, &oldCfg)).DoOnUpdate(&newCfg, &oldCfg)
		validateResult(t, field.ErrorList{}, errs)
	})

	t.Run("spec emptiness", func(t *testing.T) {
		oldCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream: "quay.io",
		})
		newCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{})
		errs := NewValidator(env.dnsResolverAllOK, fixFakeClient(&oldCfg)).DoOnUpdate(&newCfg, &oldCfg)
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
		errs := NewValidator(env.dnsResolverAllOK, fixFakeClient(&oldCfg)).DoOnUpdate(&newCfg, &oldCfg)
		validateResult(t, field.ErrorList{
			field.Invalid(env.volumeSizeFieldPath, NewVolumeSize, "field is immutable"),
			field.Invalid(env.volumeStorageClassNameFieldPath, ptr.To(NewStorageClassName), "field is immutable"),
			field.Invalid(env.garbageCollectionTTLFieldPath, &registrycacheext.GarbageCollection{
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

		config2 := buildConfig("config2", "test", registrycache.RegistryCacheConfigSpec{
			Upstream: "docker.io",
		})

		errs := NewValidator(env.dnsResolverAllOK, fixFakeClient(&oldCfg, &config2)).DoOnUpdate(&newCfg, &oldCfg)
		validateResult(t, field.ErrorList{
			field.Duplicate(fieldPathSpec("upstream"), "docker.io"),
		}, errs)
	})

	t.Run("upstream not resolvable", func(t *testing.T) {
		oldCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream: "quay.io",
		})
		newCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream: "some.incorrect.repo.io",
		})

		errs := NewValidator(env.dnsResolver, fixFakeClient(&oldCfg)).DoOnUpdate(&newCfg, &oldCfg)
		validateResult(t, field.ErrorList{
			field.Invalid(fieldPathSpec("upstream"), "some.incorrect.repo.io", "upstream is not DNS resolvable"),
		}, errs)
	})

	t.Run("remoteURL not resolvable", func(t *testing.T) {
		oldCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream: "quay.io",
		})
		newCfg := buildConfig("config1", "default", registrycache.RegistryCacheConfigSpec{
			Upstream:  "docker.io",
			RemoteURL: ptr.To("https://registry-not-existing.not-exists.io"),
		})
		errs := NewValidator(env.dnsResolver, fixFakeClient(&oldCfg)).DoOnUpdate(&newCfg, &oldCfg)
		validateResult(t, field.ErrorList{
			field.Invalid(fieldPathSpec("remoteURL"), ptr.To("https://registry-not-existing.not-exists.io"), "remoteURL is not DNS resolvable"),
		}, errs)
	})

	t.Run("secret validity", func(t *testing.T) {
		t.Run("invalid structure", func(t *testing.T) {
			oldCfg := buildConfig("config-with-invalid-secret", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(env.validSecret.Name),
			})
			newCfg := buildConfig("config-with-invalid-secret", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(env.invalidSecret.Name),
			})
			errs := NewValidator(env.dnsResolverAllOK, fixFakeClient(&env.invalidSecret, &oldCfg)).DoOnUpdate(&newCfg, &oldCfg)
			validateResult(t, field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), env.invalidSecret.Name, "two data entries"),
				field.Invalid(fieldPathSpec("secretReferenceName"), env.invalidSecret.Name, "missing \"username\" data entry"),
				field.Invalid(fieldPathSpec("secretReferenceName"), env.invalidSecret.Name, "missing \"password\" data entry"),
			}, errs)
		})
		t.Run("mutable secret", func(t *testing.T) {
			oldCfg := buildConfig("config-with-invalid-secret", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(env.validSecret.Name),
			})
			newCfg := buildConfig("config-with-invalid-secret", "default", registrycache.RegistryCacheConfigSpec{
				Upstream:            "docker.io",
				SecretReferenceName: ptr.To(env.mutableSecret.Name),
			})
			errs := NewValidator(env.dnsResolverAllOK, fixFakeClient(&env.validSecret, &env.mutableSecret)).DoOnUpdate(&newCfg, &oldCfg)
			validateResult(t, field.ErrorList{
				field.Invalid(fieldPathSpec("secretReferenceName"), env.mutableSecret.Name, "should be immutable"),
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

func fixFakeClient(initObjs ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	_ = registrycache.AddToScheme(scheme)
	_ = v1.AddToScheme(scheme)

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(initObjs...).Build()
}
