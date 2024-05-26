package service

type MockAuthGenerator struct {
	MockedValidateToken      func(data []string, token string) (bool, map[string]interface{}, error)
	MockedCreateAccessToken  func(data map[string]string) (string, error)
	MockedCreateRefreshToken func(data map[string]string) (string, error)
}

func NewMockAuthGenerator(
	mockedValidateToken func(data []string, token string) (bool, map[string]interface{}, error),
	mockedCreateAccessToken func(data map[string]string) (string, error),
	mockedCreateRefreshToken func(data map[string]string) (string, error),
) AuthGenerator {
	return &MockAuthGenerator{
		MockedValidateToken:      mockedValidateToken,
		MockedCreateAccessToken:  mockedCreateAccessToken,
		MockedCreateRefreshToken: mockedCreateRefreshToken,
	}
}

func (m MockAuthGenerator) CreateAccessToken(data map[string]string) (string, error) {
	return m.MockedCreateAccessToken(data)
}

func (m MockAuthGenerator) CreateRefreshToken(data map[string]string) (string, error) {
	return m.MockedCreateRefreshToken(data)
}

func (m MockAuthGenerator) ValidateToken(data []string, token string) (bool, map[string]interface{}, error) {
	return m.MockedValidateToken(data, token)
}
