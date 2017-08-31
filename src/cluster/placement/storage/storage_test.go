// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package storage

import (
	"testing"

	"github.com/m3db/m3cluster/generated/proto/placementpb"
	"github.com/m3db/m3cluster/kv"
	"github.com/m3db/m3cluster/kv/mem"
	"github.com/m3db/m3cluster/placement"

	"github.com/stretchr/testify/require"
)

func TestStorageWithSinglePlacement(t *testing.T) {
	ps := newTestPlacementStorage(placement.NewOptions())
	err := ps.Delete()
	require.Error(t, err)
	require.Equal(t, kv.ErrNotFound, err)

	p := placement.NewPlacement().
		SetInstances([]placement.Instance{}).
		SetShards([]uint32{}).
		SetReplicaFactor(0)

	err = ps.SetIfNotExist(p)
	require.NoError(t, err)

	err = ps.SetIfNotExist(p)
	require.Error(t, err)
	require.Equal(t, kv.ErrAlreadyExists, err)

	pGet, v, err := ps.Placement()
	require.NoError(t, err)
	require.Equal(t, 1, v)
	require.Equal(t, p.SetVersion(1), pGet)

	err = ps.CheckAndSet(p, v)
	require.NoError(t, err)

	err = ps.CheckAndSet(p, v)
	require.Error(t, err)
	require.Equal(t, kv.ErrVersionMismatch, err)

	pGet, v, err = ps.Placement()
	require.NoError(t, err)
	require.Equal(t, 2, v)
	require.Equal(t, p.SetVersion(2), pGet)

	err = ps.Delete()
	require.NoError(t, err)

	_, _, err = ps.Placement()
	require.Error(t, err)
	require.Equal(t, kv.ErrNotFound, err)

	err = ps.SetIfNotExist(p)
	require.NoError(t, err)

	pGet, v, err = ps.Placement()
	require.NoError(t, err)
	require.Equal(t, 1, v)
	require.Equal(t, p.SetVersion(1), pGet)

	proto, v, err := ps.PlacementProto()
	require.NoError(t, err)
	require.Equal(t, 1, v)

	actualP, err := placement.NewPlacementFromProto(proto.(*placementpb.Placement))
	require.NoError(t, err)
	require.Equal(t, p.SetVersion(0), actualP)
}

func TestStorageWithPlacementSnapshots(t *testing.T) {
	ps := newTestPlacementStorage(placement.NewOptions().SetIsStaged(true))

	p := placement.NewPlacement().
		SetInstances([]placement.Instance{}).
		SetShards([]uint32{}).
		SetReplicaFactor(0)

	err := ps.SetIfNotExist(p)
	require.NoError(t, err)

	err = ps.SetIfNotExist(p)
	require.Error(t, err)
	require.Equal(t, kv.ErrAlreadyExists, err)

	pGet1, v, err := ps.Placement()
	require.NoError(t, err)
	require.Equal(t, 1, v)
	require.Equal(t, p.SetVersion(1), pGet1)

	err = ps.CheckAndSet(p, v)
	require.NoError(t, err)

	err = ps.CheckAndSet(p, v)
	require.Error(t, err)
	require.Equal(t, kv.ErrVersionMismatch, err)

	pGet2, v, err := ps.Placement()
	require.NoError(t, err)
	require.Equal(t, 2, v)
	require.Equal(t, p.SetVersion(2), pGet2)

	newProto, v, err := ps.PlacementProto()
	require.NoError(t, err)
	require.Equal(t, 2, v)

	newPs, err := placement.NewPlacementsFromProto(newProto.(*placementpb.PlacementSnapshots))
	require.NoError(t, err)
	require.Equal(t, pGet1.SetVersion(0), newPs[0])
	require.Equal(t, pGet2.SetVersion(0), newPs[1])

	err = ps.Delete()
	require.NoError(t, err)

	_, _, err = ps.Placement()
	require.Error(t, err)
	require.Equal(t, kv.ErrNotFound, err)

	err = ps.SetIfNotExist(p)
	require.NoError(t, err)

	pGet3, v, err := ps.Placement()
	require.NoError(t, err)
	require.Equal(t, 1, v)
	require.Equal(t, p.SetVersion(1), pGet3)
}

func newTestPlacementStorage(pOpts placement.Options) placement.Storage {
	return NewPlacementStorage(mem.NewStore(), "key", pOpts)
}