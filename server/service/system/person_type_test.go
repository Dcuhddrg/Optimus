package system

import (
	"context"
	"strconv"
	"testing"
	"time"

	"optimus/server/global"
	projectModel "optimus/server/model/project"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"

	"github.com/stretchr/testify/require"
)

func TestPersonTypeClassification(t *testing.T) {
	cfg := systemModel.SystemLDAPConfig{ExternalTypeValues: "contractor, Vendor"}
	require.Equal(t, PersonExternal, personTypeOf(LDAPUser{EmployeeTypes: []string{"staff", "CONTRACTOR"}}, cfg), "any value, case-insensitive")
	require.Equal(t, PersonExternal, personTypeOf(LDAPUser{EmployeeTypes: []string{"vendor"}}, cfg))
	require.Equal(t, PersonEmployee, personTypeOf(LDAPUser{EmployeeTypes: []string{"staff"}}, cfg))
	require.Equal(t, PersonEmployee, personTypeOf(LDAPUser{}, cfg), "no employeeType: employee (known limitation)")
	require.Equal(t, PersonEmployee, personTypeOf(LDAPUser{EmployeeTypes: []string{"contractor"}}, systemModel.SystemLDAPConfig{}), "empty list: nobody is external")
	require.Equal(t, "contractor,Vendor", joinTypeValues([]string{" contractor ", "", "Vendor", "vendor"}))
}

func TestEffectiveRolesForExternalAccounts(t *testing.T) {
	expires := time.Now().Add(time.Hour)
	sponsor := uint(9)
	roles := []systemModel.SystemRoleSummary{{ID: 1, Code: AdminRoleCode}, {ID: 2, Code: "security-approver", Approvable: true}, {ID: 3, Code: "ops"}}

	employee := systemModel.SystemUser{PersonType: PersonEmployee}
	require.Len(t, effectiveRoles(employee, roles), 3)
	require.Len(t, effectiveRoles(systemModel.SystemUser{}, roles), 3, "unset column is an employee")

	external := systemModel.SystemUser{PersonType: PersonExternal, ExpiresAt: &expires}
	require.Empty(t, effectiveRoles(external, roles), "no sponsor: no roles")
	external = systemModel.SystemUser{PersonType: PersonExternal, SponsorID: &sponsor}
	require.Empty(t, effectiveRoles(external, roles), "no expiry: no roles")
	external.ExpiresAt = &expires
	require.Equal(t, []systemModel.SystemRoleSummary{{ID: 3, Code: "ops"}}, effectiveRoles(external, roles), "custom roles only")
}

func TestExpiryIsTheEndOfTheLastValidDay(t *testing.T) {
	end, err := endOfDay("2026-12-31")
	require.NoError(t, err)
	day := time.Date(2026, 12, 31, 0, 0, 0, 0, time.Local)
	require.False(t, accountExpired(PersonExternal, &end, day.Add(23*time.Hour+59*time.Minute)))
	require.True(t, accountExpired(PersonExternal, &end, day.AddDate(0, 0, 1)))
	require.False(t, accountExpired(PersonEmployee, &end, day.AddDate(1, 0, 0)), "employees never expire")
	require.Equal(t, "2026-12-31", *expiresOn(&end))
	_, err = endOfDay("31/12/2026")
	require.Error(t, err)
}

func externalEntry(username, id string) LDAPUser {
	e := dirEntry(username, id)
	e.EmployeeTypes = []string{"contractor"}
	return e
}

// A new external account gets no role; turning an employee external through
// login updates the person type and its authorization at once.
func TestLDAPLoginWritesPersonType(t *testing.T) {
	db := openTestDB(t)
	enableLDAP(t, db)
	seedDefaultRole(t, db)
	dir := newFakeDirectory(externalEntry("eve", "id-eve"), dirEntry("dan", "id-dan"))

	eve, failure := syncAs(t, dir, "eve")
	require.Nil(t, failure)
	require.Equal(t, PersonExternal, eve.PersonType)
	roles, err := queryUserRoles(context.Background(), eve.ID)
	require.NoError(t, err)
	require.Empty(t, roles, "external accounts are created without the user role")

	dan, failure := syncAs(t, dir, "dan")
	require.Nil(t, failure)
	require.Equal(t, PersonEmployee, dan.PersonType)
	require.NotEmpty(t, loadState(t, dan.ID).RoleIDs)

	dir.set(externalEntry("dan", "id-dan"))
	dan, failure = syncAs(t, dir, "dan")
	require.Nil(t, failure)
	require.Equal(t, PersonExternal, dan.PersonType)
	state := loadState(t, dan.ID)
	require.Equal(t, PersonExternal, state.PersonType)
	require.Empty(t, state.RoleIDs, "cache invalidated: no roles until expiry and sponsor are set")
}

func newExternalUser(t *testing.T, username string) systemModel.SystemUser {
	t.Helper()
	user := systemModel.SystemUser{UUID: "u-" + username, Username: username, Source: "ldap", Status: DefaultUserStatus, PersonType: PersonExternal}
	require.NoError(t, global.OPTIMUS_DB.Create(&user).Error)
	return user
}

func TestUpdateUserExternalSetsControlsAndRestoresRoles(t *testing.T) {
	db := openTestDB(t)
	svc := &UserService{}
	ctx := context.Background()
	sponsor := systemModel.SystemUser{UUID: "u-sam", Username: "sam", Source: "ldap", Status: DefaultUserStatus, PersonType: PersonEmployee}
	require.NoError(t, db.Create(&sponsor).Error)
	disabled := systemModel.SystemUser{UUID: "u-old", Username: "old", Source: "ldap", Status: "disabled", PersonType: PersonEmployee}
	require.NoError(t, db.Create(&disabled).Error)
	eve := newExternalUser(t, "eve")
	ops := systemModel.SystemRole{Code: "ops", Name: "ops"}
	require.NoError(t, db.Create(&ops).Error)
	require.NoError(t, db.Create(&systemModel.SystemUserRole{UserID: eve.ID, RoleID: ops.ID}).Error)
	require.Empty(t, loadState(t, eve.ID).RoleIDs)

	day := "2099-06-30"
	_, err := svc.UpdateUserExternal(ctx, 1, eve.ID, systemRequest.UpdateUserExternalRequest{ExpiresOn: &day, SponsorID: &disabled.ID}, AuditMeta{})
	require.ErrorIs(t, err, ErrInvalidSponsor)
	bad := "30/06/2099"
	_, err = svc.UpdateUserExternal(ctx, 1, eve.ID, systemRequest.UpdateUserExternalRequest{ExpiresOn: &bad, SponsorID: &sponsor.ID}, AuditMeta{})
	require.ErrorIs(t, err, ErrInvalidExpiry)
	_, err = svc.UpdateUserExternal(ctx, 1, sponsor.ID, systemRequest.UpdateUserExternalRequest{ExpiresOn: &day, SponsorID: &sponsor.ID}, AuditMeta{})
	require.ErrorIs(t, err, ErrNotExternal)

	// Only one of the two: still no roles.
	summary, err := svc.UpdateUserExternal(ctx, 1, eve.ID, systemRequest.UpdateUserExternalRequest{ExpiresOn: &day}, AuditMeta{})
	require.NoError(t, err)
	require.Equal(t, day, *summary.ExpiresOn)
	require.Empty(t, loadState(t, eve.ID).RoleIDs)

	_, err = svc.UpdateUserExternal(ctx, 1, eve.ID, systemRequest.UpdateUserExternalRequest{ExpiresOn: &day, SponsorID: &sponsor.ID}, AuditMeta{})
	require.NoError(t, err)
	require.Equal(t, []uint{ops.ID}, loadState(t, eve.ID).RoleIDs, "both set: roles apply on the next request")

	past := "2020-01-01"
	_, err = svc.UpdateUserExternal(ctx, 1, eve.ID, systemRequest.UpdateUserExternalRequest{ExpiresOn: &past, SponsorID: &sponsor.ID}, AuditMeta{})
	require.NoError(t, err)
	state := loadState(t, eve.ID)
	require.True(t, state.Expired(time.Now()))
	require.Equal(t, DefaultUserStatus, state.Status, "expiry does not change the status")

	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Where("action = ? AND result = ?", "user.external.update", "success").Last(&audit).Error)
	require.Contains(t, audit.Detail, "expiresOn")

	list, err := svc.ListUsers(ctx)
	require.NoError(t, err)
	for _, item := range list.Items {
		if item.ID == eve.ID {
			require.True(t, item.Expired)
			require.Equal(t, "sam", item.Sponsor)
			require.False(t, item.SponsorInvalid)
		}
	}
	require.NoError(t, db.Model(&sponsor).Update("status", "disabled").Error)
	list, err = svc.ListUsers(ctx)
	require.NoError(t, err)
	for _, item := range list.Items {
		if item.ID == eve.ID {
			require.True(t, item.SponsorInvalid, "a disabled sponsor is flagged, the account is untouched")
		}
	}
}

func TestExternalAccountsCannotHoldAdminOrApprovalRoles(t *testing.T) {
	db := openTestDB(t)
	admin := systemModel.SystemRole{Code: AdminRoleCode, Name: "admin"}
	approver := systemModel.SystemRole{Code: "security-approver", Name: "sec", Approvable: true}
	custom := systemModel.SystemRole{Code: "user-manager", Name: "um"}
	for _, r := range []*systemModel.SystemRole{&admin, &approver, &custom} {
		require.NoError(t, db.Create(r).Error)
	}
	eve := newExternalUser(t, "eve")
	svc := &UserService{}
	ctx := context.Background()
	require.ErrorIs(t, svc.UpdateUserRoles(ctx, 1, eve.ID, []uint{admin.ID}, AuditMeta{}), ErrExternalRoleRestricted)
	require.ErrorIs(t, svc.UpdateUserRoles(ctx, 1, eve.ID, []uint{custom.ID, approver.ID}, AuditMeta{}), ErrExternalRoleRestricted)
	require.NoError(t, svc.UpdateUserRoles(ctx, 1, eve.ID, []uint{custom.ID}, AuditMeta{}), "custom roles are the admin's call")
}

// Expired accounts are refused at login with their own message.
func TestLDAPLoginRefusesExpiredExternalAccount(t *testing.T) {
	db := openTestDB(t)
	enableLDAP(t, db)
	seedDefaultRole(t, db)
	newLDAPServiceForTest(t, newFakeDirectory(externalEntry("eve", "id-eve")))
	past := time.Now().Add(-time.Hour)
	sponsor := uint(1)
	eve := systemModel.SystemUser{UUID: "u-eve", Username: "eve", Source: "ldap", Status: DefaultUserStatus, PersonType: PersonExternal, ExternalID: strp("id-eve"), ExpiresAt: &past, SponsorID: &sponsor}
	require.NoError(t, db.Create(&eve).Error)
	_, err := (&AuthService{}).Login(context.Background(), systemRequest.LoginRequest{Username: "eve", Password: "pw"}, AuditMeta{})
	require.ErrorIs(t, err, ErrAccountExpired)
	require.Equal(t, "account_expired", lastLoginReason(t, db))
}

// The sync preview lists an employee turning external with the roles that
// stop applying and the projects they still own; execution writes the type.
func TestLDAPSyncFlagsEmployeeTurningExternal(t *testing.T) {
	f := newSyncFixture(t, externalEntry("dan", "id-dan"), externalEntry("newbie", "id-newbie"))
	dan := f.account(t, "dan", strp("id-dan"))
	require.NoError(t, f.db.Model(&dan).Updates(map[string]any{"display_name": "DAN", "email": "dan@optimus.test", "department": "ops"}).Error)
	admin := systemModel.SystemRole{Code: AdminRoleCode, Name: "admin"}
	require.NoError(t, f.db.Create(&admin).Error)
	require.NoError(t, f.db.Create(&systemModel.SystemUserRole{UserID: dan.ID, RoleID: admin.ID}).Error)
	project := projectModel.Project{Name: "Ops tools", Code: "ops-tools-" + strconv.FormatInt(time.Now().UnixNano(), 36), OwnerID: dan.ID, Status: "active"}
	require.NoError(t, f.db.Create(&project).Error)
	t.Cleanup(func() { f.db.Unscoped().Delete(&project) })
	require.NoError(t, f.db.Create(&projectModel.UserProject{UserID: dan.ID, ProjectID: project.ID, ProjectRole: "owner"}).Error)

	preview := f.preview(t)
	var found bool
	for _, item := range preview.Items {
		if item.Account != nil && item.Account.ID == dan.ID {
			found = true
			require.Equal(t, "update", item.Category)
			require.Equal(t, []string{AdminRoleCode}, item.RestrictedRoles)
			require.Equal(t, []string{project.Code}, item.OwnedProjects)
		}
	}
	require.True(t, found)
	f.execute(t, systemRequest.LDAPSyncExecuteRequest{PreviewID: preview.PreviewID})
	require.Equal(t, PersonExternal, f.reload(t, dan.ID).PersonType)
	require.Empty(t, loadState(t, dan.ID).RoleIDs, "admin role ignored at once")

	var newbie systemModel.SystemUser
	require.NoError(t, f.db.Where("username = ?", "newbie").First(&newbie).Error)
	require.Equal(t, PersonExternal, newbie.PersonType)
	roles, err := queryUserRoles(context.Background(), newbie.ID)
	require.NoError(t, err)
	require.Empty(t, roles, "external accounts created by sync get no role")
}

// The person type settings default on first save and keep an explicit empty
// list ("nobody is external").
func TestLDAPConfigPersonTypeSettings(t *testing.T) {
	openTestDB(t)
	svc := newLDAPServiceForTest(t, newFakeDirectory())
	shown, err := svc.GetConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, "employeeType", shown.EmployeeTypeAttr)
	require.Equal(t, []string{"contractor"}, shown.ExternalTypeValues)

	req := validLDAPRequest()
	empty := []string{}
	req.ExternalTypeValues = &empty
	saved, err := svc.SaveConfig(context.Background(), 1, req, AuditMeta{})
	require.NoError(t, err)
	require.Equal(t, "employeeType", saved.EmployeeTypeAttr)
	require.Empty(t, saved.ExternalTypeValues)
	cfg, err := svc.getLDAPConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, "", cfg.ExternalTypeValues)

	values := []string{"contractor", " vendor "}
	req.ExternalTypeValues = &values
	saved, err = svc.SaveConfig(context.Background(), 1, req, AuditMeta{})
	require.NoError(t, err)
	require.Equal(t, []string{"contractor", "vendor"}, saved.ExternalTypeValues)
}
