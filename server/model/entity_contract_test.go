package model_test

import (
	"encoding/json"
	"flag"
	"os"
	"sort"
	"sync"
	"testing"

	cloud "optimus/server/model/cloud"
	project "optimus/server/model/project"
	system "optimus/server/model/system"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm/schema"
)

var entityContractOutput = flag.String("entity-contract-output", "", "record pre-migration schema/JSON contract at an explicit path")

// Capture observable fields and SQL metadata, excluding Go embedding paths and
// field declaration order, which the structural migration intentionally changes.
func TestEntitySchemaAndJSONContracts(t *testing.T) {
	entities := []any{system.SystemUser{}, system.SystemRole{}, system.SystemPermission{}, system.SystemLDAPConfig{}, system.SystemUserRole{}, system.SystemRolePermission{}, system.SystemAuditLog{}, system.SystemDataMigration{}, project.Project{}, project.UserProject{}, cloud.CloudAccount{}, cloud.CloudAsset{}}
	actual := map[string]any{}
	for _, entity := range entities {
		s, err := schema.Parse(entity, &sync.Map{}, schema.NamingStrategy{})
		require.NoError(t, err)
		encoded, err := json.Marshal(entity)
		require.NoError(t, err)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(encoded, &payload))
		fields := map[string]any{}
		for _, f := range s.Fields {
			if f.DBName == "" {
				continue
			}
			fields[f.DBName] = map[string]any{"type": f.FieldType.String(), "dataType": f.DataType, "pk": f.PrimaryKey, "autoIncrement": f.AutoIncrement, "notNull": f.NotNull, "unique": f.Unique, "hasDefault": f.HasDefaultValue, "default": f.DefaultValue, "size": f.Size, "precision": f.Precision, "scale": f.Scale, "autoCreate": f.AutoCreateTime, "autoUpdate": f.AutoUpdateTime}
		}
		indexes := map[string]any{}
		for _, index := range s.ParseIndexes() {
			cols := []string{}
			for _, f := range index.Fields {
				cols = append(cols, f.DBName)
			}
			indexes[index.Name] = map[string]any{"class": index.Class, "type": index.Type, "where": index.Where, "option": index.Option, "fields": cols}
		}
		pk := append([]string{}, s.PrimaryFieldDBNames...)
		sort.Strings(pk)
		actual[s.Table] = map[string]any{"json": payload, "fields": fields, "indexes": indexes, "primaryKeys": pk, "deleteClauses": len(s.DeleteClauses)}
	}
	snapshot, err := json.MarshalIndent(actual, "", "  ")
	require.NoError(t, err)
	if *entityContractOutput != "" {
		require.NoError(t, os.WriteFile(*entityContractOutput, snapshot, 0600))
		return
	}
	expected, err := os.ReadFile("testdata/entities.json")
	require.NoError(t, err)
	require.JSONEq(t, string(expected), string(snapshot))
}
