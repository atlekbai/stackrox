package singleton

import (
	"time"

	"github.com/stackrox/rox/pkg/buildinfo"
	"github.com/stackrox/rox/pkg/license/validator"
	"github.com/stackrox/rox/pkg/utils"
)

func init() {
	utils.Must(
		validatorInstance.RegisterSigningKey(
			validator.EC256,
			// projects/stackrox-dev/locations/global/keyRings/licensing-qa/cryptoKeys/qa-license-signer/cryptoKeyVersions/1
			[]byte{
				0x30, 0x59, 0x30, 0x13, 0x06, 0x07, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x02,
				0x01, 0x06, 0x08, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x03, 0x01, 0x07, 0x03,
				0x42, 0x00, 0x04, 0xcd, 0xbd, 0x69, 0x82, 0xd5, 0xd5, 0x0d, 0x0e, 0xef,
				0x75, 0xe8, 0x30, 0x7c, 0x26, 0x2b, 0x8b, 0xea, 0xfe, 0x94, 0x3a, 0xee,
				0xcd, 0xf1, 0xc5, 0x50, 0xfb, 0xe2, 0x27, 0x67, 0xf3, 0x76, 0x17, 0x52,
				0x85, 0xe0, 0x1d, 0x92, 0xc8, 0xa9, 0xb1, 0xf8, 0x95, 0x62, 0x1c, 0xdd,
				0x91, 0x04, 0x3d, 0x6f, 0x15, 0x2c, 0x8a, 0xee, 0x14, 0xd5, 0xca, 0xde,
				0xbe, 0x82, 0xbe, 0x67, 0xab, 0xc0, 0xbd,
			},
			validator.SigningKeyRestrictions{
				EarliestNotValidBefore:        buildinfo.BuildTimestamp().Add(-7 * 24 * time.Hour),
				LatestNotValidAfter:           buildinfo.BuildTimestamp().Add(30 * 24 * time.Hour),
				MaxDuration:                   16 * 24 * time.Hour,
				AllowOffline:                  true,
				AllowNoNodeLimit:              true,
				AllowNoBuildFlavorRestriction: true,
				DeploymentEnvironments: []string{
					"gcp/ultra-current-825",
					"azure/66c57ff5-f49f-4510-ae04-e26d3ad2ee63",
					"aws/051999192406",
					"aws/880732477823", // k@stackrox.com"
					"aws/405534842425", // ll@stackrox.com
				},
			}),
	)
}
