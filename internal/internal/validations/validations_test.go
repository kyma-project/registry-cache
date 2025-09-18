package validations

import (
	registrycache "github.com/kyma-project/registry-cache/api/v1beta1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
	"slices"
	"strings"
	"testing"
	"time"
)

const (
	InvalidUpstreamPort           = "docker.io:77777"
	InvalidRemoteURL              = "docker.io"
	InvalidVolumeSize             = "-1"
	InvalidVolumeStorageClassName = "Invalid.Name"
	InvalidGarbageCollectionValue = -1
	InvalidHttpProxyUrl           = "http//invalid-url"
	InvalidHttpsProxyUrl          = "https//invalid-url"

	NewVolumeSize             = "20Gi"
	NewStorageClassName       = "nonstandard"
	NewGarbageCollectionValue = 1 * time.Hour
)

func TestDo(t *testing.T) {

	upstreamFieldPath := field.NewPath("spec").Child("upstream")
	remoteURLFieldPath := field.NewPath("spec").Child("remoteURL")
	volumeSizeFieldPath := field.NewPath("spec").Child("volume").Child("size")
	volumeStorageClassNameFieldPath := field.NewPath("spec").Child("volume").Child("storageClassName")
	garbageCollectionTTLFieldPath := field.NewPath("spec").Child("garbageCollection").Child("ttl")
	httpProxyFieldPath := field.NewPath("spec").Child("volume").Child("proxy").Child("httpProxy")
	httpsProxyFieldPath := field.NewPath("spec").Child("volume").Child("proxy").Child("httpsProxy")

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
			"user":     []byte("dXNlcg=="),
			"password": []byte("cGFzc3dvcmQ="),
		},
		Immutable: ptr.To(false),
	}

	for _, tt := range []struct {
		name string
		registrycache.RegistryCacheConfig
		existingConfigs []registrycache.RegistryCacheConfig
		errorsList      field.ErrorList
		secrets         []v1.Secret
	}{
		{
			name: "valid spec",
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream: "docker.io",
				},
			},
			errorsList: field.ErrorList{},
		},
		{
			name: "empty spec",
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{},
			},
			errorsList: field.ErrorList{
				field.Required(field.NewPath("spec"), "spec cannot be empty"),
			},
		},
		{
			name: "invalid spec",
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:  InvalidUpstreamPort,
					RemoteURL: ptr.To(InvalidUpstreamPort),
					Volume: &registrycache.Volume{
						Size:             ptr.To(resource.MustParse(InvalidVolumeSize)),
						StorageClassName: ptr.To(InvalidVolumeStorageClassName),
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
				field.Invalid(volumeStorageClassNameFieldPath, InvalidVolumeStorageClassName, "an RFC 1123 subdomain must consist of alphanumeric characters"),
				field.Invalid(garbageCollectionTTLFieldPath, InvalidGarbageCollectionValue, "ttl must be a non-negative duration"),
				field.Invalid(httpProxyFieldPath, InvalidHttpProxyUrl, "some error"),
				field.Invalid(httpsProxyFieldPath, InvalidHttpsProxyUrl, "some error"),
			},
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
				field.Invalid(field.NewPath("spec").Child("upstream"), "docker.io", "duplicated upstream"),
			},
		},
		{
			name: "upstream non-resolvable",
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
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
				field.Invalid(field.NewPath("spec").Child("upstream"), "docker.io", "duplicated upstream"),
			},
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
				field.NotFound(field.NewPath("spec").Child("secretReferenceName"), "non-existent-secret"),
			},
		},
		{
			name:    "secret with incorrect structure",
			secrets: []v1.Secret{secretWithIncorrectStructure},
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:            "docker.io",
					SecretReferenceName: ptr.To(secretWithIncorrectStructure.Name),
				},
			},
			errorsList: field.ErrorList{
				field.Invalid(field.NewPath("spec").Child("secretReferenceName"), secretWithIncorrectStructure.Name, "invalid secret reference"),
			},
		},
		{
			name: "mutable secret",
			secrets: []v1.Secret{
				mutableSecret,
			},
			RegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream:            "docker.io",
					SecretReferenceName: ptr.To(mutableSecret.Name),
				},
			},
			errorsList: field.ErrorList{
				field.Invalid(field.NewPath("spec").Child("secretReferenceName"), mutableSecret.Name, "should be immutable"),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			errs := NewValidator(tt.secrets, tt.existingConfigs).Do(&tt.RegistryCacheConfig)

			validateResult(t, errs, tt.errorsList)
		})
	}
}

func TestDoOnUpdate(t *testing.T) {

	volumeSizeFieldPath := field.NewPath("spec").Child("volume").Child("size")
	volumeStorageClassNameFieldPath := field.NewPath("spec").Child("volume").Child("storageClassName")
	garbageCollectionTTLFieldPath := field.NewPath("spec").Child("garbageCollection").Child("ttl")

	for _, tt := range []struct {
		name                   string
		newRegistryCacheConfig registrycache.RegistryCacheConfig
		oldRegistryCacheConfig registrycache.RegistryCacheConfig
		existingConfigs        []registrycache.RegistryCacheConfig
		errorsList             field.ErrorList
		secrets                []v1.Secret
	}{
		{
			name: "valid spec",
			newRegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream: "docker.io",
				},
			},
			oldRegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Upstream: "docker.io",
				},
			},
			errorsList: field.ErrorList{},
		},
		{
			name: "invalid spec",
			newRegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Volume: &registrycache.Volume{
						Size:             ptr.To(resource.MustParse("10Gi")),
						StorageClassName: ptr.To("standard"),
					},
					GarbageCollection: &registrycache.GarbageCollection{
						TTL: metav1.Duration{Duration: 0},
					},
				},
			},
			oldRegistryCacheConfig: registrycache.RegistryCacheConfig{
				Spec: registrycache.RegistryCacheConfigSpec{
					Volume: &registrycache.Volume{
						Size:             ptr.To(resource.MustParse(NewVolumeSize)),
						StorageClassName: ptr.To(NewStorageClassName),
					},
					GarbageCollection: &registrycache.GarbageCollection{
						TTL: metav1.Duration{Duration: 1 * time.Hour},
					},
				},
			},
			errorsList: field.ErrorList{
				field.Invalid(volumeSizeFieldPath, NewVolumeSize, "field is immutable"),
				field.Invalid(volumeStorageClassNameFieldPath, NewStorageClassName, "field is immutable"),
				field.Invalid(garbageCollectionTTLFieldPath, NewGarbageCollectionValue, "garbage collection cannot be enabled"),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			errs := NewValidator(tt.secrets, tt.existingConfigs).DoOnUpdate(&tt.newRegistryCacheConfig, &tt.oldRegistryCacheConfig)

			validateResult(t, tt.errorsList, errs)
		})
	}
}

func validateResult(t *testing.T, expectedErrors field.ErrorList, actualErrors field.ErrorList) bool {
	require.Equal(t, len(expectedErrors), len(actualErrors))

	for _, expectedErr := range expectedErrors {
		actualErrIndex := slices.IndexFunc(actualErrors, func(err *field.Error) bool {
			return err.Type == expectedErr.Type && expectedErr.Field == err.Field
		})
		require.NotEqual(t, -1, actualErrIndex, "actual error not found: %v", expectedErr)

		actualFieldError := actualErrors[actualErrIndex]
		require.Equal(t, expectedErr.BadValue, actualFieldError.BadValue)
		require.True(t, strings.Contains(actualFieldError.Detail, expectedErr.Detail))
	}
}
