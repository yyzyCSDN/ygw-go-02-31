package service

import (
	"fmt"
	"sync"

	"example.com/eventledger/audit"
	"example.com/eventledger/batch"
	"example.com/eventledger/model"
	"example.com/eventledger/query"
	"example.com/eventledger/snapshot"
	"example.com/eventledger/store"
)

type Service struct {
	mu        sync.Mutex
	store     *store.Store
	audit     *audit.Log
	snapshots *snapshot.Manager
}

func New() *Service {
	return &Service{store: store.New(100), audit: audit.New(200), snapshots: snapshot.New(20)}
}

func (service *Service) Save(actor string, value model.EventRecord) (model.EventRecord, error) {
	saved, err := service.store.Put(value)
	if err != nil {
		return model.EventRecord{}, err
	}
	service.audit.Append("save", saved.ID, actor, saved.Name)
	return saved, nil
}

func (service *Service) Remove(actor, id string) error {
	removed, err := service.store.Delete(id)
	if err != nil {
		return err
	}
	service.audit.Append("remove", id, actor, removed.Name)
	return nil
}

func (service *Service) Get(id string) (model.EventRecord, bool) {
	return service.store.Get(id)
}

func (service *Service) List(filter query.Filter, field query.SortField, descending bool) ([]model.EventRecord, error) {
	values := query.Select(service.store.List(), filter)
	return query.Sort(values, field, descending)
}

func (service *Service) Apply(actor string, operations []batch.Operation) error {
	service.mu.Lock()
	defer service.mu.Unlock()
	result, err := batch.Apply(service.store.List(), operations)
	if err != nil {
		return err
	}
	if err := service.store.ReplaceAll(result.Values); err != nil {
		return err
	}
	service.audit.Append("batch", "*", actor, fmt.Sprintf("created=%d updated=%d deleted=%d", result.Created, result.Updated, result.Deleted))
	return nil
}

func (service *Service) Capture(reason string) snapshot.Snapshot {
	snapshot := service.snapshots.Capture(reason, service.store.List())
	service.audit.Append("snapshot", fmt.Sprint(snapshot.ID), "system", reason)
	return snapshot
}

func (service *Service) Restore(actor string, id uint64) error {
	service.mu.Lock()
	defer service.mu.Unlock()
	values, err := service.snapshots.Restore(id)
	if err != nil {
		return err
	}
	if err := service.store.ReplaceAll(values); err != nil {
		return err
	}
	service.audit.Append("restore", fmt.Sprint(id), actor, "")
	return nil
}

func (service *Service) Audit() []audit.Entry { return service.audit.Entries() }

func (service *Service) Revision() uint64 { return service.store.Revision() }
