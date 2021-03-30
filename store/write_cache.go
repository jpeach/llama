// Copyright 2020 Nelson Elhage
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package store

import (
	"context"

	"github.com/nelhage/llama/store/internal/storeutil"
)

type WriteCachingStore struct {
	inner Store
	seen  storeutil.Cache
}

func WriteCaching(inner Store) Store {
	return &WriteCachingStore{
		inner: inner,
	}
}

type storeHasheder interface {
	StoreHashed(ctx context.Context, obj []byte, hash string) (string, error)
}

func (w *WriteCachingStore) Store(ctx context.Context, obj []byte) (string, error) {
	hash := storeutil.HashObject(obj)
	u := w.seen.StartUpload(hash)
	defer u.Rollback()
	var got string
	var err error
	if sh, ok := w.inner.(storeHasheder); ok {
		got, err = sh.StoreHashed(ctx, obj, hash)
	} else {
		got, err = w.inner.Store(ctx, obj)
	}
	if err == nil {
		u.Complete()
	}
	return got, err
}

func (w *WriteCachingStore) GetObjects(ctx context.Context, gets []GetRequest) {
	w.inner.GetObjects(ctx, gets)
	for i := range gets {
		if gets[i].Err == nil {
			u := w.seen.StartUpload(gets[i].Id)
			u.Complete()
		}
	}
}
