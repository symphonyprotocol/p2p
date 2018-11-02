package models

type SyncType int

const (
	SYNC_TYPE_BLOCK	SyncType = iota
	SYNC_TYPE_FILE
	SYNC_TYPE_CONFIG
	SYNC_TYPE_RAW

	// maybe more.
)

type SyncStrategy int

const (
	SYNC_STRATEGY_CLOSER SyncStrategy = iota
	SYNC_STRATEGY_ALL
	
	
)

