package migration

// Manager is an interface that describe the common behavior of a migration
// manager.
//
// Any manager is divider in two parts: the `Source` and `Target`. In short, the
// `Source` is the origin of the migrations (eg. SQL files or Go scripts). A
// `Target` is database technology that the migrations will be action on.
//
// Integrating the `Source` and the `Target`, the `Manager` is responsible for
// running migrations with its methods `Migrate`, `Rewind`, `Reset`, `Up` and
// `Down`.
type Manager interface {
	Source() Source
	Target() Target
	MigrationsExecuted() ([]Migration, error)
	MigrationsPending() ([]Migration, error)
	Migrate(listener Reporter) ([]*MigrationSummary, error)
	Rewind(listener Reporter) ([]*MigrationSummary, error)
	Reset(listener Reporter) ([]*MigrationSummary, []*MigrationSummary, error)
	Do(listener Reporter) (*MigrationSummary, error)
	Undo(listener Reporter) (*MigrationSummary, error)
}
