package cloud

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBackfillCloudProjectScopeRejectsNilDB(t *testing.T) {
	require.Error(t, BackfillCloudProjectScope(nil))
}

func TestSeedCloudDataRejectsNilDB(t *testing.T) {
	require.Error(t, SeedCloudData(nil))
}
