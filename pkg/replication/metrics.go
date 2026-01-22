// Package replication provides replication metrics types.
package replication

// TargetMetrics represents replication metrics for a single target.
type TargetMetrics struct {
	ReplicatedCount                  uint64        `json:"replicationCount,omitempty"`
	ReplicatedSize                   uint64        `json:"completedReplicationSize,omitempty"`
	BandWidthLimitInBytesPerSecond   int64         `json:"limitInBits,omitempty"`
	CurrentBandwidthInBytesPerSecond float64       `json:"currentBandwidth,omitempty"`
	Failed                           TimedErrStats `json:"failed,omitempty"`
	PendingSize                      uint64        `json:"pendingReplicationSize,omitempty"`
	ReplicaSize                      uint64        `json:"replicaSize,omitempty"`
	FailedSize                       uint64        `json:"failedReplicationSize,omitempty"`
	PendingCount                     uint64        `json:"pendingReplicationCount,omitempty"`
	FailedCount                      uint64        `json:"failedReplicationCount,omitempty"`
}

// Metrics represents replication metrics for a bucket.
type Metrics struct {
	Stats           map[string]TargetMetrics `json:"Stats,omitempty"`
	ReplicatedSize  uint64                   `json:"completedReplicationSize,omitempty"`
	ReplicaSize     uint64                   `json:"replicaSize,omitempty"`
	ReplicaCount    int64                    `json:"replicaCount,omitempty"`
	ReplicatedCount int64                    `json:"replicationCount,omitempty"`
	Errors          TimedErrStats            `json:"failed,omitempty"`
	PendingSize     uint64                   `json:"pendingReplicationSize,omitempty"`
	FailedSize      uint64                   `json:"failedReplicationSize,omitempty"`
	PendingCount    uint64                   `json:"pendingReplicationCount,omitempty"`
	FailedCount     uint64                   `json:"failedReplicationCount,omitempty"`
}

// RStat holds count and bytes for replication metrics.
type RStat struct {
	Count float64 `json:"count"`
	Bytes int64   `json:"bytes"`
}

// TimedErrStats holds error stats for a time period.
type TimedErrStats struct {
	LastMinute RStat `json:"lastMinute"`
	LastHour   RStat `json:"lastHour"`
	Totals     RStat `json:"totals"`
}
