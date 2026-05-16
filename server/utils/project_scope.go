package utils

const ProjectScopeContextKey = "project_scope"

type ProjectScope struct {
	CurrentProjectID     uint
	CurrentProjectCode   string
	CurrentProjectRole   string
	AccessibleProjectIDs []uint
}
