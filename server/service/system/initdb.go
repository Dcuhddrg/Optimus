package system

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"go.uber.org/zap"
)

var (
	ErrMissingDBContext        = errors.New("missing db in context")
	ErrMissingDependentContext = errors.New("missing dependent value in context")
)

// SubInitializer is implemented by source/* packages, as GVA: each initializer
// migrates its tables and writes its seed data.
type SubInitializer interface {
	InitializerName() string
	MigrateTable(ctx context.Context) (next context.Context, err error)
	InitializeData(ctx context.Context) (next context.Context, err error)
	TableCreated(ctx context.Context) bool
	DataInserted(ctx context.Context) bool
}

type orderedInitializer struct {
	order int
	SubInitializer
}

type initSlice []*orderedInitializer

var (
	initializers initSlice
	cache        map[string]*orderedInitializer
)

// RegisterInit registers an initializer from a source package init(), as GVA.
func RegisterInit(order int, i SubInitializer) {
	if initializers == nil {
		initializers = initSlice{}
	}
	if cache == nil {
		cache = map[string]*orderedInitializer{}
	}
	name := i.InitializerName()
	if _, existed := cache[name]; existed {
		panic(fmt.Sprintf("Name conflict on %s", name))
	}
	ni := orderedInitializer{order, i}
	initializers = append(initializers, &ni)
	cache[name] = &ni
}

type InitDBService struct{}

var InitDBServiceApp = new(InitDBService)

// InitDB runs registered initializers in order at every startup. GVA migrates
// tables every start (RegisterTables) but seeds only once, through its web
// database init. Optimus has no web init and its idempotent seeds also carry
// upgrades (#62 Casbin grants, cloud project backfill), so it seeds every start
// unless DataInserted reports done. A migration failure stops the run (GVA
// exits); a seed failure is recorded and later initializers still run (#51).
func (initDBService *InitDBService) InitDB(ctx context.Context) error {
	if len(initializers) == 0 {
		return errors.New("no initializers registered")
	}
	sort.Sort(&initializers)
	var errs []error
	next := ctx
	for _, init := range initializers {
		n, err := init.MigrateTable(next)
		if err != nil {
			errs = append(errs, logInitFailure(init.InitializerName(), "migrate table", err))
			return errors.Join(errs...)
		}
		next = n
	}
	for _, init := range initializers {
		if init.DataInserted(next) {
			continue
		}
		n, err := init.InitializeData(next)
		if err != nil {
			errs = append(errs, logInitFailure(init.InitializerName(), "initialize data", err))
			continue
		}
		next = n
	}
	return errors.Join(errs...)
}

func logInitFailure(name, action string, err error) error {
	zap.L().Error(name+" init failed", zap.String("action", action), zap.Error(err))
	return fmt.Errorf("%s %s: %w", name, action, err)
}

func (a initSlice) Len() int           { return len(a) }
func (a initSlice) Less(i, j int) bool { return a[i].order < a[j].order }
func (a initSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
