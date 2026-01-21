package objectlock

import "errors"

var (
	// ErrInvalidObjectLockState indicates an unsupported object lock state.
	ErrInvalidObjectLockState = errors.New("objectlock: invalid object lock state")

	// ErrInvalidRetentionMode indicates an unsupported retention mode.
	ErrInvalidRetentionMode = errors.New("objectlock: invalid retention mode")

	// ErrInvalidRetentionPeriod indicates an invalid retention period configuration.
	ErrInvalidRetentionPeriod = errors.New("objectlock: retention must specify either days or years")

	// ErrInvalidLegalHoldStatus indicates an unsupported legal hold status.
	ErrInvalidLegalHoldStatus = errors.New("objectlock: invalid legal hold status")

	// ErrInvalidRetentionDate indicates a missing retain-until date.
	ErrInvalidRetentionDate = errors.New("objectlock: retain-until date must be set")

	// ErrNoObjectLockConfig indicates that no object lock configuration exists.
	ErrNoObjectLockConfig = errors.New("objectlock: bucket has no object lock configuration")
)
