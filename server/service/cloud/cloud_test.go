package cloud

import (
	"testing"

	cloudModel "optimus/server/model/cloud"

	"github.com/stretchr/testify/require"
)

func TestCloudAccountCarriesProjectID(t *testing.T) {
	account := cloudModel.CloudAccount{ProjectID: 7}

	require.Equal(t, uint(7), account.ProjectID)
}

func TestBackfillCloudProjectScopeRejectsNilDB(t *testing.T) {
	require.Error(t, BackfillCloudProjectScope(nil))
}
