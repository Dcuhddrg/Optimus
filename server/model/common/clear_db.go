package common

// ClearDB describes rows to delete periodically: rows whose CompareField is
// older than Interval, as GVA's model/common.ClearDB.
type ClearDB struct {
	TableName    string
	CompareField string
	Interval     string
}
