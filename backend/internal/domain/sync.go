package domain

type SyncNodeResult struct {
	NodeId   int64
	NodeName string
	Added    int
	Removed  int
	Success  bool
	Error    string
}
