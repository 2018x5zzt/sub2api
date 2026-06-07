package service

import "context"

type compactKeepaliveRepoStub struct {
	values map[string]string
}

func (s *compactKeepaliveRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *compactKeepaliveRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if s.values == nil {
		return "", ErrSettingNotFound
	}
	if v, ok := s.values[key]; ok {
		return v, nil
	}
	return "", ErrSettingNotFound
}

func (s *compactKeepaliveRepoStub) Set(ctx context.Context, key, value string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	s.values[key] = value
	return nil
}

func (s *compactKeepaliveRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *compactKeepaliveRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *compactKeepaliveRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *compactKeepaliveRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}
