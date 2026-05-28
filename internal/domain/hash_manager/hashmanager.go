package hashmanager

import "github.com/xianghuzhao/kdfcrypt"

type HashValidator interface {
	CalculateHash(password string) string
	Validate(password string, encoded string) bool
}

type HashValidatorImpl struct {
}

func NewHashValidator() HashValidator {
	return &HashValidatorImpl{}
}

func (h *HashValidatorImpl) CalculateHash(password string) string {
	encoded, _ := kdfcrypt.Encode(password, &kdfcrypt.Option{
		Algorithm:        "argon2id",
		Param:            "m=65536,t=1,p=4",
		RandomSaltLength: 16,
		HashLength:       32,
	})
	return encoded
}

func (h *HashValidatorImpl) Validate(password string, encoded string) bool {
	ok, err := kdfcrypt.Verify(password, encoded)
	if err != nil || !ok {
		return false
	}
	return true
}
