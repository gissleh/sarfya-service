package jsonemphasisdata

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya-service/emphasis"
	"os"
)

type Storage map[string]*emphasis.FitResult

func (s Storage) FindEmphasis(ctx context.Context, id string) (*emphasis.FitResult, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	res, ok := s[id]
	if !ok {
		return nil, sarfya.ErrExampleNotFound
	}

	return res, nil
}

func (s Storage) SaveEmphasisInput(_ context.Context, _ emphasis.Input) error {
	return errors.New("emphasis storage is read-only")
}

func Load(fileName string) (Storage, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	res := make(Storage)
	err = json.NewDecoder(file).Decode(&res)
	_ = file.Close()
	if err != nil {
		return nil, err
	}

	return res, nil
}
