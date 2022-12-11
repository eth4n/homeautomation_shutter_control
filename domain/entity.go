package domain

type Entity interface {
	GetRawId() string
	GetUniqueId() string
	UpdateState(state *string)
	Subscribe()
	UnSubscribe()
	GetAppState() *State
	SetAppState(appState *State)
}
