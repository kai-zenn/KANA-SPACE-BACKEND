package bcrypt

import "golang.org/x/crypto/bcrypt"

type Interface interface {
  GenerateHashPassword(password string) (string, error)
  CompareHashPassword(hashPassword string, password string) error
}

type cryptoBcrypt struct {}

func NewCryptoBcrypt() Interface {
	return &cryptoBcrypt{}
}

func (c *cryptoBcrypt) GenerateHashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (c *cryptoBcrypt) CompareHashPassword(hashPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
}
